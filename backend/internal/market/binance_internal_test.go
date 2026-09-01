package market

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestCombinedEventParsesKlineAndTrade(t *testing.T) {
	var kline binanceKlineEvent
	kline.K.Symbol = "ETHUSDT"
	kline.K.Interval = "1h"
	kline.K.Start = 1_000
	kline.K.Open, kline.K.High, kline.K.Low, kline.K.Close, kline.K.Volume = "100", "110", "90", "105", "12.5"
	klineJSON, _ := json.Marshal(kline)
	event, err := (binanceCombinedEvent{Stream: "ethusdt@kline_1h", Data: klineJSON}).liveEvent()
	if err != nil || event.Type != "CANDLE_UPDATE" || event.Candle.Symbol != "ETHUSDT" {
		t.Fatalf("kline event=%+v err=%v", event, err)
	}

	tradeJSON, _ := json.Marshal(binanceAggTradeEvent{Symbol: "SOLUSDT", TradeID: 42, Price: "155.25", Quantity: "2.5", TradeTime: 2_000, BuyerMade: true})
	event, err = (binanceCombinedEvent{Stream: "solusdt@aggTrade", Data: tradeJSON}).liveEvent()
	if err != nil || event.Type != "TRADE_TICK" || event.Trade.Side != "SELL" || event.Trade.Price != 155.25 {
		t.Fatalf("trade event=%+v err=%v", event, err)
	}
}

func TestCombinedKlineDistinguishesLowPriceFromLastTradeID(t *testing.T) {
	payload := json.RawMessage(`{"stream":"btcusdt@kline_5m","data":{"s":"BTCUSDT","k":{"t":1788246600000,"L":6642835204,"s":"BTCUSDT","i":"5m","o":"78775.25","c":"78726.51","h":"78775.26","l":"78718.60","v":"14.41","x":false}}}`)
	var wrapped binanceCombinedEvent
	if err := json.Unmarshal(payload, &wrapped); err != nil {
		t.Fatal(err)
	}
	event, err := wrapped.liveEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Candle.Low != 78718.60 || event.Candle.Close != 78726.51 {
		t.Fatalf("candle=%+v", event.Candle)
	}
}

func TestCombinedEndpointUsesSingleStreamConnection(t *testing.T) {
	got := combinedStreamEndpoint("wss://stream.binance.com:9443/ws", []string{"btcusdt@kline_5m", "ethusdt@aggTrade"})
	want := "wss://stream.binance.com:9443/stream?streams=btcusdt@kline_5m/ethusdt@aggTrade"
	if got != want {
		t.Fatalf("endpoint=%q want=%q", got, want)
	}
}

func TestLiveKlineRejectsInvalidNumbers(t *testing.T) {
	var event binanceKlineEvent
	event.K.Symbol, event.K.Interval, event.K.Start = "BTCUSDT", "5m", 1
	event.K.Open, event.K.High, event.K.Low = "not-a-number", "2", "1"
	event.K.Close, event.K.Volume = "1.5", "10"
	if _, err := event.candle(); err == nil {
		t.Fatal("invalid numeric payload accepted")
	}
}

func TestInvalidLiveRequestClosesStream(t *testing.T) {
	stream := NewBinance(nil).StreamLiveCandles(context.Background(), "", "1m")
	if _, ok := <-stream; ok {
		t.Fatal("invalid live request produced a candle")
	}
}

func TestLiveStreamReconnectsAfterDisconnect(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept websocket: %v", err)
			return
		}
		if attempt == 1 {
			_ = conn.Close(websocket.StatusInternalError, "forced disconnect")
			return
		}

		var event binanceKlineEvent
		event.K.Symbol = "BTCUSDT"
		event.K.Interval = "5m"
		event.K.Start = 1_000
		event.K.Open = "100"
		event.K.High = "110"
		event.K.Low = "90"
		event.K.Close = "105"
		event.K.Volume = "12.5"
		event.K.Closed = true
		if err := wsjson.Write(r.Context(), conn, event); err != nil {
			t.Errorf("write reconnected candle: %v", err)
		}
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	binance := NewBinance(nil)
	binance.WSBaseURL = "ws" + strings.TrimPrefix(server.URL, "http")
	stream := binance.StreamLiveCandles(ctx, "BTCUSDT", "5m")

	select {
	case candle, ok := <-stream:
		if !ok {
			t.Fatal("stream closed after disconnect instead of reconnecting")
		}
		if candle.OpenTime != 1_000 || candle.Close != 105 || !candle.IsClosed {
			t.Fatalf("reconnected candle = %+v", candle)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("timed out waiting for candle after reconnect")
	}
	cancel()
	if attempts.Load() < 2 {
		t.Fatalf("connection attempts = %d, want at least 2", attempts.Load())
	}
	select {
	case _, ok := <-stream:
		if ok {
			t.Fatal("stream produced another candle after cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("stream did not close after cancellation")
	}
}

func TestLiveReconnectBackoffIsBounded(t *testing.T) {
	backoff := liveReconnectInitialBackoff
	want := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 30 * time.Second, 30 * time.Second}
	for i, expected := range want {
		backoff = nextLiveReconnectBackoff(backoff)
		if backoff != expected {
			t.Fatalf("step %d backoff = %s, want %s", i+1, backoff, expected)
		}
	}
}
