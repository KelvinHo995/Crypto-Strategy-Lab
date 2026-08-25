package experiment

import (
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type Strategy interface {
	Analyze(candles []market.Candle) strategy.Signal
}

type Result struct {
	ID               string            `json:"id"`
	CandidateID      string            `json:"candidateId"`
	Strategies       []string          `json:"strategies"`
	Params           map[string]any    `json:"params"`
	Policy           string            `json:"policy"`
	StrategyVersions map[string]string `json:"strategyVersions"`
	DatasetPeriod    string            `json:"datasetPeriod"`
	Return           float64           `json:"return"`
	MDD              float64           `json:"mdd"`
	TradeCount       int               `json:"tradeCount"`
	WinRate          float64           `json:"winRate"`
	Wins             int               `json:"wins"`
	Losses           int               `json:"losses"`
	TotalProfit      float64           `json:"totalProfit"`
	Status           string            `json:"status"`
	CreatedAt        int64             `json:"createdAt"`
}
