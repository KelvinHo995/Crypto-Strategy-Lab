package strategy

// StrategyInstance is one independently configured strategy within a
// composite — its own type, its own params, its own weight. A candidate can
// hold more than one instance of the same type (e.g. MA(20) and MA(50)
// combined), which a flat "one shared params bag" model can't express.
type StrategyInstance struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params,omitempty"`
	Weight float64        `json:"weight,omitempty"` // only meaningful when Policy == "weighted"
}

type CandidateStrategy struct {
	ID        string
	Instances []StrategyInstance
	Policy    string // "majority" | "weighted"
}

type StrategyGenerator interface {
	Generate() CandidateStrategy
}

// BuildFromCandidate resolves a candidate into a CombinedStrategy by
// building each instance via the registry's factory system with its own
// params, then wrapping them with the candidate's combination policy.
func BuildFromCandidate(reg *Registry, c CandidateStrategy) (Strategy, error) {
	var strategies []Strategy
	var weights []float64

	for _, inst := range c.Instances {
		s, err := reg.BuildWithParams(inst.Type, inst.Params)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, s)
		weights = append(weights, inst.Weight)
	}

	policy := resolveCombinationPolicy(c.Policy, weights)
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
