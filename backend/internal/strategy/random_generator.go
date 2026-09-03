package strategy

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

var _ StrategyGenerator = (*RandomGenerator)(nil)

// MaxGeneratedLookback is the largest MinLookback any candidate this
// generator can produce will ever report — MA's maLongWindow tops out at
// 200 below, so MinLookback (LongWindow+1) tops out at 201. Bump this if
// the ranges in generateRandomParams change.
const MaxGeneratedLookback = 201

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

	instances := make([]StrategyInstance, len(picked))
	for i, name := range picked {
		instances[i] = StrategyInstance{
			Type:   name,
			Params: generateRandomParamsFor(name),
			Weight: 0.1 + float64(randomInt(90))/100.0, // 0.1-1.0; ignored under majority policy
		}
	}

	return CandidateStrategy{
		ID:        generateID(),
		Instances: instances,
		Policy:    policy,
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

func generateRandomParamsFor(name string) map[string]any {
	switch name {
	case "MA":
		return map[string]any{
			"maShortWindow": 5 + randomInt(46),   // 5-50
			"maLongWindow":  50 + randomInt(151), // 50-200
		}
	case "RSI":
		return map[string]any{
			"rsiPeriod":     7 + randomInt(22),           // 7-28
			"rsiOverbought": float64(65 + randomInt(16)), // 65-80
			"rsiOversold":   float64(20 + randomInt(16)), // 20-35
		}
	case "Bollinger":
		return map[string]any{
			"bollingerPeriod": 10 + randomInt(31),                // 10-40
			"bollingerStdDev": 1.5 + float64(randomInt(15))/10.0, // 1.5-3.0
		}
	case "SR":
		return map[string]any{
			"srWindow":    10 + randomInt(41),                    // 10-50
			"srTolerance": 0.001 + float64(randomInt(10))/1000.0, // 0.1% to 1%
		}
	case "SMC":
		return map[string]any{
			"smcLookback": 5 + randomInt(26), // 5-30
		}
	default:
		return map[string]any{}
	}
}
