package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestSearchLoop_StartsImmediatelyAndProducesCandidates(t *testing.T) {
	repo := newFakeRepo()
	candles := &fakeCandleRepo{}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), repo, httpx.Dependencies{Candles: candles})
	defer router.Close()
	srv := httptest.NewServer(router)
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"maxCandidates":3}`
	start := time.Now()
	resp, err := http.Post(srv.URL+"/search/loop", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", resp.StatusCode)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("took %s to respond, want near-instant (proves it doesn't wait for candidates to finish)", elapsed)
	}
	var out httpx.StartLoopResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.SearchID == "" || out.Status != "STARTED" || out.MaxCandidates != 3 {
		t.Fatalf("got %+v", out)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		results, err := repo.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		matching := 0
		completed := 0
		for _, r := range results {
			if r.SearchID != out.SearchID {
				continue
			}
			matching++
			if r.Status == "COMPLETED" || r.Status == "FAILED" {
				completed++
			}
			if r.SearchTotal != 3 {
				t.Fatalf("result %s has SearchTotal=%d, want 3", r.ID, r.SearchTotal)
			}
		}
		if matching == 3 && completed == 3 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out: %d/3 candidates created, %d/3 finished", matching, completed)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestSearchLoop_ValidatesMaxCandidatesRange(t *testing.T) {
	for _, body := range []string{
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"maxCandidates":1}`,
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"maxCandidates":201}`,
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"maxCandidates":5,"maxDurationSeconds":10}`,
	} {
		srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), newFakeRepo()))
		resp, err := http.Post(srv.URL+"/search/loop", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		srv.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 for %s", resp.StatusCode, body)
		}
	}
}

func TestSearchLoop_RealProgressOverWebSocket(t *testing.T) {
	candles := &fakeCandleRepo{}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Candles: candles})
	defer router.Close()
	srv := httptest.NewServer(router)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"maxCandidates":3}`
	resp, err := http.Post(srv.URL+"/search/loop", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("start=%d", resp.StatusCode)
	}

	maxTested := 0
	for maxTested < 3 {
		var envelope struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := wsjson.Read(ctx, conn, &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Type != "SEARCH_PROGRESS" {
			continue // e.g. LEADERBOARD_UPDATE, which has an array payload
		}
		var progress struct {
			Tested int `json:"tested"`
			Total  int `json:"total"`
		}
		if err := json.Unmarshal(envelope.Payload, &progress); err != nil {
			t.Fatal(err)
		}
		if progress.Total != 3 {
			t.Fatalf("total = %d, want 3 (real SearchTotal, not a placeholder)", progress.Total)
		}
		if progress.Tested < maxTested {
			t.Fatalf("tested went backwards: %d after %d", progress.Tested, maxTested)
		}
		maxTested = progress.Tested
	}
}
