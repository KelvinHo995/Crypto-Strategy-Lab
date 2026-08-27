package experiment

import "sort"

// Score rewards return and consistency while penalising drawdown. All inputs
// are percentages, so the weights are directly comparable and reproducible.
func Score(r Result) float64 {
	return 0.50*r.Return + 0.30*r.WinRate - 0.20*r.MDD
}

// Rank returns a copy ordered best-first; callers' slices are never mutated.
// Non-completed jobs stay below completed results while progress is visible.
func Rank(results []Result) []Result {
	ranked := append([]Result(nil), results...)
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Status != ranked[j].Status {
			return ranked[i].Status == "COMPLETED"
		}
		return Score(ranked[i]) > Score(ranked[j])
	})
	return ranked
}
