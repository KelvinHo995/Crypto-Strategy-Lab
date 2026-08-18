package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

// DummyStrategy is a brand new strategy created just for this test
// It proves that we can extend the system without touching core code.
var _ strategy.Strategy = (*DummyStrategy)(nil)

type DummyStrategy struct {
	AlwaysSignal strategy.Signal
}

func (s *DummyStrategy) Name() string {
	return "Dummy"
}

func (s *DummyStrategy) Analyze(candles []market.Candle) strategy.Signal {
	return s.AlwaysSignal
}

func TestStrategy_Extensibility(t *testing.T) {
	// 1. We create the core registry
	registry := strategy.NewRegistry()

	// 2. We register our brand new strategy from the outside
	myDummy := &DummyStrategy{AlwaysSignal: strategy.Buy}
	registry.Register(myDummy)

	// 3. The system can now retrieve and use it
	s, ok := registry.Get("Dummy")
	if !ok {
		t.Fatalf("Expected to find registered Dummy strategy")
	}

	if s.Name() != "Dummy" {
		t.Errorf("Expected name Dummy, got %s", s.Name())
	}

	// 4. It works as expected
	signal := s.Analyze(nil)
	if signal != strategy.Buy {
		t.Errorf("Expected Buy signal, got %v", signal)
	}
}
