package market

import (
	"strings"
	"testing"
)

func TestCandleUpsertStatement(t *testing.T) {
	candles := []Candle{
		{Symbol: "BTCUSDT", Timeframe: "1h", OpenTime: 1, Open: 10, High: 12, Low: 9, Close: 11, Volume: 3},
		{Symbol: "BTCUSDT", Timeframe: "1h", OpenTime: 2, Open: 11, High: 13, Low: 10, Close: 12, Volume: 4},
	}
	query, args := candleUpsertStatement(candles)
	if len(args) != 16 {
		t.Fatalf("argument count = %d, want 16", len(args))
	}
	for _, fragment := range []string{"($1,$2,$3,$4,$5,$6,$7,$8)", "($9,$10,$11,$12,$13,$14,$15,$16)", "ON CONFLICT"} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query missing %q: %s", fragment, query)
		}
	}
}
