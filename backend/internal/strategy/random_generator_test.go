package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestRandomGenerator_Generate(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.Register(strategy.NewMAStrategy(20, 50))
	registry.Register(strategy.NewRSIStrategy(14, 70, 30))
	registry.Register(strategy.NewBollingerStrategy(20, 2))

	generator := strategy.NewRandomGenerator(registry)

	candidate := generator.Generate()

	if candidate.ID == "" {
		t.Errorf("Expected candidate ID to not be empty")
	}

	if len(candidate.Strategies) == 0 {
		t.Errorf("Expected candidate to have strategies")
	}

	if candidate.Policy != "majority" && candidate.Policy != "weighted" {
		t.Errorf("Expected valid policy, got %s", candidate.Policy)
	}

	// Test BuildFromCandidate
	strat, err := strategy.BuildFromCandidate(registry, candidate)
	if err != nil {
		t.Errorf("BuildFromCandidate failed: %v", err)
	}

	if strat == nil {
		t.Errorf("Expected built strategy, got nil")
	}
}

func TestBuildFromCandidate_Error(t *testing.T) {
	registry := strategy.NewRegistry()
	candidate := strategy.CandidateStrategy{
		ID:         "123",
		Strategies: []string{"NonExistent"},
	}

	_, err := strategy.BuildFromCandidate(registry, candidate)
	if err == nil {
		t.Errorf("Expected error for non-existent strategy")
	}
}
