package strategy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type mockSentimentClient struct {
	score float64
	err   error
}

func (m *mockSentimentClient) FetchSentiment(ctx context.Context, timestamp int64) (float64, error) {
	return m.score, m.err
}

func TestSentimentStrategy_Analyze(t *testing.T) {
	candles := []market.Candle{{OpenTime: 1234567890}}

	tests := []struct {
		name       string
		client     strategy.SentimentClient
		baseSignal strategy.Signal
		want       strategy.Signal
	}{
		{
			name:       "buy on high sentiment",
			client:     &mockSentimentClient{score: 0.9, err: nil},
			baseSignal: strategy.Hold,
			want:       strategy.Buy,
		},
		{
			name:       "sell on low sentiment",
			client:     &mockSentimentClient{score: 0.1, err: nil},
			baseSignal: strategy.Hold,
			want:       strategy.Sell,
		},
		{
			name:       "fallback to base on neutral sentiment",
			client:     &mockSentimentClient{score: 0.5, err: nil},
			baseSignal: strategy.Buy, // Base says Buy
			want:       strategy.Buy,
		},
		{
			name:       "fallback to base on error",
			client:     &mockSentimentClient{score: 0, err: errors.New("api down")},
			baseSignal: strategy.Sell, // Base says Sell
			want:       strategy.Sell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &DummyStrategy{AlwaysSignal: tt.baseSignal}
			s := strategy.NewSentimentStrategy(base, tt.client, 0.8)

			if got := s.Analyze(candles); got != tt.want {
				t.Errorf("SentimentStrategy.Analyze() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSentimentFactorySupportsStandaloneAndDecoratedStrategies(t *testing.T) {
	client := &mockSentimentClient{score: 0.5}
	standalone := strategy.NewSentimentFactory(nil, client, 0.8)(nil)
	if standalone.Name() != "Sentiment" {
		t.Fatalf("standalone name = %q", standalone.Name())
	}

	baseFactory := func(map[string]any) strategy.Strategy {
		return &DummyStrategy{AlwaysSignal: strategy.Buy}
	}
	decorated := strategy.NewSentimentFactory(baseFactory, client, 0.8)(nil)
	if decorated.Name() != "Sentiment+Dummy" {
		t.Fatalf("decorated name = %q", decorated.Name())
	}
}
