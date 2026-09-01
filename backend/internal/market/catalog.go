package market

import "strings"

var supportedTimeframes = []string{"5m", "15m", "1h", "4h"}

type Info struct {
	Symbol     string   `json:"symbol"`
	BaseAsset  string   `json:"baseAsset"`
	QuoteAsset string   `json:"quoteAsset"`
	Timeframes []string `json:"timeframes"`
}

var defaultCatalog = []Info{
	{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	{Symbol: "ETHUSDT", BaseAsset: "ETH", QuoteAsset: "USDT"},
	{Symbol: "BNBUSDT", BaseAsset: "BNB", QuoteAsset: "USDT"},
	{Symbol: "SOLUSDT", BaseAsset: "SOL", QuoteAsset: "USDT"},
	{Symbol: "XRPUSDT", BaseAsset: "XRP", QuoteAsset: "USDT"},
	{Symbol: "ADAUSDT", BaseAsset: "ADA", QuoteAsset: "USDT"},
	{Symbol: "DOGEUSDT", BaseAsset: "DOGE", QuoteAsset: "USDT"},
	{Symbol: "AVAXUSDT", BaseAsset: "AVAX", QuoteAsset: "USDT"},
}

func DefaultCatalog() []Info {
	out := make([]Info, len(defaultCatalog))
	for i, item := range defaultCatalog {
		out[i] = item
		out[i].Timeframes = SupportedTimeframes()
	}
	return out
}

func SupportedTimeframes() []string {
	return append([]string(nil), supportedTimeframes...)
}

func IsSupportedSymbol(symbol string) bool {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	for _, item := range defaultCatalog {
		if item.Symbol == symbol {
			return true
		}
	}
	return false
}

func IsSupportedTimeframe(timeframe string) bool {
	return validTimeframe(strings.TrimSpace(timeframe))
}
