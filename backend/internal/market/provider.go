package market

import "context"

type Provider interface {
	FetchHistoricalCandles(ctx context.Context, symbol, timeframe string, from, to int64) ([]Candle, error)
}
type LiveProvider interface {
	StreamLiveCandles(ctx context.Context, symbol, timeframe string) <-chan Candle
}

type CombinedLiveProvider interface {
	StreamMarketEvents(ctx context.Context, symbols, timeframes []string) <-chan LiveEvent
}

type CandleRepository interface {
	Upsert(ctx context.Context, candles []Candle) error
	Range(ctx context.Context, symbol, timeframe string, from, to int64) ([]Candle, error)
}
