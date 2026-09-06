package experiment_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type architectureProofStrategy struct{}

func (architectureProofStrategy) Name() string                            { return "ArchitectureProof" }
func (architectureProofStrategy) Analyze([]market.Candle) strategy.Signal { return strategy.Buy }

// This is the executable version of the deck's "add MACD" test: a plugin
// declared outside the strategy package passes through composition, backtest,
// evaluation, and ranking without changes to any of those components.
func TestStrategyPluginRunsThroughExperimentPipeline(t *testing.T) {
	registry := strategy.NewRegistry()
	if err := registry.RegisterPlugin(strategy.Plugin{
		Name: "ArchitectureProof", Default: architectureProofStrategy{},
		Factory:      func(map[string]any) strategy.Strategy { return architectureProofStrategy{} },
		RandomParams: strategy.NoRandomParams, Version: "test-version",
	}); err != nil {
		t.Fatal(err)
	}
	candidate := strategy.CandidateStrategy{ID: "proof", Policy: "majority", Instances: []strategy.StrategyInstance{{Type: "ArchitectureProof"}}}
	plugin, err := strategy.BuildFromCandidate(registry, candidate)
	if err != nil {
		t.Fatal(err)
	}
	candles := []market.Candle{
		{Symbol: "BTCUSDT", Close: 100, Open: 100, High: 101, Low: 99},
		{Symbol: "BTCUSDT", Close: 105, Open: 100, High: 106, Low: 99},
		{Symbol: "BTCUSDT", Close: 110, Open: 105, High: 111, Low: 104},
	}
	trades := experiment.NewBacktester(experiment.Config{Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1, Window: 1}).Run(plugin, candles)
	metrics := (experiment.Evaluator{StartingCapital: 1000}).Evaluate(trades)
	ranked := experiment.Rank([]experiment.Result{{ID: "proof", Return: metrics.Return, MDD: metrics.MDD, WinRate: metrics.WinRate, TradeCount: metrics.TradeCount, Status: "COMPLETED"}})
	if len(trades) != 1 || len(ranked) != 1 || ranked[0].ID != "proof" {
		t.Fatalf("plugin did not cross the full pipeline: trades=%d ranked=%+v", len(trades), ranked)
	}
}
