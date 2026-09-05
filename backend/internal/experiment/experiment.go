package experiment

import (
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type Strategy interface {
	Analyze(candles []market.Candle) strategy.Signal
}

// MinCandlesForBacktest is the fewest raw candles a requested range must
// have before a search is even attempted. It's tied to
// strategy.MaxGeneratedLookback (not a separate guess) so the two can't
// silently drift apart the way they did before — one extra candle beyond
// the worst-case lookback, so there's always at least one index left to
// actually evaluate once that window is satisfied.
const MinCandlesForBacktest = strategy.MaxGeneratedLookback + 1

type Result struct {
	ID               string                            `json:"id"`
	SearchID         string                            `json:"searchId"`
	SearchTotal      int                               `json:"searchTotal"`
	CandidateID      string                            `json:"candidateId"`
	Pair             string                            `json:"pair,omitempty"`
	Timeframe        string                            `json:"timeframe,omitempty"`
	Instances        []strategy.StrategyInstance       `json:"instances"`
	Policy           string                            `json:"policy"`
	StrategyVersions map[string]string                 `json:"strategyVersions"`
	SentimentModels  []strategy.SentimentModelIdentity `json:"sentimentModels,omitempty"`
	DatasetPeriod    string                            `json:"datasetPeriod"`
	Return           float64                           `json:"return"`
	MDD              float64                           `json:"mdd"`
	TradeCount       int                               `json:"tradeCount"`
	WinRate          float64                           `json:"winRate"`
	Wins             int                               `json:"wins"`
	Losses           int                               `json:"losses"`
	TotalProfit      float64                           `json:"totalProfit"`
	Status           string                            `json:"status"`
	CreatedAt        int64                             `json:"createdAt"`
	UpdatedAt        int64                             `json:"updatedAt"`
}
