package experiment_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestRankUsesWeightedScoreAndDoesNotMutateInput(t *testing.T) {
	input := []experiment.Result{
		{ID: "high-return-high-risk", Pair: "BTCUSDT", Timeframe: "5m", Instances: []strategy.StrategyInstance{{Type: "MA"}}, TradeCount: 10, Return: 20, WinRate: 40, MDD: 30, Status: "COMPLETED"},
		{ID: "balanced", Pair: "BTCUSDT", Timeframe: "5m", Instances: []strategy.StrategyInstance{{Type: "RSI"}}, TradeCount: 10, Return: 15, WinRate: 70, MDD: 5, Status: "COMPLETED"},
		{ID: "pending", Return: 100, WinRate: 100, Status: "PENDING"},
	}
	ranked := experiment.Rank(input)
	if ranked[0].ID != "balanced" || ranked[2].ID != "pending" {
		t.Fatalf("rank order = %q, %q, %q", ranked[0].ID, ranked[1].ID, ranked[2].ID)
	}
	if input[0].ID != "high-return-high-risk" {
		t.Fatal("Rank mutated input")
	}
}

func TestRankKeepsNoTradeAndLegacyResultsOutOfCompetitiveRanks(t *testing.T) {
	input := []experiment.Result{
		{ID: "legacy-zero", Status: "COMPLETED"},
		{ID: "real-loss", Pair: "BTCUSDT", Timeframe: "5m", Instances: []strategy.StrategyInstance{{Type: "MA"}}, TradeCount: 12, Return: -20, WinRate: 25, MDD: 30, Status: "COMPLETED"},
		{ID: "failed", TradeCount: 99, Return: 100, Status: "FAILED"},
	}

	ranked := experiment.Rank(input)
	if ranked[0].ID != "real-loss" || ranked[1].ID != "legacy-zero" || ranked[2].ID != "failed" {
		t.Fatalf("rank order = %q, %q, %q", ranked[0].ID, ranked[1].ID, ranked[2].ID)
	}
	if !experiment.IsRankEligible(ranked[0]) || experiment.IsRankEligible(ranked[1]) {
		t.Fatal("rank eligibility classification is incorrect")
	}
}
