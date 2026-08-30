package httpx

import (
	"math"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

// fixtureCandles is the deterministic fallback used only by routers constructed
// without a CandleRepository (unit tests). Production injects Postgres candles.
func fixtureCandles(pair string, from, to int64) []market.Candle {
	const count = 120
	step := (to - from) / (count - 1)
	if step < 1 {
		step = 1
	}

	candles := make([]market.Candle, count)
	for i := 0; i < count; i++ {
		price := 100 + 2*math.Sin(float64(i)/6)
		candles[i] = market.Candle{
			Symbol: pair, Timeframe: "fixture", OpenTime: from + int64(i)*step,
			Open: price, High: price + 1, Low: price - 1, Close: price + 0.5, Volume: 1, IsClosed: true,
		}
	}
	return candles
}
