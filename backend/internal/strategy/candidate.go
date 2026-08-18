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
		s, err := buildStrategy(name, c.Params)
		if err != nil {
			// Fallback: try getting default from registry
			defaultS, ok := reg.Get(name)
			if !ok {
				return nil, fmt.Errorf("strategy %s not found in registry", name)
			}
			s = defaultS
		}
		strategies = append(strategies, s)
	}

	return strategies, nil
}

func buildStrategy(name string, params map[string]any) (Strategy, error) {
	switch name {
	case "MA":
		short, _ := toInt(params["maShortWindow"], 20)
		long, _ := toInt(params["maLongWindow"], 50)
		return NewMAStrategy(short, long), nil
	case "RSI":
		period, _ := toInt(params["rsiPeriod"], 14)
		ob, _ := toFloat(params["rsiOverbought"], 70.0)
		os, _ := toFloat(params["rsiOversold"], 30.0)
		return NewRSIStrategy(period, ob, os), nil
	case "Bollinger":
		period, _ := toInt(params["bollingerPeriod"], 20)
		sd, _ := toFloat(params["bollingerStdDev"], 2.0)
		return NewBollingerStrategy(period, sd), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
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
