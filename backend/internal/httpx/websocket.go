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

// SearchProgressPayload is the SEARCH_PROGRESS wire payload. SearchID/Status/
// Reason are additive (omitted on ordinary per-candidate progress ticks) —
// only SearchLoopStopped sets them, to signal a search-run ending early
// (STOPPED/FAILED) instead of leaving the frontend waiting for tested==total
// forever when fewer than the requested candidates were ever enqueued.
type SearchProgressPayload struct {
	Tested   int    `json:"tested"`
	Total    int    `json:"total,omitempty"`
	SearchID string `json:"searchId,omitempty"`
	Status   string `json:"status,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

func (h *Hub) JobUpdated(result experiment.Result) {
	if result.Status != "COMPLETED" && result.Status != "FAILED" {
		return // no partial-progress broadcast on RUNNING — tested/total only means something once a candidate finishes
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	bySearch, err := h.repo.ListBySearch(ctx, result.SearchID)
	if err != nil {
		return
	}
	tested := 0
	for _, r := range bySearch {
		if r.Status == "COMPLETED" || r.Status == "FAILED" {
			tested++
		}
	}
	h.Broadcast(Event{Type: "SEARCH_PROGRESS", Payload: SearchProgressPayload{
		Tested: tested, Total: result.SearchTotal, SearchID: result.SearchID,
	}})

	if result.Status == "COMPLETED" {
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

// SearchLoopStopped broadcasts the terminal search-run signal for a loop
// that stopped before generating its full requested candidate count — see
// experiment.RunSearchLoop's onStopped parameter.
func (h *Hub) SearchLoopStopped(searchID, status, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tested := 0
	if bySearch, err := h.repo.ListBySearch(ctx, searchID); err == nil {
		for _, r := range bySearch {
			if r.Status == "COMPLETED" || r.Status == "FAILED" {
				tested++
			}
		}
	}
	h.Broadcast(Event{Type: "SEARCH_PROGRESS", Payload: SearchProgressPayload{
		Tested: tested, SearchID: searchID, Status: status, Reason: reason,
	}})
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
