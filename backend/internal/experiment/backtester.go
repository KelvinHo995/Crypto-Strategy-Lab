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
	AllowShort      bool // opt-in: a Sell signal with no open position opens a short instead of being ignored
}

type Backtester struct{ cfg Config }

func NewBacktester(cfg Config) *Backtester { return &Backtester{cfg: cfg} }

func (b *Backtester) Run(strat Strategy, candles []market.Candle) []Trade {
	if strat == nil || len(candles) == 0 || !b.cfg.valid() {
		return nil
	}
	var trades []Trade
	var pos *openPosition
	capital := b.cfg.StartingCapital

	for i, c := range candles {
		exitedThisCandle := false
		if pos != nil && i > pos.entryIndex {
			if exitPrice, hit := checkStopTarget(pos, c); hit {
				t := b.settle(pos, c.OpenTime, exitPrice)
				trades = append(trades, t)
				capital += t.Profit
				pos = nil
				exitedThisCandle = true
			}
		}
		// A stop/target is resolved using this candle's OHLC. Re-entering at the
		// same candle's Open would travel backwards in time and add lookahead bias.
		if exitedThisCandle {
			continue
		}

		if i < b.cfg.Window {
			continue
		}
		window := candles[i-b.cfg.Window : i] // excludes candle i, avoids lookahead
		signal := strat.Analyze(window)

		switch {
		case signal == strategy.Buy && pos == nil:
			pos = b.open(c, i, capital, Long)
		case signal == strategy.Buy && pos != nil && pos.direction == Short:
			t := b.settle(pos, c.OpenTime, c.Open)
			trades = append(trades, t)
			capital += t.Profit
			pos = nil
		case signal == strategy.Sell && pos != nil && pos.direction == Long:
			t := b.settle(pos, c.OpenTime, c.Open)
			trades = append(trades, t)
			capital += t.Profit
			pos = nil
		case signal == strategy.Sell && pos == nil && b.cfg.AllowShort:
			pos = b.open(c, i, capital, Short)
		}
	}

	if pos != nil {
		last := candles[len(candles)-1]
		t := b.settle(pos, last.OpenTime, last.Close)
		trades = append(trades, t)
	}
	return trades
}

func (c Config) valid() bool {
	return c.StartingCapital > 0 && c.PositionSizePct > 0 && c.PositionSizePct <= 1 &&
		c.StopLossPct >= 0 && c.StopLossPct < 1 && c.TakeProfitPct >= 0 &&
		c.FeePct >= 0 && c.SlippageBps >= 0 && c.Window > 0
}

type openPosition struct {
	entryIndex int
	entryTime  int64
	direction  Direction
	entryPrice float64
	idealEntry float64
	stopLoss   float64
	takeProfit float64
	volumeUSD  float64
	entryFee   float64
}

func (b *Backtester) open(c market.Candle, i int, capital float64, direction Direction) *openPosition {
	ideal := c.Open
	// Opening a long buys at entry; opening a short sells at entry — each gets
	// the fill on the unfavorable side of applySlippage, same as the exit fill below.
	fill := applySlippage(ideal, direction == Long, b.cfg.SlippageBps)
	volumeUSD := capital * b.cfg.PositionSizePct
	stopLoss, takeProfit := fill*(1-b.cfg.StopLossPct), fill*(1+b.cfg.TakeProfitPct)
	if direction == Short {
		// Inverted: a short loses money as price rises, profits as it falls.
		stopLoss, takeProfit = fill*(1+b.cfg.StopLossPct), fill*(1-b.cfg.TakeProfitPct)
	}
	return &openPosition{
		entryIndex: i, entryTime: c.OpenTime, direction: direction,
		entryPrice: fill, idealEntry: ideal,
		stopLoss:   stopLoss,
		takeProfit: takeProfit,
		volumeUSD:  volumeUSD,
		entryFee:   volumeUSD * b.cfg.FeePct,
	}
}

func (b *Backtester) settle(pos *openPosition, exitTime int64, idealExit float64) Trade {
	// Closing a long sells; covering a short buys back.
	fill := applySlippage(idealExit, pos.direction == Short, b.cfg.SlippageBps)
	qty := pos.volumeUSD / pos.entryPrice
	exitValue := qty * fill
	exitFee := exitValue * b.cfg.FeePct
	fees := pos.entryFee + exitFee

	var slippageCost, profit float64
	if pos.direction == Long {
		slippageCost = (pos.entryPrice-pos.idealEntry)*qty + (qty*idealExit - exitValue)
		profit = exitValue - pos.volumeUSD - fees
	} else {
		slippageCost = (pos.idealEntry-pos.entryPrice)*qty + (exitValue - qty*idealExit)
		profit = pos.volumeUSD - exitValue - fees
	}

	return Trade{
		Pair: b.cfg.Pair, EntryTime: pos.entryTime, Direction: pos.direction,
		VolumeUSD: pos.volumeUSD, EntryPrice: pos.entryPrice,
		StopLoss: pos.stopLoss, TakeProfit: pos.takeProfit,
		ExitPrice: fill, ExitTime: exitTime,
		TransactionCost: fees, Slippage: slippageCost, Profit: profit,
	}
}

func checkStopTarget(pos *openPosition, c market.Candle) (exitPrice float64, hit bool) {
	if pos.direction == Short {
		switch {
		case c.Open >= pos.stopLoss: // gap above SL: cannot fill at the better trigger price
			return c.Open, true
		case c.Open <= pos.takeProfit: // gap below TP: fill at the available open
			return c.Open, true
		case c.High >= pos.stopLoss: // SL wins if both hit same candle
			return pos.stopLoss, true
		case c.Low <= pos.takeProfit:
			return pos.takeProfit, true
		default:
			return 0, false
		}
	}
	switch {
	case c.Open <= pos.stopLoss: // gap below SL: cannot fill at the better trigger price
		return c.Open, true
	case c.Open >= pos.takeProfit: // gap above TP: fill at the available open
		return c.Open, true
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
