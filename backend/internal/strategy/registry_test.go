package strategy_test

import (
	"sort"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

// mockStrategy implements strategy.Strategy for testing purposes
type mockStrategy struct {
	name string
}

func (m *mockStrategy) Name() string {
	return m.name
}

func (m *mockStrategy) Analyze(candles []market.Candle) strategy.Signal {
	return strategy.Hold
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := strategy.NewRegistry()

	// Test getting a non-existent strategy
	if _, ok := registry.Get("NON_EXISTENT"); ok {
		t.Errorf("Expected to not find 'NON_EXISTENT' strategy")
	}

	// Register a mock strategy
	mock := &mockStrategy{name: "MOCK"}
	registry.Register(mock)

	// Test getting the registered strategy
	s, ok := registry.Get("MOCK")
	if !ok {
		t.Fatalf("Expected to find 'MOCK' strategy")
	}

	if s.Name() != "MOCK" {
		t.Errorf("Expected strategy name 'MOCK', got '%s'", s.Name())
	}
}

func TestRegistry_List(t *testing.T) {
	registry := strategy.NewRegistry()

	// Initial list should be empty
	if len(registry.List()) != 0 {
		t.Errorf("Expected empty list initially")
	}

	// Register a few mock strategies
	registry.Register(&mockStrategy{name: "A"})
	registry.Register(&mockStrategy{name: "C"})
	registry.Register(&mockStrategy{name: "B"})

	// List should return all registered names
	list := registry.List()
	if len(list) != 3 {
		t.Fatalf("Expected list of length 3, got %d", len(list))
	}

	// Sort to compare since map iteration order is not guaranteed
	sort.Strings(list)
	if list[0] != "A" || list[1] != "B" || list[2] != "C" {
		t.Errorf("Expected list ['A', 'B', 'C'], got %v", list)
	}
}
