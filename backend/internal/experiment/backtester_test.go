package experiment_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

const eps = 1e-6

func almostEqual(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}

// returns signals in order, then Hold
type scriptedStrategy struct {
	signals []strategy.Signal
	calls   int
}

func (s *scriptedStrategy) Analyze(_ []market.Candle) strategy.Signal {
	if s.calls >= len(s.signals) {
		return strategy.Hold
	}
	sig := s.signals[s.calls]
	s.calls++
	return sig
}

// records windows seen, to check lookahead avoidance
type windowCapturingStrategy struct {
	seen [][]market.Candle
}

func (s *windowCapturingStrategy) Analyze(c []market.Candle) strategy.Signal {
	s.seen = append(s.seen, c)
	return strategy.Hold
}

func candle(openTime int64, open, high, low, close float64) market.Candle {
	return market.Candle{
		Symbol: "BTCUSDT", Timeframe: "1h", OpenTime: openTime,
		Open: open, High: high, Low: low, Close: close, Volume: 1,
	}
}

func TestBacktester_BuySell_NoFeeNoSlippage(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104),  // entry candle: fill at Open=100
		candle(3, 110, 111, 109, 110), // exit candle: fill at Open=110
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy, strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	want := experiment.Trade{
		Pair: "BTCUSDT", EntryTime: 2, Direction: experiment.Long,
		VolumeUSD: 1000, EntryPrice: 100, StopLoss: 50, TakeProfit: 150,
		ExitPrice: 110, ExitTime: 3, TransactionCost: 0, Slippage: 0, Profit: 100,
	}
	assertTradeEqual(t, tr, want)
}

func TestBacktester_BuySell_WithFeeAndSlippage(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1010, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0.01, SlippageBps: 100, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104),
		candle(3, 110, 111, 109, 110),
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy, strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	want := experiment.Trade{
		Pair: "BTCUSDT", EntryTime: 2, Direction: experiment.Long,
		VolumeUSD: 1010, EntryPrice: 101, StopLoss: 50.5, TakeProfit: 151.5,
		ExitPrice: 108.9, ExitTime: 3, TransactionCost: 20.99, Slippage: 21, Profit: 58.01,
	}
	assertTradeEqual(t, tr, want)
}

func TestBacktester_StopLossHit(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.05, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104), // entry at 100 -> SL=95, TP=150
		candle(3, 98, 99, 90, 95),    // Low=90 breaches SL=95
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	if !almostEqual(tr.ExitPrice, 95) {
		t.Errorf("ExitPrice = %v, want 95 (the SL trigger price)", tr.ExitPrice)
	}
	if !almostEqual(tr.Profit, -50) {
		t.Errorf("Profit = %v, want -50", tr.Profit)
	}
}

func TestBacktester_TakeProfitHit(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.05, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104),  // entry at 100 -> SL=50, TP=105
		candle(3, 102, 108, 101, 106), // High=108 breaches TP=105
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	if !almostEqual(tr.ExitPrice, 105) {
		t.Errorf("ExitPrice = %v, want 105 (the TP trigger price)", tr.ExitPrice)
	}
	if !almostEqual(tr.Profit, 50) {
		t.Errorf("Profit = %v, want 50", tr.Profit)
	}
}

func TestBacktester_BothSLAndTPHit_SLWins(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.05, TakeProfitPct: 0.05, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104), // entry at 100 -> SL=95, TP=105
		candle(3, 100, 110, 90, 100), // both Low<=95 and High>=105 on the same candle
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	if !almostEqual(trades[0].ExitPrice, 95) {
		t.Errorf("ExitPrice = %v, want 95 (SL must win the tie-break, not TP=105)", trades[0].ExitPrice)
	}
}

