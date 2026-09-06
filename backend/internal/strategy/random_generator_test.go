package strategy_test

import (
	"reflect"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type testMomentumStrategy struct{ window int }

var _ strategy.Strategy = (*testMomentumStrategy)(nil)

func (s *testMomentumStrategy) Name() string { return "TestMomentum" }
func (s *testMomentumStrategy) Analyze(candles []market.Candle) strategy.Signal {
	if len(candles) < s.window || candles[len(candles)-1].Close <= candles[len(candles)-2].Close {
		return strategy.Hold
	}
	return strategy.Buy
}

func testMomentumFactory(params map[string]any) strategy.Strategy {
	window, ok := params["window"].(int)
	if !ok {
		window = 2
	}
	return &testMomentumStrategy{window: window}
}

func testMomentumRandomParams(source strategy.RandomSource) map[string]any {
	return map[string]any{"window": 2 + source.Intn(9)}
}

func testMomentumPlugin() strategy.Plugin {
	return strategy.Plugin{
		Name:         "TestMomentum",
		Default:      &testMomentumStrategy{window: 2},
		Factory:      testMomentumFactory,
		RandomParams: testMomentumRandomParams,
		Version:      "test-version",
	}
}

func TestPluginExtensionPathGeneratesBuildsAndExecutesTestMomentum(t *testing.T) {
	registry := strategy.NewRegistry()
	if err := registry.RegisterPlugin(testMomentumPlugin()); err != nil {
		t.Fatal(err)
	}

	generator := strategy.NewRandomGeneratorWithSeed(registry, 41)

	candidate := generator.Generate()
	if len(candidate.Instances) != 1 || candidate.Instances[0].Type != "TestMomentum" {
		t.Fatalf("generated instances = %+v, want TestMomentum", candidate.Instances)
	}
	window, ok := candidate.Instances[0].Params["window"].(int)
	if !ok || window < 2 || window > 10 {
		t.Fatalf("generated params = %+v, want valid momentum window", candidate.Instances[0].Params)
	}
	built, err := strategy.BuildFromCandidate(registry, candidate)
	if err != nil {
		t.Fatal(err)
	}
	candles := make([]market.Candle, window)
	for i := range candles {
		candles[i].Close = float64(i + 1)
	}
	if got := built.Analyze(candles); got != strategy.Buy {
		t.Fatalf("composite signal = %s, want BUY", got)
	}
}

func TestRandomGeneratorSupportsEightPluginsWithoutCountAssumptions(t *testing.T) {
	registry := registryWithBuiltins(t)
	if err := registry.RegisterPlugin(testMomentumPlugin()); err != nil {
		t.Fatal(err)
	}
	if got := len(registry.Plugins()); got != 8 {
		t.Fatalf("plugins = %d, want 8", got)
	}
	generator := strategy.NewRandomGeneratorWithSeed(registry, 8)
	foundMomentum := false
	for i := 0; i < 100; i++ {
		candidate := generator.Generate()
		if _, err := strategy.BuildFromCandidate(registry, candidate); err != nil {
			t.Fatalf("build candidate %d: %v", i, err)
		}
		for _, instance := range candidate.Instances {
			foundMomentum = foundMomentum || instance.Type == "TestMomentum"
		}
	}
	if !foundMomentum {
		t.Fatal("registered TestMomentum was never selected")
	}
}

func TestRandomGeneratorIsDeterministicWithSeed(t *testing.T) {
	firstRegistry := registryWithBuiltins(t)
	secondRegistry := strategy.NewRegistry()
	plugins := firstRegistry.Plugins()
	for i := len(plugins) - 1; i >= 0; i-- {
		if err := secondRegistry.RegisterPlugin(plugins[i]); err != nil {
			t.Fatal(err)
		}
	}
	first := strategy.GenerateCandidates(strategy.NewRandomGeneratorWithSeed(firstRegistry, 99), 25)
	second := strategy.GenerateCandidates(strategy.NewRandomGeneratorWithSeed(secondRegistry, 99), 25)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed and plugin set produced different candidates")
	}
}

func registryWithBuiltins(t *testing.T) *strategy.Registry {
	t.Helper()
	registry := strategy.NewRegistry()
	plugins := []strategy.Plugin{
		strategy.NewMAPlugin(strategy.NewMAStrategy(20, 50)),
		strategy.NewRSIPlugin(strategy.NewRSIStrategy(14, 70, 30)),
		strategy.NewBollingerPlugin(strategy.NewBollingerStrategy(20, 2)),
		strategy.NewSRPlugin(strategy.NewSRStrategy(20, 0.005)),
		strategy.NewSMCPlugin(strategy.NewSMCStrategy(10)),
		strategy.NewMACDPlugin(strategy.NewMACDStrategy(12, 26, 9)),
		strategy.NewSentimentPlugin(nil, 0.7),
	}
	for _, plugin := range plugins {
		if err := registry.RegisterPlugin(plugin); err != nil {
			t.Fatalf("register %s: %v", plugin.Name, err)
		}
	}
	return registry
}

func TestBuildFromCandidate_Error(t *testing.T) {
	registry := strategy.NewRegistry()
	candidate := strategy.CandidateStrategy{
		ID:        "123",
		Instances: []strategy.StrategyInstance{{Type: "NonExistent"}},
	}

	_, err := strategy.BuildFromCandidate(registry, candidate)
	if err == nil {
		t.Errorf("Expected error for non-existent strategy")
	}
}
