package httpx

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
type Hub struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
	repo    experiment.Repository
}

func NewHub(repo experiment.Repository) *Hub {
	return &Hub{clients: make(map[chan Event]struct{}), repo: repo}
}
func (h *Hub) subscribe() (chan Event, func()) {
	ch := make(chan Event, 32)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if _, ok := h.clients[ch]; ok {
			delete(h.clients, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
}
func (h *Hub) Broadcast(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- event:
		default:
		}
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
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		events, unsubscribe := hub.subscribe()
		defer unsubscribe()
		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				_ = conn.Close(websocket.StatusNormalClosure, "closed")
				return
			case event, ok := <-events:
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
