package strategy

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

var _ StrategyGenerator = (*RandomGenerator)(nil)

type RandomGenerator struct {
	Registry *Registry
}

func NewRandomGenerator(registry *Registry) *RandomGenerator {
	return &RandomGenerator{
		Registry: registry,
	}
}

func (g *RandomGenerator) Generate() CandidateStrategy {
	availableStrategies := g.Registry.List()
	
	// Randomly pick 2 strategies for this candidate
	var picked []string
	if len(availableStrategies) > 0 {
		idx1 := randomInt(len(availableStrategies))
		picked = append(picked, availableStrategies[idx1])
		
		if len(availableStrategies) > 1 {
			idx2 := randomInt(len(availableStrategies))
			// Just a simple attempt to pick a different one, not strict
			if idx2 == idx1 {
				idx2 = (idx1 + 1) % len(availableStrategies)
			}
			picked = append(picked, availableStrategies[idx2])
		}
	}

	policy := "majority"
	if randomInt(2) == 1 {
		policy = "weighted"
	}

	return CandidateStrategy{
		ID:         generateID(),
		Strategies: picked,
		Params:     map[string]any{}, // Can generate random params here if needed
		Policy:     policy,
	}
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
