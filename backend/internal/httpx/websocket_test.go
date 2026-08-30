package httpx_test

import (
	"bytes"
	"context"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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
