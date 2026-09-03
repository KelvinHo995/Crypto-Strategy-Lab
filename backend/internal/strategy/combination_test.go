package strategy

import (
	"testing"
)

func TestMajorityPolicy_Combine(t *testing.T) {
	policy := MajorityPolicy{}

	tests := []struct {
		signals  []Signal
		expected Signal
	}{
		{[]Signal{Buy, Buy, Sell}, Buy},
		{[]Signal{Sell, Sell, Buy}, Sell},
		{[]Signal{Buy, Sell, Hold}, Hold},
		{[]Signal{Hold, Hold, Hold}, Hold},
	}

	for _, tt := range tests {
		if got := policy.Combine(tt.signals); got != tt.expected {
			t.Errorf("Combine(%v) = %v, want %v", tt.signals, got, tt.expected)
		}
	}
}

func TestWeightedPolicy_Combine(t *testing.T) {
	policy := WeightedPolicy{Weights: []float64{0.6, 0.3, 0.1}}

	tests := []struct {
		signals  []Signal
		expected Signal
	}{
		{[]Signal{Buy, Sell, Sell}, Buy},   // 0.6 - 0.3 - 0.1 = 0.2 -> Buy
		{[]Signal{Sell, Buy, Buy}, Sell},   // -0.6 + 0.3 + 0.1 = -0.2 -> Sell
		{[]Signal{Hold, Buy, Sell}, Buy},   // 0 + 0.3 - 0.1 = 0.2 -> Buy
		{[]Signal{Hold, Hold, Hold}, Hold}, // 0 -> Hold
		{[]Signal{Buy, Buy, Sell}, Buy},    // 0.6 + 0.3 - 0.1 = 0.8 -> Buy
	}

	for _, tt := range tests {
		if got := policy.Combine(tt.signals); got != tt.expected {
			t.Errorf("Combine(%v) = %v, want %v", tt.signals, got, tt.expected)
		}
	}
}

func TestCombinedStrategy_MinLookback(t *testing.T) {
	ma := NewMAStrategy(20, 50)       // MinLookback = 51
	rsi := NewRSIStrategy(14, 70, 30) // MinLookback = 15
	combined := NewCombinedStrategy([]Strategy{ma, rsi}, MajorityPolicy{}, "")

	if got := combined.MinLookback(); got != 51 {
		t.Fatalf("MinLookback() = %d, want 51 (max across wrapped strategies, not either one alone)", got)
	}
}

// Before this fix, weighted policy always auto-computed roughly-equal
// weights server-side regardless of what a caller asked for. Custom weights
// must actually change the outcome now, not just be accepted and ignored.
func TestResolveCombinationPolicy_RespectsCustomWeights(t *testing.T) {
	policy := resolveCombinationPolicy("weighted", []float64{0.9, 0.1})
	weighted, ok := policy.(WeightedPolicy)
	if !ok {
		t.Fatalf("policy = %T, want WeightedPolicy", policy)
	}
	if weighted.Weights[0] <= weighted.Weights[1] {
		t.Fatalf("weights = %v, want first much larger than second — custom weights must be respected, not auto-equalized", weighted.Weights)
	}
	// A weak second vote should not be able to override a dominant first vote.
	if got := weighted.Combine([]Signal{Sell, Buy}); got != Sell {
		t.Fatalf("Combine() = %v, want Sell (weight 0.9 on Sell should dominate weight 0.1 on Buy)", got)
	}
}

func TestResolveCombinationPolicy_FallsBackToEqualWeightsWhenUnset(t *testing.T) {
	policy := resolveCombinationPolicy("weighted", []float64{0, 0, 0})
	weighted, ok := policy.(WeightedPolicy)
	if !ok {
		t.Fatal("policy is not WeightedPolicy")
	}
	for _, w := range weighted.Weights {
		if w != 1.0/3.0 {
			t.Fatalf("weights = %v, want equal fallback when none were supplied", weighted.Weights)
		}
	}
}
