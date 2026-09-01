package httpx

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
type Hub struct {
	mu      sync.RWMutex
	clients map[*hubClient]struct{}
	repo    experiment.Repository
}

type hubClient struct {
	events      chan Event
	tradeEvents chan Event
	mu          sync.RWMutex
	candles     map[string]struct{}
	trades      map[string]struct{}
}

type clientCommand struct {
	Type    string `json:"type"`
	Payload struct {
		Symbol    string `json:"symbol"`
		Timeframe string `json:"timeframe"`
	} `json:"payload"`
}

func NewHub(repo experiment.Repository) *Hub {
	return &Hub{clients: make(map[*hubClient]struct{}), repo: repo}
}
func (h *Hub) subscribe() (*hubClient, func()) {
	client := &hubClient{events: make(chan Event, 64), tradeEvents: make(chan Event, 128), candles: make(map[string]struct{}), trades: make(map[string]struct{})}
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	return client, func() {
		h.mu.Lock()
		if _, ok := h.clients[client]; ok {
			delete(h.clients, client)
			close(client.events)
			close(client.tradeEvents)
		}
		h.mu.Unlock()
	}
}
func (h *Hub) Broadcast(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if !client.accepts(event) {
			continue
		}
		if event.Type == "TRADE_TICK" {
			select {
			case client.tradeEvents <- event:
			default:
			}
		} else {
			select {
			case client.events <- event:
			default:
			}
		}
	}
}

func subscriptionKey(symbol, timeframe string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + ":" + strings.TrimSpace(timeframe)
}

func (c *hubClient) apply(command clientCommand) {
	symbol := strings.ToUpper(strings.TrimSpace(command.Payload.Symbol))
	timeframe := strings.TrimSpace(command.Payload.Timeframe)
	if !market.IsSupportedSymbol(symbol) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	switch command.Type {
	case "SUBSCRIBE_CANDLES":
		if market.IsSupportedTimeframe(timeframe) {
			c.candles[subscriptionKey(symbol, timeframe)] = struct{}{}
		}
	case "UNSUBSCRIBE_CANDLES":
		delete(c.candles, subscriptionKey(symbol, timeframe))
	case "SUBSCRIBE_TRADES":
		c.trades[symbol] = struct{}{}
	case "UNSUBSCRIBE_TRADES":
		delete(c.trades, symbol)
	}
}

func (c *hubClient) accepts(event Event) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	switch event.Type {
	case "CANDLE_UPDATE":
		candle, ok := event.Payload.(market.Candle)
		if !ok {
			return false
		}
		_, ok = c.candles[subscriptionKey(candle.Symbol, candle.Timeframe)]
		return ok
	case "TRADE_TICK":
		trade, ok := event.Payload.(market.TradeTick)
		if !ok {
			return false
		}
		_, ok = c.trades[strings.ToUpper(trade.Symbol)]
		return ok
	default:
		return true
	}
}
func (h *Hub) JobUpdated(result experiment.Result) {
	if result.Status == "RUNNING" {
		h.Broadcast(Event{Type: "SEARCH_PROGRESS", Payload: map[string]int{"tested": 0, "total": 1}})
		return
	}
	if result.Status != "COMPLETED" && result.Status != "FAILED" {
		return
	}
	h.Broadcast(Event{Type: "SEARCH_PROGRESS", Payload: map[string]int{"tested": 1, "total": 1}})
	if result.Status == "COMPLETED" {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		results, err := h.repo.List(ctx)
		if err == nil {
			ranked := experiment.Rank(results)
			if len(ranked) > 10 {
				ranked = ranked[:10]
			}
			h.Broadcast(Event{Type: "LEADERBOARD_UPDATE", Payload: ranked})
		}
	}
}
func serveWebSocket(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
		})
		if err != nil {
			return
		}
		defer conn.CloseNow()
		client, unsubscribe := hub.subscribe()
		defer unsubscribe()
		ctx := r.Context()
		readerDone := make(chan struct{})
		go func() {
			defer close(readerDone)
			for {
				var command clientCommand
				if err := wsjson.Read(ctx, conn, &command); err != nil {
					return
				}
				client.apply(command)
			}
		}()
		for {
			select {
			case event, ok := <-client.events:
				if !ok {
					return
				}
				if err := wsjson.Write(ctx, conn, event); err != nil {
					return
				}
				continue
			default:
			}
			select {
			case <-ctx.Done():
				_ = conn.Close(websocket.StatusNormalClosure, "closed")
				return
			case <-readerDone:
				return
			case event, ok := <-client.events:
				if !ok {
					return
				}
				if err := wsjson.Write(ctx, conn, event); err != nil {
					return
				}
			case event, ok := <-client.tradeEvents:
				if !ok {
					return
				}
				if err := wsjson.Write(ctx, conn, event); err != nil {
					return
				}
			}
		}
	}
}
