package market

type Candle struct {
	Symbol    string  `json:"symbol"`
	Timeframe string  `json:"timeframe"`
	OpenTime  int64   `json:"openTime"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}
