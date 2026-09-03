package strategy

import "testing"

// The whole point of StrategyInstance: two entries of the same type must be
// able to carry independently different params — a flat shared-params-bag
// model couldn't express MA(20) and MA(50) coexisting in one composite.
func TestBuildFromCandidate_SameTypeDifferentParamsCoexist(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFactory("MA", MAFactory)

	candidate := CandidateStrategy{
		ID: "c1",
		Instances: []StrategyInstance{
			{Type: "MA", Params: map[string]any{"maShortWindow": 5, "maLongWindow": 10}},
			{Type: "MA", Params: map[string]any{"maShortWindow": 40, "maLongWindow": 50}},
		},
		Policy: "majority",
	}

	built, err := BuildFromCandidate(reg, candidate)
	if err != nil {
		t.Fatal(err)
	}

	combined, ok := built.(*CombinedStrategy)
	if !ok {
		t.Fatalf("built = %T, want *CombinedStrategy", built)
	}
	if len(combined.Strategies) != 2 {
		t.Fatalf("wrapped strategies = %d, want 2 (both MA instances, independently built)", len(combined.Strategies))
	}

	first, ok := combined.Strategies[0].(*MAStrategy)
	if !ok || first.ShortWindow != 5 || first.LongWindow != 10 {
		t.Fatalf("first MA instance = %+v, want ShortWindow=5 LongWindow=10", combined.Strategies[0])
	}
	second, ok := combined.Strategies[1].(*MAStrategy)
	if !ok || second.ShortWindow != 40 || second.LongWindow != 50 {
		t.Fatalf("second MA instance = %+v, want ShortWindow=40 LongWindow=50 (not collapsed onto the first instance's params)", combined.Strategies[1])
	}
}
