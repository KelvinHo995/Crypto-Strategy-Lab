package strategy

type CandidateStrategy struct {
	ID         string
	Strategies []string
	Params     map[string]any
	Policy     string // "majority" | "weighted"
}

type StrategyGenerator interface {
	Generate() CandidateStrategy
}

// BuildFromCandidate resolves a candidate into a CombinedStrategy by
// looking up each strategy name via the registry's factory system,
// then wrapping them with the candidate's combination policy.
func BuildFromCandidate(reg *Registry, c CandidateStrategy) (Strategy, error) {
	var strategies []Strategy

	for _, name := range c.Strategies {
		s, err := reg.BuildWithParams(name, c.Params)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, s)
	}

	policy := resolveCombinationPolicy(c.Policy, len(strategies))
	return NewCombinedStrategy(strategies, policy, ""), nil
}

func toInt(v any, fallback int) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	default:
		return fallback, false
	}
}

func toFloat(v any, fallback float64) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	default:
		return fallback, false
	}
}
