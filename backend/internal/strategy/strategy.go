package strategy

import (
	"fmt"
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

// LookbackAware lets a strategy report the fewest trailing candles it needs
// before it can produce anything but Hold — the backtester uses this to size
// its sliding window per-candidate instead of guessing one shared constant.
type LookbackAware interface {
	MinLookback() int
}

type StrategyFactory func(params map[string]any) Strategy

// Registry holds registered strategy implementations and factories.
// It is NOT safe for concurrent use. All Register calls must
// happen during application startup before any concurrent Get/List.
type Registry struct {
	strategies map[string]Strategy
	factories  map[string]StrategyFactory
}

func NewRegistry() *Registry {
	return &Registry{
		strategies: make(map[string]Strategy),
		factories:  make(map[string]StrategyFactory),
	}
}

func (r *Registry) Register(s Strategy) {
	r.strategies[s.Name()] = s
}

func (r *Registry) RegisterFactory(name string, f StrategyFactory) {
	r.factories[name] = f
}

func (r *Registry) Get(name string) (Strategy, bool) {
	s, ok := r.strategies[name]
	return s, ok
}

func (r *Registry) BuildWithParams(name string, params map[string]any) (Strategy, error) {
	if f, ok := r.factories[name]; ok {
		return f(params), nil
	}
	// Fallback: try getting default from registry
	if s, ok := r.strategies[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("strategy %s not found in registry", name)
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
