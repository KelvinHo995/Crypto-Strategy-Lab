package strategy

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

var _ Strategy = (*SentimentStrategy)(nil)

const SentimentStrategyVersion = "v1"

type SentimentModelIdentity struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SentimentModelProvider interface {
	SentimentModels() []SentimentModelIdentity
}

// SentimentClient provides a directional score in the range 0..1.
// Values above 0.5 are positive, values below 0.5 are negative, and 0.5 is neutral.
type SentimentClient interface {
	FetchSentiment(ctx context.Context, timestamp int64) (float64, error)
}

// ModelAwareSentimentClient is implemented by production lookups that can
// identify the persisted model observation behind a score.
type ModelAwareSentimentClient interface {
	FetchSentimentWithModel(ctx context.Context, timestamp int64) (score float64, modelName, modelVersion string, err error)
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

	modelsMu sync.Mutex
	models   map[SentimentModelIdentity]struct{}
}

func NewSentimentStrategy(base Strategy, client SentimentClient, threshold float64) *SentimentStrategy {
	if client == nil {
		client = &DummySentimentClient{}
	}
	if threshold < 0.5 || threshold > 1 {
		threshold = 0.7
	}
	return &SentimentStrategy{
		BaseStrategy: base,
		Client:       client,
		Threshold:    threshold,
	}
}

// NewSentimentFactory creates standalone sentiment strategies when baseFactory
// is nil, or reusable sentiment decorators around any supplied strategy factory.
func NewSentimentFactory(baseFactory StrategyFactory, client SentimentClient, threshold float64) StrategyFactory {
	return func(params map[string]any) Strategy {
		var base Strategy
		if baseFactory != nil {
			base = baseFactory(params)
		}
		return NewSentimentStrategy(base, client, threshold)
	}
}

func (s *SentimentStrategy) Name() string {
	if s.BaseStrategy != nil {
		return "Sentiment+" + s.BaseStrategy.Name()
	}
	return "Sentiment"
}

// MinLookback defers to BaseStrategy's own requirement, since Analyze falls
// back to it on a neutral/failed sentiment lookup — the same candles slice
// has to satisfy the base strategy too. Sentiment itself only needs 1.
func (s *SentimentStrategy) MinLookback() int {
	if la, ok := s.BaseStrategy.(LookbackAware); ok {
		return la.MinLookback()
	}
	return 1
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

	score, err := s.fetchSentiment(ctx, lastCandle.OpenTime)
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

func (s *SentimentStrategy) fetchSentiment(ctx context.Context, timestamp int64) (float64, error) {
	client, modelAware := s.Client.(ModelAwareSentimentClient)
	if !modelAware {
		return s.Client.FetchSentiment(ctx, timestamp)
	}
	score, modelName, modelVersion, err := client.FetchSentimentWithModel(ctx, timestamp)
	if err != nil {
		return 0, err
	}
	identity := SentimentModelIdentity{Name: strings.TrimSpace(modelName), Version: strings.TrimSpace(modelVersion)}
	if identity.Name == "" || identity.Version == "" {
		return 0, errors.New("sentiment model identity is required")
	}
	s.modelsMu.Lock()
	if s.models == nil {
		s.models = make(map[SentimentModelIdentity]struct{})
	}
	s.models[identity] = struct{}{}
	s.modelsMu.Unlock()
	return score, nil
}

func (s *SentimentStrategy) SentimentModels() []SentimentModelIdentity {
	s.modelsMu.Lock()
	models := make([]SentimentModelIdentity, 0, len(s.models))
	for identity := range s.models {
		models = append(models, identity)
	}
	s.modelsMu.Unlock()
	if provider, ok := s.BaseStrategy.(SentimentModelProvider); ok {
		models = append(models, provider.SentimentModels()...)
	}
	return normalizeSentimentModels(models)
}

func normalizeSentimentModels(models []SentimentModelIdentity) []SentimentModelIdentity {
	unique := make(map[SentimentModelIdentity]struct{}, len(models))
	for _, identity := range models {
		if identity.Name != "" && identity.Version != "" {
			unique[identity] = struct{}{}
		}
	}
	result := make([]SentimentModelIdentity, 0, len(unique))
	for identity := range unique {
		result = append(result, identity)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].Version < result[j].Version
		}
		return result[i].Name < result[j].Name
	})
	return result
}
