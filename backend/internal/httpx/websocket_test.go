package httpx_test

import (
	"bytes"
	"context"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeCombinedLive struct {
	events chan market.LiveEvent
}

func (f *fakeCombinedLive) StreamLiveCandles(context.Context, string, string) <-chan market.Candle {
	closed := make(chan market.Candle)
	close(closed)
	return closed
}

func (f *fakeCombinedLive) StreamMarketEvents(context.Context, []string, []string) <-chan market.LiveEvent {
	return f.events
}

func TestWebSocketReceivesSearchProgress(t *testing.T) {
	router := httpx.NewRouter(newTestRegistry(), newFakeRepo())
	defer router.Close()
	srv := httptest.NewServer(router)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"strategies":["MA"]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("start=%d", resp.StatusCode)
	}
	for {
		var event struct {
			Type string `json:"type"`
		}
		if err := wsjson.Read(ctx, conn, &event); err != nil {
			t.Fatal(err)
		}
		if event.Type == "SEARCH_PROGRESS" {
			return
		}
	}
}

func TestWebSocketFiltersSubscribedMarketEvents(t *testing.T) {
	live := &fakeCombinedLive{events: make(chan market.LiveEvent, 4)}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Live: live})
	defer router.Close()
	server := httptest.NewServer(router)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	command := map[string]any{"type": "SUBSCRIBE_CANDLES", "payload": map[string]string{"symbol": "ETHUSDT", "timeframe": "1h"}}
	if err := wsjson.Write(ctx, conn, command); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	live.events <- market.LiveEvent{Type: "CANDLE_UPDATE", Candle: market.Candle{Symbol: "BTCUSDT", Timeframe: "1h", OpenTime: 1}}
	live.events <- market.LiveEvent{Type: "CANDLE_UPDATE", Candle: market.Candle{Symbol: "ETHUSDT", Timeframe: "1h", OpenTime: 2}}

	var event struct {
		Type    string        `json:"type"`
		Payload market.Candle `json:"payload"`
	}
	if err := wsjson.Read(ctx, conn, &event); err != nil {
		t.Fatal(err)
	}
	if event.Type != "CANDLE_UPDATE" || event.Payload.Symbol != "ETHUSDT" || event.Payload.OpenTime != 2 {
		t.Fatalf("event=%+v", event)
	}
}
