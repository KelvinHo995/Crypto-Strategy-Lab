package strategy

import (
	"sort"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

type Signal string

const (
	Buy  Signal = "BUY"
	Sell Signal = "SELL"
	Hold Signal = "HOLD"
)

type Strategy interface {
	Name() string
	Analyze(candles []market.Candle) Signal
}

// Registry holds registered strategy implementations.
// It is NOT safe for concurrent use. All Register calls must
// happen during application startup before any concurrent Get/List.
type Registry struct {
	strategies map[string]Strategy
}

func NewRegistry() *Registry {
	return &Registry{strategies: make(map[string]Strategy)}
}

func (r *Registry) Register(s Strategy) {
	r.strategies[s.Name()] = s
}

func (r *Registry) Get(name string) (Strategy, bool) {
	s, ok := r.strategies[name]
	return s, ok
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
