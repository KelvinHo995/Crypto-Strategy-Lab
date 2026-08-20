package experiment

type Metrics struct {
	Return      float64 // %
	MDD         float64 // %, positive number
	TradeCount  int
	WinRate     float64 // %
	Wins        int
	Losses      int
	TotalProfit float64 // $
}

type Evaluator struct{ StartingCapital float64 }

func (e Evaluator) Evaluate(trades []Trade) Metrics {
	m := Metrics{TradeCount: len(trades)}
	equity, peak := e.StartingCapital, e.StartingCapital
	maxDD := 0.0

	for _, t := range trades {
		m.TotalProfit += t.Profit
		if t.Profit > 0 {
			m.Wins++
		} else if t.Profit < 0 {
			m.Losses++
		}
		equity += t.Profit
		if equity > peak {
			peak = equity
		}
		if peak > 0 {
			if dd := (peak - equity) / peak; dd > maxDD {
				maxDD = dd
			}
		}
	}

	if e.StartingCapital > 0 {
		m.Return = (equity - e.StartingCapital) / e.StartingCapital * 100
	}
	if len(trades) > 0 {
		m.WinRate = float64(m.Wins) / float64(len(trades)) * 100
	}
	m.MDD = maxDD * 100
	return m
}
