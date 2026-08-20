package strategy

type CandidateStrategy struct {
	ID         string
	Strategies []string
	Params     map[string]any
	Policy     string // "majority" | "weighted"
}
