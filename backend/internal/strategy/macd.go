package strategy

import "github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"

var _ Strategy = (*MACDStrategy)(nil)
var _ LookbackAware = (*MACDStrategy)(nil)

const MACDStrategyVersion = "v1"

type MACDStrategy struct {
	FastPeriod   int
	SlowPeriod   int
	SignalPeriod int
}

func NewMACDStrategy(fastPeriod, slowPeriod, signalPeriod int) *MACDStrategy {
	return &MACDStrategy{FastPeriod: fastPeriod, SlowPeriod: slowPeriod, SignalPeriod: signalPeriod}
}

func MACDFactory(params map[string]any) Strategy {
	fast, _ := toInt(params["macdFastPeriod"], 12)
	slow, _ := toInt(params["macdSlowPeriod"], 26)
	signal, _ := toInt(params["macdSignalPeriod"], 9)
	return NewMACDStrategy(fast, slow, signal)
}

func MACDRandomParams(source RandomSource) map[string]any {
	return map[string]any{
		"macdFastPeriod":   8 + source.Intn(9),
		"macdSlowPeriod":   20 + source.Intn(21),
		"macdSignalPeriod": 5 + source.Intn(8),
	}
}

func NewMACDPlugin(defaultStrategy *MACDStrategy) Plugin {
	return Plugin{Name: "MACD", Default: defaultStrategy, Factory: MACDFactory, RandomParams: MACDRandomParams, Version: MACDStrategyVersion}
}

func (s *MACDStrategy) Name() string { return "MACD" }

func (s *MACDStrategy) MinLookback() int {
	return s.SlowPeriod + s.SignalPeriod
}

func (s *MACDStrategy) Analyze(candles []market.Candle) Signal {
	if s.FastPeriod <= 0 || s.SlowPeriod <= 0 || s.SignalPeriod <= 0 || s.FastPeriod >= s.SlowPeriod || len(candles) < s.MinLookback() {
		return Hold
	}

	closes := make([]float64, len(candles))
	for i := range candles {
		closes[i] = candles[i].Close
	}
	fast := ema(closes, s.FastPeriod)
	slow := ema(closes, s.SlowPeriod)
	macd := make([]float64, len(closes))
	for i := range closes {
		macd[i] = fast[i] - slow[i]
	}
	signal := ema(macd, s.SignalPeriod)
	last := len(macd) - 1
	if macd[last-1] <= signal[last-1] && macd[last] > signal[last] {
		return Buy
	}
	if macd[last-1] >= signal[last-1] && macd[last] < signal[last] {
		return Sell
	}
	return Hold
}

func ema(values []float64, period int) []float64 {
	result := make([]float64, len(values))
	if len(values) == 0 || period <= 0 {
		return result
	}
	multiplier := 2.0 / float64(period+1)
	result[0] = values[0]
	for i := 1; i < len(values); i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}
	return result
}
