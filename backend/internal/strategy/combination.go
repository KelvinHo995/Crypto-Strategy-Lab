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

// MinLookback is the max across every wrapped strategy — the whole
// combination shares one sliding window (see Analyze), so it has to be
// large enough for the neediest strategy in the mix.
func (cs *CombinedStrategy) MinLookback() int {
	max := 1
	for _, s := range cs.Strategies {
		if la, ok := s.(LookbackAware); ok {
			if n := la.MinLookback(); n > max {
				max = n
			}
		}
	}
	return max
}

func (cs *CombinedStrategy) SentimentModels() []SentimentModelIdentity {
	var models []SentimentModelIdentity
	for _, child := range cs.Strategies {
		if provider, ok := child.(SentimentModelProvider); ok {
			models = append(models, provider.SentimentModels()...)
		}
	}
	return normalizeSentimentModels(models)
}

func resolveCombinationPolicy(policyName string, weights []float64) CombinationPolicy {
	if policyName == "weighted" {
		return WeightedPolicy{Weights: normalizeWeights(weights)}
	}
	return MajorityPolicy{}
}

// normalizeWeights scales the supplied weights to sum to 1. If none were
// supplied (all zero — e.g. a caller that only cares about majority voting
// left them unset), falls back to equal weights rather than producing a
// policy that can never fire.
func normalizeWeights(weights []float64) []float64 {
	if len(weights) == 0 {
		return nil
	}
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	normalized := make([]float64, len(weights))
	if sum <= 0 {
		equal := 1.0 / float64(len(weights))
		for i := range normalized {
			normalized[i] = equal
		}
		return normalized
	}
	for i, w := range weights {
		normalized[i] = w / sum
	}
	return normalized
}
