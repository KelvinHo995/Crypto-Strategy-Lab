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
	IsClosed  bool    `json:"isClosed"`
}

type TradeTick struct {
	Symbol    string  `json:"symbol"`
	TradeID   int64   `json:"tradeId"`
	TradeTime int64   `json:"tradeTime"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Side      string  `json:"side"`
}

type LiveEvent struct {
	Type   string
	Candle Candle
	Trade  TradeTick
}
