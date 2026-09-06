package strategy

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

type Signal string

const (
	Buy  Signal = "BUY"
	Sell Signal = "SELL"
	Hold Signal = "HOLD"
)

type Strategy interface {
	Name() string
	Analyze(candles []market.Candle) Signal
}

// LookbackAware lets a strategy report the fewest trailing candles it needs
// before it can produce anything but Hold — the backtester uses this to size
// its sliding window per-candidate instead of guessing one shared constant.
type LookbackAware interface {
	MinLookback() int
}

type StrategyFactory func(params map[string]any) Strategy

// Registry holds registered strategy implementations and factories.
// It is NOT safe for concurrent use. All Register calls must
// happen during application startup before any concurrent Get/List.
type Registry struct {
	strategies map[string]Strategy
	factories  map[string]StrategyFactory
	plugins    map[string]Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		strategies: make(map[string]Strategy),
		factories:  make(map[string]StrategyFactory),
		plugins:    make(map[string]Plugin),
	}
}

func (r *Registry) Register(s Strategy) {
	r.strategies[s.Name()] = s
}

func (r *Registry) RegisterFactory(name string, f StrategyFactory) {
	r.factories[name] = f
}

// RegisterPlugin atomically installs one complete source-level plugin.
func (r *Registry) RegisterPlugin(plugin Plugin) error {
	name := strings.TrimSpace(plugin.Name)
	if name == "" {
		return fmt.Errorf("register strategy plugin: empty name")
	}
	if plugin.Default == nil || isNilStrategy(plugin.Default) {
		return fmt.Errorf("register strategy plugin %s: nil strategy", name)
	}
	if plugin.Default.Name() != name {
		return fmt.Errorf("register strategy plugin %s: default strategy is named %s", name, plugin.Default.Name())
	}
	if plugin.Factory == nil {
		return fmt.Errorf("register strategy plugin %s: nil factory", name)
	}
	if plugin.RandomParams == nil {
		return fmt.Errorf("register strategy plugin %s: nil random parameter generator", name)
	}
	if strings.TrimSpace(plugin.Version) == "" {
		return fmt.Errorf("register strategy plugin %s: empty version", name)
	}
	if _, exists := r.strategies[name]; exists {
		return fmt.Errorf("register strategy plugin %s: duplicate name", name)
	}
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("register strategy plugin %s: duplicate factory", name)
	}
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("register strategy plugin %s: duplicate metadata", name)
	}
	plugin.Name = name
	plugin.Version = strings.TrimSpace(plugin.Version)
	r.strategies[name] = plugin.Default
	r.factories[name] = plugin.Factory
	r.plugins[name] = plugin
	return nil
}

func isNilStrategy(value Strategy) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (r *Registry) Plugins() []Plugin {
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Name < plugins[j].Name })
	return plugins
}

func (r *Registry) VersionsFor(instances []StrategyInstance) (map[string]string, error) {
	versions := make(map[string]string, len(instances))
	for _, instance := range instances {
		plugin, ok := r.plugins[instance.Type]
		if !ok {
			return nil, fmt.Errorf("strategy plugin metadata for %s is not registered", instance.Type)
		}
		versions[instance.Type] = plugin.Version
	}
	return versions, nil
}

func (r *Registry) Get(name string) (Strategy, bool) {
	s, ok := r.strategies[name]
	return s, ok
}

func (r *Registry) BuildWithParams(name string, params map[string]any) (Strategy, error) {
	if f, ok := r.factories[name]; ok {
		return f(params), nil
	}
	// Fallback: try getting default from registry
	if s, ok := r.strategies[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("strategy %s not found in registry", name)
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
