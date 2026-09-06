package experiment

import (
	"math"
	"sort"
)

// Score rewards return and consistency while penalising drawdown. All inputs
// are percentages, so the weights are directly comparable and reproducible.
func Score(r Result) float64 {
	score := 0.50*r.Return + 0.30*r.WinRate - 0.20*r.MDD
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return math.Inf(-1)
	}
	return score
}

// IsRankEligible identifies results that can be compared fairly and traced
// back to a reproducible experiment. Zero-trade and legacy rows stay available
// as history, but must not outrank an evaluated strategy merely because their
// default metrics are all zero.
func IsRankEligible(r Result) bool {
	return r.Status == "COMPLETED" &&
		r.TradeCount > 0 &&
		len(r.Instances) > 0 &&
		r.Pair != "" &&
		r.Timeframe != ""
}

// Rank returns a copy ordered best-first; callers' slices are never mutated.
// Competitive, reproducible results come first. Completed but ineligible
// history follows, then non-completed jobs while progress is visible.
func Rank(results []Result) []Result {
	ranked := make([]Result, len(results))
	copy(ranked, results)
	sort.SliceStable(ranked, func(i, j int) bool {
		if IsRankEligible(ranked[i]) != IsRankEligible(ranked[j]) {
			return IsRankEligible(ranked[i])
		}
		if ranked[i].Status != ranked[j].Status {
			return ranked[i].Status == "COMPLETED"
		}
		return Score(ranked[i]) > Score(ranked[j])
	})
	return ranked
}
