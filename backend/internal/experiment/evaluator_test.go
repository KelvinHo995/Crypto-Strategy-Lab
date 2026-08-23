package experiment_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

func TestEvaluator_EmptyTrades(t *testing.T) {
	m := experiment.Evaluator{StartingCapital: 1000}.Evaluate(nil)

	want := experiment.Metrics{}
	if m != want {
		t.Errorf("Evaluate(nil) = %+v, want zero Metrics %+v", m, want)
	}
}

func TestEvaluator_AllWinning(t *testing.T) {
	trades := []experiment.Trade{
		{Profit: 100},
		{Profit: 50},
	}
	m := experiment.Evaluator{StartingCapital: 1000}.Evaluate(trades)

	if !almostEqual(m.WinRate, 100) {
		t.Errorf("WinRate = %v, want 100", m.WinRate)
	}
	if !almostEqual(m.MDD, 0) {
		t.Errorf("MDD = %v, want 0 (equity only ever goes up)", m.MDD)
	}
	if m.Wins != 2 || m.Losses != 0 {
		t.Errorf("Wins/Losses = %d/%d, want 2/0", m.Wins, m.Losses)
	}
}

func TestEvaluator_MixedWinLoss(t *testing.T) {
	trades := []experiment.Trade{
		{Profit: 100},
		{Profit: -40},
		{Profit: 30},
		{Profit: -10},
	}
	m := experiment.Evaluator{StartingCapital: 1000}.Evaluate(trades)

	if m.Wins != 2 || m.Losses != 2 {
		t.Errorf("Wins/Losses = %d/%d, want 2/2", m.Wins, m.Losses)
	}
	if !almostEqual(m.WinRate, 50) {
		t.Errorf("WinRate = %v, want 50", m.WinRate)
	}
	if !almostEqual(m.TotalProfit, 80) {
		t.Errorf("TotalProfit = %v, want 80", m.TotalProfit)
	}
	if m.TradeCount != 4 {
		t.Errorf("TradeCount = %d, want 4", m.TradeCount)
	}
}

func TestEvaluator_MaxDrawdown_PeakToTrough_NotFinalState(t *testing.T) {
	// equity: 1000 -> 1200 (peak) -> 900 (trough, dd=25%) -> 1000 (ends flat)
	// MDD must be 25%, not the ~16.7% drawdown from peak to the final state.
	trades := []experiment.Trade{
		{Profit: 200},
		{Profit: -300},
		{Profit: 100},
	}
	m := experiment.Evaluator{StartingCapital: 1000}.Evaluate(trades)

	if !almostEqual(m.MDD, 25) {
		t.Errorf("MDD = %v, want 25 (largest peak-to-trough drop, not the final drawdown)", m.MDD)
	}
}

func TestEvaluator_Return(t *testing.T) {
	trades := []experiment.Trade{
		{Profit: 150},
		{Profit: -50},
	}
	m := experiment.Evaluator{StartingCapital: 1000}.Evaluate(trades)

	// final equity = 1000 + 150 - 50 = 1100 -> Return = 10%
	if !almostEqual(m.Return, 10) {
		t.Errorf("Return = %v, want 10", m.Return)
	}
	if !almostEqual(m.TotalProfit, 100) {
		t.Errorf("TotalProfit = %v, want 100", m.TotalProfit)
	}
}
