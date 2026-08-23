package experiment

import (
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type Config struct {
	Pair            string
	StartingCapital float64
	PositionSizePct float64 // 1.0 = full compounding
	StopLossPct     float64 // e.g. 0.02 = 2%
	TakeProfitPct   float64 // e.g. 0.04 = 4%
	FeePct          float64 // e.g. 0.001 = 10bps
	SlippageBps     float64 // e.g. 5 = 5bps
	Window          int
}

type Backtester struct{ cfg Config }

func NewBacktester(cfg Config) *Backtester { return &Backtester{cfg: cfg} }

func (b *Backtester) Run(strat Strategy, candles []market.Candle) []Trade {
	var trades []Trade
	var pos *openPosition
	capital := b.cfg.StartingCapital

	for i, c := range candles {
		if pos != nil && i > pos.entryIndex {
			if exitPrice, hit := checkStopTarget(pos, c); hit {
				t := b.settle(pos, c.OpenTime, exitPrice)
				trades = append(trades, t)
				capital += t.Profit
				pos = nil
			}
		}

		if i < b.cfg.Window {
			continue
		}
		window := candles[i-b.cfg.Window : i] // excludes candle i, avoids lookahead
		signal := strat.Analyze(window)

		switch {
		case signal == strategy.Buy && pos == nil:
			pos = b.open(c, i, capital)
		case signal == strategy.Sell && pos != nil:
			t := b.settle(pos, c.OpenTime, c.Open)
			trades = append(trades, t)
			capital += t.Profit
			pos = nil
		}
	}

	if pos != nil {
		last := candles[len(candles)-1]
		t := b.settle(pos, last.OpenTime, last.Close)
		trades = append(trades, t)
	}
	return trades
}

type openPosition struct {
	entryIndex int
	entryTime  int64
	entryPrice float64
	idealEntry float64
	stopLoss   float64
	takeProfit float64
	volumeUSD  float64
	entryFee   float64
}

func (b *Backtester) open(c market.Candle, i int, capital float64) *openPosition {
	ideal := c.Open
	fill := applySlippage(ideal, true, b.cfg.SlippageBps)
	volumeUSD := capital * b.cfg.PositionSizePct
	return &openPosition{
		entryIndex: i, entryTime: c.OpenTime,
		entryPrice: fill, idealEntry: ideal,
		stopLoss:   fill * (1 - b.cfg.StopLossPct),
		takeProfit: fill * (1 + b.cfg.TakeProfitPct),
		volumeUSD:  volumeUSD,
		entryFee:   volumeUSD * b.cfg.FeePct,
	}
}

func (b *Backtester) settle(pos *openPosition, exitTime int64, idealExit float64) Trade {
	fill := applySlippage(idealExit, false, b.cfg.SlippageBps)
	qty := pos.volumeUSD / pos.entryPrice
	exitValue := qty * fill
	exitFee := exitValue * b.cfg.FeePct
	fees := pos.entryFee + exitFee
	slippageCost := (pos.entryPrice-pos.idealEntry)*qty + (qty*idealExit - exitValue)
	profit := exitValue - pos.volumeUSD - fees

	return Trade{
		Pair: b.cfg.Pair, EntryTime: pos.entryTime, Direction: Long,
		VolumeUSD: pos.volumeUSD, EntryPrice: pos.entryPrice,
		StopLoss: pos.stopLoss, TakeProfit: pos.takeProfit,
		ExitPrice: fill, ExitTime: exitTime,
		TransactionCost: fees, Slippage: slippageCost, Profit: profit,
	}
}

func checkStopTarget(pos *openPosition, c market.Candle) (exitPrice float64, hit bool) {
	switch {
	case c.Low <= pos.stopLoss: // SL wins if both hit same candle
		return pos.stopLoss, true
	case c.High >= pos.takeProfit:
		return pos.takeProfit, true
	default:
		return 0, false
	}
}

func applySlippage(price float64, isBuy bool, bps float64) float64 {
	factor := bps / 10000
	if isBuy {
		return price * (1 + factor)
	}
	return price * (1 - factor)
}