func TestBacktester_StopGapUsesOpenAndDoesNotReenterSameCandle(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.05, TakeProfitPct: 0.5, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100), candle(1, 100, 101, 99, 100),
		candle(2, 100, 101, 99, 100), // buy at 100, SL=95
		candle(3, 90, 94, 88, 91),    // gaps through SL
		candle(4, 92, 93, 91, 92),
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy, strategy.Buy}}
	trades := experiment.NewBacktester(cfg).Run(strat, candles)
	if len(trades) != 2 {
		t.Fatalf("got %d trades, want gap exit plus a later re-entry", len(trades))
	}
	if !almostEqual(trades[0].ExitPrice, 90) {
		t.Fatalf("ExitPrice = %v, want gap open 90", trades[0].ExitPrice)
	}
	if strat.calls != 2 {
		t.Fatalf("Analyze calls = %d, want 2 (no call on stop candle)", strat.calls)
	}
	if trades[1].EntryTime != 4 {
		t.Fatalf("second entry time = %d, want 4 (must not re-enter on stop candle 3)", trades[1].EntryTime)
	}
}

func TestBacktester_InvalidInputsReturnNoTrades(t *testing.T) {
	valid := experiment.Config{StartingCapital: 1000, PositionSizePct: 1, StopLossPct: .02, TakeProfitPct: .04, Window: 2}
	candles := []market.Candle{candle(0, 100, 101, 99, 100), candle(1, 100, 101, 99, 100), candle(2, 100, 101, 99, 100)}
	if got := experiment.NewBacktester(valid).Run(nil, candles); len(got) != 0 {
		t.Fatal("nil strategy should not trade")
	}
	invalid := valid
	invalid.PositionSizePct = 2
	if got := experiment.NewBacktester(invalid).Run(&scriptedStrategy{signals: []strategy.Signal{strategy.Buy}}, candles); len(got) != 0 {
		t.Fatal("invalid config should not trade")
	}
}

func TestBacktester_SellWithNoPosition_NoOp(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 101, 99, 100),
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 0 {
		t.Fatalf("got %d trades, want 0 (Sell with no open position must be a no-op)", len(trades))
	}
}

func TestBacktester_BuyWhilePositioned_NoPyramiding(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104),  // entry at 100
		candle(3, 102, 103, 101, 102), // second Buy here must be a no-op (already positioned)
		candle(4, 108, 109, 107, 108), // Sell here closes the single position
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy, strategy.Buy, strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1 (a repeated Buy signal must not open a second position)", len(trades))
	}
	tr := trades[0]
	if tr.EntryTime != 2 || tr.ExitTime != 4 {
		t.Errorf("EntryTime/ExitTime = %v/%v, want 2/4", tr.EntryTime, tr.ExitTime)
	}
	if !almostEqual(tr.Profit, 80) {
		t.Errorf("Profit = %v, want 80", tr.Profit)
	}
}

func TestBacktester_OpenPositionAtDatasetEnd_ForceClosed(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104),  // entry at 100, no Sell ever comes
		candle(3, 105, 106, 104, 112), // dataset ends here, still positioned
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Buy}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1 (an open position at dataset end must be force-closed)", len(trades))
	}
	tr := trades[0]
	if !almostEqual(tr.ExitPrice, 112) {
		t.Errorf("ExitPrice = %v, want 112 (the final candle's Close)", tr.ExitPrice)
	}
	if tr.ExitTime != 3 {
		t.Errorf("ExitTime = %v, want 3 (the final candle's OpenTime)", tr.ExitTime)
	}
	if !almostEqual(tr.Profit, 120) {
		t.Errorf("Profit = %v, want 120", tr.Profit)
	}
}

func TestBacktester_ShortSellCover_NoFeeNoSlippage(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
		AllowShort: true,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104), // entry candle: short opened at Open=100
		candle(3, 90, 91, 89, 90),    // exit candle: covered at Open=90
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Sell, strategy.Buy}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	want := experiment.Trade{
		Pair: "BTCUSDT", EntryTime: 2, Direction: experiment.Short,
		VolumeUSD: 1000, EntryPrice: 100, StopLoss: 150, TakeProfit: 50,
		ExitPrice: 90, ExitTime: 3, TransactionCost: 0, Slippage: 0, Profit: 100,
	}
	assertTradeEqual(t, tr, want)
}

func TestBacktester_ShortSell_NoOpWhenAllowShortDisabled(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
		// AllowShort intentionally left false (default) — Sell with no position must stay a no-op.
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 101, 99, 100),
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 0 {
		t.Fatalf("got %d trades, want 0 (AllowShort=false must keep Sell-with-no-position a no-op)", len(trades))
	}
}

