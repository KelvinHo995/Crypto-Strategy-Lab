package experiment_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

func TestRankUsesWeightedScoreAndDoesNotMutateInput(t *testing.T) {
	input := []experiment.Result{
		{ID: "high-return-high-risk", Return: 20, WinRate: 40, MDD: 30, Status: "COMPLETED"},
		{ID: "balanced", Return: 15, WinRate: 70, MDD: 5, Status: "COMPLETED"},
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
