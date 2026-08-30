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