func TestBacktester_ShortStopLossHit(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.05, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 2,
		AllowShort: true,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104),  // short entry at 100 -> SL=105, TP=50
		candle(3, 102, 110, 101, 108), // High=110 breaches SL=105
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	if !almostEqual(tr.ExitPrice, 105) {
		t.Errorf("ExitPrice = %v, want 105 (the short's SL trigger price)", tr.ExitPrice)
	}
	if !almostEqual(tr.Profit, -50) {
		t.Errorf("Profit = %v, want -50 (price rose against the short)", tr.Profit)
	}
}

func TestBacktester_ShortTakeProfitHit(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.05, FeePct: 0, SlippageBps: 0, Window: 2,
		AllowShort: true,
	}
	candles := []market.Candle{
		candle(0, 100, 101, 99, 100),
		candle(1, 100, 101, 99, 100),
		candle(2, 100, 105, 99, 104), // short entry at 100 -> SL=150, TP=95
		candle(3, 98, 99, 92, 94),    // Low=92 breaches TP=95
	}
	strat := &scriptedStrategy{signals: []strategy.Signal{strategy.Sell}}

	trades := experiment.NewBacktester(cfg).Run(strat, candles)

	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	tr := trades[0]
	if !almostEqual(tr.ExitPrice, 95) {
		t.Errorf("ExitPrice = %v, want 95 (the short's TP trigger price)", tr.ExitPrice)
	}
	if !almostEqual(tr.Profit, 50) {
		t.Errorf("Profit = %v, want 50 (price fell in the short's favor)", tr.Profit)
	}
}

func TestBacktester_WindowExcludesCurrentCandle(t *testing.T) {
	cfg := experiment.Config{
		Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
		StopLossPct: 0.5, TakeProfitPct: 0.5, FeePct: 0, SlippageBps: 0, Window: 3,
	}
	var candles []market.Candle
	for i := int64(0); i < 6; i++ {
		candles = append(candles, candle(i, 100, 101, 99, 100))
	}
	strat := &windowCapturingStrategy{}

	experiment.NewBacktester(cfg).Run(strat, candles)

	if len(strat.seen) != 3 {
		t.Fatalf("Analyze called %d times, want 3", len(strat.seen))
	}
	for k, window := range strat.seen {
		i := int64(3 + k)
		if len(window) != cfg.Window {
			t.Errorf("call %d: window length = %d, want %d", k, len(window), cfg.Window)
		}
		last := window[len(window)-1]
		if last.OpenTime != i-1 {
			t.Errorf("call %d: window's last candle OpenTime = %d, want %d", k, last.OpenTime, i-1)
		}
	}
}

func assertTradeEqual(t *testing.T, got, want experiment.Trade) {
	t.Helper()
	if got.Pair != want.Pair || got.Direction != want.Direction {
		t.Errorf("Pair/Direction = %v/%v, want %v/%v", got.Pair, got.Direction, want.Pair, want.Direction)
	}
	if !almostEqual(float64(got.EntryTime), float64(want.EntryTime)) || !almostEqual(float64(got.ExitTime), float64(want.ExitTime)) {
		t.Errorf("EntryTime/ExitTime = %v/%v, want %v/%v", got.EntryTime, got.ExitTime, want.EntryTime, want.ExitTime)
	}
	fields := []struct {
		name      string
		got, want float64
	}{
		{"VolumeUSD", got.VolumeUSD, want.VolumeUSD},
		{"EntryPrice", got.EntryPrice, want.EntryPrice},
		{"StopLoss", got.StopLoss, want.StopLoss},
		{"TakeProfit", got.TakeProfit, want.TakeProfit},
		{"ExitPrice", got.ExitPrice, want.ExitPrice},
		{"TransactionCost", got.TransactionCost, want.TransactionCost},
		{"Slippage", got.Slippage, want.Slippage},
		{"Profit", got.Profit, want.Profit},
	}
	for _, f := range fields {
		if !almostEqual(f.got, f.want) {
			t.Errorf("%s = %v, want %v", f.name, f.got, f.want)
		}
	}
}
