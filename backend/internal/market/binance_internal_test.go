package market

import (
	"context"
	"testing"
)

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
