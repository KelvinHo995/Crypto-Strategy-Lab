package strategy

func GenerateCandidates(gen StrategyGenerator, count int) []CandidateStrategy {
	candidates := make([]CandidateStrategy, count)
	for i := range candidates {
		candidates[i] = gen.Generate()
	}
	return candidates
}
