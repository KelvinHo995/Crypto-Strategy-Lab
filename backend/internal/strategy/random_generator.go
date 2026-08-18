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

	// Randomly pick 2 or 3 strategies
	count := 2
	if len(availableStrategies) >= 3 && randomInt(2) == 1 {
		count = 3
	}
	if count > len(availableStrategies) {
		count = len(availableStrategies)
	}

	// Shuffle and take first 'count'
	picked := make([]string, len(availableStrategies))
	copy(picked, availableStrategies)
	for i := len(picked) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		picked[i], picked[j] = picked[j], picked[i]
	}
	picked = picked[:count]

	policy := "majority"
	if randomInt(2) == 1 {
		policy = "weighted"
	}

	params := generateRandomParams(picked)

	return CandidateStrategy{
		ID:         generateID(),
		Strategies: picked,
		Params:     params,
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
	if _, err := rand.Read(b); err != nil {
		// Fallback: should never happen with crypto/rand
		return "fallback-id"
	}
	return hex.EncodeToString(b)
}

func generateRandomParams(strategies []string) map[string]any {
	params := map[string]any{}
	for _, name := range strategies {
		switch name {
		case "MA":
			params["maShortWindow"] = 5 + randomInt(46)   // 5-50
			params["maLongWindow"] = 50 + randomInt(151)   // 50-200
		case "RSI":
			params["rsiPeriod"] = 7 + randomInt(22)        // 7-28
			params["rsiOverbought"] = float64(65 + randomInt(16)) // 65-80
			params["rsiOversold"] = float64(20 + randomInt(16))   // 20-35
		case "Bollinger":
			params["bollingerPeriod"] = 10 + randomInt(31)        // 10-40
			params["bollingerStdDev"] = 1.5 + float64(randomInt(15))/10.0 // 1.5-3.0
		}
	}
	return params
}
