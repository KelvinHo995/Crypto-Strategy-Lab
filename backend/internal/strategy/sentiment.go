package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

var _ Strategy = (*SentimentStrategy)(nil)

// SentimentClient represents a way to fetch sentiment.
// In reality, Person 1 will implement this client in the market/sentiment package
// and we just inject it here.
type SentimentClient interface {
	FetchSentiment(ctx context.Context, timestamp int64) (float64, error)
}

// DummyClient for testing when Person 1 hasn't finished the real client
type DummySentimentClient struct{}

func (d *DummySentimentClient) FetchSentiment(ctx context.Context, timestamp int64) (float64, error) {
	// Returns a neutral score
	return 0.5, nil
}

type SentimentStrategy struct {
	BaseStrategy Strategy
	Client       SentimentClient
	Threshold    float64
}

func NewSentimentStrategy(base Strategy, client SentimentClient, threshold float64) *SentimentStrategy {
	if client == nil {
		client = &DummySentimentClient{}
	}
	return &SentimentStrategy{
		BaseStrategy: base,
		Client:       client,
		Threshold:    threshold,
	}
}

func (s *SentimentStrategy) Name() string {
	return fmt.Sprintf("Sentiment+EmptyBase")
}

func (s *SentimentStrategy) Analyze(candles []market.Candle) Signal {
	if len(candles) == 0 {
		return Hold
	}

	// For MVP, we use the timestamp of the last candle to fetch sentiment
	lastCandle := candles[len(candles)-1]

	// Normally we'd pass a real context, using Background here for simplicity in Analyze signature
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	score, err := s.Client.FetchSentiment(ctx, lastCandle.OpenTime)
	if err != nil {
		// Degradation: if sentiment service is down, fall back to base strategy
		if s.BaseStrategy != nil {
			return s.BaseStrategy.Analyze(candles)
		}
		return Hold
	}

	// Use sentiment to decide
	if score > s.Threshold {
		return Buy
	} else if score < (1.0 - s.Threshold) { // assuming score is 0.0 to 1.0
		return Sell
	}

	// If neutral, fallback to base strategy if present
	if s.BaseStrategy != nil {
		return s.BaseStrategy.Analyze(candles)
	}

	return Hold
}
