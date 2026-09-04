import { useEffect, useMemo, useState } from 'react';
import { TradingChart } from '../../market/components/TradingChart';
import { fetchMarketDataDTO } from '../../market/services/mockMarketData';
import { fetchCandles } from '../../../shared/api';
import { useAppMode } from '../../../shared/auth';
import { tradesToMarkers } from '../../../shared/stores/useExperimentStore';
import type { Candle } from '../../../types/candle';
import type { ExperimentResult, Trade } from '../../../types/backtest';

interface BacktestChartProps {
  experiment: ExperimentResult;
  trades: Trade[];
  highlightedTrade: Trade | null;
}

// A dedicated, static chart for one backtest's own candles + trade markers —
// deliberately NOT the same live multi-timeframe ChartCard used on the
// Market tab. That component runs up to 4 independent instances (the 2x2
// grid) each with their own live WebSocket ticks, symbol/timeframe pickers
// and reload button, all reacting to one global "loaded experiment" — every
// one of those extra moving parts was a fresh way for a stale candle range
// to silently outrace the experiment's own range and re-scatter trade
// markers across the wrong window. This chart has none of that: it fetches
// once for exactly this experiment's own pair/period, never re-fetches on
// its own, and isn't shared with anything else.
export function BacktestChart({ experiment, trades, highlightedTrade }: BacktestChartProps) {
  const mode = useAppMode();
  const [candles, setCandles] = useState<Candle[]>([]);
  const [ma20Line, setMa20Line] = useState<number[]>([]);
  const [fitNonce, setFitNonce] = useState(0);
  const [status, setStatus] = useState<'loading' | 'ready' | 'unavailable'>('loading');

  // Per spec (Trade Detail): clicking one trade highlights just its own
  // ENTRY/EXIT, not every trade in the run at once — with 100+ trades in a
  // single backtest, drawing all of them together turns into an unreadable
  // wall of arrows and dots.
  const markers = useMemo(
    () => tradesToMarkers(highlightedTrade ? [highlightedTrade] : []),
    [highlightedTrade]
  );
  const pair = trades[0]?.pair;

  useEffect(() => {
    let cancelled = false;

    void Promise.resolve().then(async () => {
      if (cancelled) return;
      setStatus('loading');

      if (mode === 'DEMO') {
        const dto = fetchMarketDataDTO(pair || 'BTCUSDT', '4h', 200);
        if (cancelled) return;
        setCandles(dto.candles);
        setMa20Line(dto.ma20Line);
        setStatus('ready');
        setFitNonce((n) => n + 1);
        return;
      }

      const [fromStr, toStr] = experiment.datasetPeriod.split('-');
      const from = Number(fromStr);
      const to = Number(toStr);
      // Backend doesn't record which timeframe a backtest actually ran on —
      // 4h is the widest available, giving the best chance a multi-month
      // range fits within the 5000-candle server cap.
      const tf = '4h';
      if (!pair || !Number.isFinite(from) || !Number.isFinite(to)) {
        setStatus('unavailable');
        return;
      }

      try {
        const apiCandles = await fetchCandles(pair, tf, from, to, 5000);
        if (cancelled) return;
        if (apiCandles.length < 20) {
          setStatus('unavailable');
          return;
        }
        const closes = apiCandles.map((c) => c.close);
        const ma = closes.map((_, i) =>
          i < 19 ? Number.NaN : closes.slice(i - 19, i + 1).reduce((a, b) => a + b, 0) / 20
        );
        setCandles(apiCandles);
        setMa20Line(ma);
        setStatus('ready');
        setFitNonce((n) => n + 1);
      } catch {
        if (!cancelled) setStatus('unavailable');
      }
    });

    return () => {
      cancelled = true;
    };
  }, [experiment.id, experiment.datasetPeriod, pair, mode]);

  if (status === 'unavailable') {
    return (
      <div style={messageStyle}>
        Historical chart unavailable for this run — no matching candle data for {pair || 'this pair'}.
      </div>
    );
  }

  return (
    <div style={chartBodyStyle}>
      {status === 'loading' ? (
        <div style={messageStyle}>Loading chart...</div>
      ) : (
        <TradingChart candles={candles} ma20Line={ma20Line} markers={markers} fitSignal={fitNonce} />
      )}
    </div>
  );
}

const chartBodyStyle: React.CSSProperties = {
  position: 'relative',
  width: '100%',
  height: '360px',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  overflow: 'hidden',
};

const messageStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  height: '100%',
  minHeight: '200px',
  color: '#64748b',
  fontSize: '0.85rem',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  backgroundColor: '#ffffff',
};
