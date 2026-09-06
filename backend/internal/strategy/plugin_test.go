package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type cyclingSource struct{ next int }

func (s *cyclingSource) Intn(max int) int {
	if max <= 1 {
		return 0
	}
	value := s.next % max
	s.next++
	return value
}

func TestBuiltInPluginsGenerateFactoryValidParameters(t *testing.T) {
	tests := []struct {
		plugin strategy.Plugin
		valid  func(strategy.Strategy) bool
	}{
		{strategy.NewMAPlugin(strategy.NewMAStrategy(20, 50)), func(s strategy.Strategy) bool {
			v, ok := s.(*strategy.MAStrategy)
			return ok && v.ShortWindow > 0 && v.ShortWindow < v.LongWindow
		}},
		{strategy.NewRSIPlugin(strategy.NewRSIStrategy(14, 70, 30)), func(s strategy.Strategy) bool {
			v, ok := s.(*strategy.RSIStrategy)
			return ok && v.Period > 0 && v.OversoldThreshold >= 0 && v.OversoldThreshold < v.OverboughtThreshold && v.OverboughtThreshold <= 100
		}},
		{strategy.NewBollingerPlugin(strategy.NewBollingerStrategy(20, 2)), func(s strategy.Strategy) bool {
			v, ok := s.(*strategy.BollingerStrategy)
			return ok && v.Period > 0 && v.StdDevMultiplier > 0
		}},
		{strategy.NewSRPlugin(strategy.NewSRStrategy(20, 0.005)), func(s strategy.Strategy) bool {
			v, ok := s.(*strategy.SRStrategy)
			return ok && v.Window > 0 && v.Tolerance >= 0
		}},
		{strategy.NewSMCPlugin(strategy.NewSMCStrategy(10)), func(s strategy.Strategy) bool {
			v, ok := s.(*strategy.SMCStrategy)
			return ok && v.SwingLookback > 0
		}},
		{strategy.NewMACDPlugin(strategy.NewMACDStrategy(12, 26, 9)), func(s strategy.Strategy) bool {
			v, ok := s.(*strategy.MACDStrategy)
			return ok && v.FastPeriod > 0 && v.FastPeriod < v.SlowPeriod && v.SignalPeriod > 0
		}},
		{strategy.NewSentimentPlugin(nil, 0.7), func(s strategy.Strategy) bool {
			_, ok := s.(*strategy.SentimentStrategy)
			return ok
		}},
	}

	for _, test := range tests {
		t.Run(test.plugin.Name, func(t *testing.T) {
			source := &cyclingSource{}
			for i := 0; i < 200; i++ {
				params := test.plugin.RandomParams(source)
				if built := test.plugin.Factory(params); !test.valid(built) {
					t.Fatalf("factory rejected generated params on iteration %d: %+v -> %+v", i, params, built)
				}
			}
		})
	}
}

func TestRegisterPluginRejectsIncompleteDescriptors(t *testing.T) {
	valid := testMomentumPlugin()
	var typedNil *testMomentumStrategy
	tests := []struct {
		name   string
		mutate func(*strategy.Plugin)
	}{
		{"empty name", func(p *strategy.Plugin) { p.Name = "" }},
		{"nil default", func(p *strategy.Plugin) { p.Default = nil }},
		{"typed nil default", func(p *strategy.Plugin) { p.Default = typedNil }},
		{"mismatched name", func(p *strategy.Plugin) { p.Name = "Different" }},
		{"nil factory", func(p *strategy.Plugin) { p.Factory = nil }},
		{"nil random params", func(p *strategy.Plugin) { p.RandomParams = nil }},
		{"empty version", func(p *strategy.Plugin) { p.Version = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := valid
			test.mutate(&plugin)
			if err := strategy.NewRegistry().RegisterPlugin(plugin); err == nil {
				t.Fatal("invalid plugin registration succeeded")
			}
		})
	}
}

func TestRegistryVersionsComeFromPluginMetadata(t *testing.T) {
	registry := strategy.NewRegistry()
	plugin := testMomentumPlugin()
	if err := registry.RegisterPlugin(plugin); err != nil {
		t.Fatal(err)
	}
	versions, err := registry.VersionsFor([]strategy.StrategyInstance{{Type: plugin.Name}})
	if err != nil {
		t.Fatal(err)
	}
	if versions[plugin.Name] != plugin.Version {
		t.Fatalf("versions = %v, want descriptor version %q", versions, plugin.Version)
	}
	if _, err := registry.VersionsFor([]strategy.StrategyInstance{{Type: "Missing"}}); err == nil {
		t.Fatal("missing plugin metadata did not return an error")
	}
}
