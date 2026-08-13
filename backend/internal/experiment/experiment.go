package experiment

import "github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"

// Experiment defines its own interface (Go idiom: consumer defines the interface)
type Strategy interface {
	Analyze(candles interface{}) strategy.Signal
}

type Result struct {
	ID               string            `json:"id"`
	CandidateID      string            `json:"candidateId"`
	StrategyVersions map[string]string `json:"strategyVersions"`
	DatasetPeriod    string            `json:"datasetPeriod"`
	Return           float64           `json:"return"`
	MDD              float64           `json:"mdd"`
	TradeCount       int               `json:"tradeCount"`
	Status           string            `json:"status"`
	CreatedAt        int64             `json:"createdAt"`
}
