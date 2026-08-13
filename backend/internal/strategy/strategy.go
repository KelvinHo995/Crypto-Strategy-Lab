package strategy

import "github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"

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

type Registry struct {
	strategies map[string]Strategy
}

func NewRegistry() *Registry {
	return &Registry{strategies: make(map[string]Strategy)}
}

func (r *Registry) Register(s Strategy) {
	r.strategies[s.Name()] = s
}
