package strategy

// RandomSource is the narrow random-number contract exposed to plugins.
// Tests can supply a seeded source without coupling plugins to math/rand.
type RandomSource interface {
	Intn(max int) int
}

type RandomParamsGenerator func(RandomSource) map[string]any

// Plugin contains the strategy-owned metadata needed by generic discovery,
// candidate generation, construction, and experiment provenance.
type Plugin struct {
	Name         string
	Default      Strategy
	Factory      StrategyFactory
	RandomParams RandomParamsGenerator
	Version      string
}

func NoRandomParams(RandomSource) map[string]any {
	return map[string]any{}
}

type PluginVersionResolver interface {
	VersionsFor([]StrategyInstance) (map[string]string, error)
}
