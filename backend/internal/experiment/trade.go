package experiment

type Direction string

const Long Direction = "LONG"

type Trade struct {
	Pair            string    `json:"pair"`
	EntryTime       int64     `json:"entryTime"`
	Direction       Direction `json:"direction"`
	VolumeUSD       float64   `json:"volumeUsd"`
	EntryPrice      float64   `json:"entryPrice"`
	StopLoss        float64   `json:"stopLoss"`
	TakeProfit      float64   `json:"takeProfit"`
	ExitPrice       float64   `json:"exitPrice"`
	ExitTime        int64     `json:"exitTime"`
	TransactionCost float64   `json:"transactionCost"` // $, entry+exit combined
	Slippage        float64   `json:"slippage"`        // $, entry+exit combined
	Profit          float64   `json:"profit"`          // $, net of cost+slippage
}
