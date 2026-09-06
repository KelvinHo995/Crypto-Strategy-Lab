package strategy

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	mathrand "math/rand"
	"sync"
)

var _ StrategyGenerator = (*RandomGenerator)(nil)

// MaxGeneratedLookback is the upper bound used by the existing search range
// validation. MA's plugin-owned long-window range currently sets that bound.
const MaxGeneratedLookback = 201

type RandomGenerator struct {
	Registry *Registry
	source   RandomSource
	mu       sync.Mutex
}

func NewRandomGenerator(registry *Registry) *RandomGenerator {
	return &RandomGenerator{
		Registry: registry,
		source:   cryptoRandomSource{},
	}
}

func NewRandomGeneratorWithSeed(registry *Registry, seed int64) *RandomGenerator {
	return &RandomGenerator{Registry: registry, source: mathrand.New(mathrand.NewSource(seed))}
}

func (g *RandomGenerator) Generate() CandidateStrategy {
	g.mu.Lock()
	defer g.mu.Unlock()

	availablePlugins := g.Registry.Plugins()

	// Randomly pick 2 or 3 strategies
	count := 2
	if len(availablePlugins) >= 3 && g.source.Intn(2) == 1 {
		count = 3
	}
	if count > len(availablePlugins) {
		count = len(availablePlugins)
	}

	// Shuffle and take first 'count'
	picked := append([]Plugin(nil), availablePlugins...)
	for i := len(picked) - 1; i > 0; i-- {
		j := g.source.Intn(i + 1)
		picked[i], picked[j] = picked[j], picked[i]
	}
	picked = picked[:count]

	policy := "majority"
	if g.source.Intn(2) == 1 {
		policy = "weighted"
	}

	instances := make([]StrategyInstance, len(picked))
	for i, plugin := range picked {
		instances[i] = StrategyInstance{
			Type:   plugin.Name,
			Params: plugin.RandomParams(g.source),
			Weight: 0.1 + float64(g.source.Intn(90))/100.0, // 0.1-1.0; ignored under majority policy
		}
	}

	return CandidateStrategy{
		ID:        generateID(g.source),
		Instances: instances,
		Policy:    policy,
	}
}

type cryptoRandomSource struct{}

func (cryptoRandomSource) Intn(max int) int {
	if max <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

func generateID(source RandomSource) string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(source.Intn(256))
	}
	return hex.EncodeToString(b)
}
