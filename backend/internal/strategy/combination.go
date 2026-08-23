package strategy

import (
	"strings"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

type CombinationPolicy interface {
	Combine(signals []Signal) Signal
}

type MajorityPolicy struct{}

func (p MajorityPolicy) Combine(signals []Signal) Signal {
	buyCount := 0
	sellCount := 0
	for _, sig := range signals {
		if sig == Buy {
			buyCount++
		} else if sig == Sell {
			sellCount++
		}
	}
	if buyCount > sellCount {
		return Buy
	} else if sellCount > buyCount {
		return Sell
	}
	return Hold
}

type WeightedPolicy struct {
	Weights []float64
}

func (p WeightedPolicy) Combine(signals []Signal) Signal {
	if len(p.Weights) != len(signals) {
		return Hold
	}
	score := 0.0
	for i, sig := range signals {
		if sig == Buy {
			score += p.Weights[i]
		} else if sig == Sell {
			score -= p.Weights[i]
		}
	}
	if score > 0 {
		return Buy
	} else if score < 0 {
		return Sell
	}
	return Hold
}

// CombinedStrategy runs multiple strategies and combines their signals using a CombinationPolicy.
var _ Strategy = (*CombinedStrategy)(nil)

type CombinedStrategy struct {
	Strategies []Strategy
	Policy     CombinationPolicy
	name       string
}

func NewCombinedStrategy(strategies []Strategy, policy CombinationPolicy, name string) *CombinedStrategy {
	return &CombinedStrategy{
		Strategies: strategies,
		Policy:     policy,
		name:       name,
	}
}

func (cs *CombinedStrategy) Name() string {
	if cs.name != "" {
		return cs.name
	}
	names := make([]string, len(cs.Strategies))
	for i, s := range cs.Strategies {
		names[i] = s.Name()
	}
	return strings.Join(names, "+")
}

func (cs *CombinedStrategy) Analyze(candles []market.Candle) Signal {
	signals := make([]Signal, len(cs.Strategies))
	for i, s := range cs.Strategies {
		signals[i] = s.Analyze(candles)
	}
	return cs.Policy.Combine(signals)
}

func resolveCombinationPolicy(policyName string, count int) CombinationPolicy {
	if policyName == "weighted" {
		weights := make([]float64, count)
		// simple default weights
		for i := 0; i < count; i++ {
			weights[i] = 1.0 / float64(count)
		}
		// giving slightly more weight to the first one just to differentiate
		if count > 0 {
			weights[0] += 0.1
		}
		return WeightedPolicy{Weights: weights}
	}
	// default to majority
	return MajorityPolicy{}
}
