package strategy

import (
	"fmt"
)

type CandidateStrategy struct {
	ID         string
	Strategies []string
	Params     map[string]any
	Policy     string // "majority" | "weighted"
}

type StrategyGenerator interface {
	Generate() CandidateStrategy
}

// BuildFromCandidate resolves a candidate into a list of actual Strategy instances.
// Currently it hardcodes creating new instances based on params, but ideally
// we could have a StrategyFactory if parameters are dynamic per candidate.
// For the sake of simplicity, we will instantiate them directly based on name.
func BuildFromCandidate(reg *Registry, c CandidateStrategy) ([]Strategy, error) {
	var strategies []Strategy

	for _, name := range c.Strategies {
		// In a real system we would use c.Params to initialize them.
		// For this lab, since we don't have a factory pattern for dynamic params,
		// we'll just get the pre-registered default instance from registry, or
		// recreate them with default params if we want isolated instances.
		// According to plan, we just resolve them.
		s, ok := reg.Get(name)
		if !ok {
			return nil, fmt.Errorf("strategy %s not found in registry", name)
		}
		strategies = append(strategies, s)
	}

	return strategies, nil
}
