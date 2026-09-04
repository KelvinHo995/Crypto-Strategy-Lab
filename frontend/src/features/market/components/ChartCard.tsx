import { useState, useEffect, useCallback } from 'react';
import { useWebSocketState, useWebSocketSubscription } from '../../../shared/hooks';
import type { Candle, MarketInfo } from '../../../types/candle';
import { TradingChart, type SRZone, type ChartMarker } from './TradingChart';
import { fetchMarketDataDTO, generateNextTick } from '../services/mockMarketData';
import { fetchCandles } from '../../../shared/api';
import { wsManager } from '../../../shared/ws';
import { useAppMode } from '../../../shared/auth';
import { useExperimentStore, formatExperimentTitle } from '../../../shared/stores/useExperimentStore';

interface ChartCardProps {
  id: number;
  defaultSymbol: string;
  defaultTimeframe: string;
  isMaximized: boolean;
  onToggleMaximize: (id: number) => void;
  markets: MarketInfo[];
}

export function ChartCard({
  id,
  defaultSymbol,
  defaultTimeframe,
  isMaximized,
  onToggleMaximize,
  markets,
}: ChartCardProps) {
  const mode = useAppMode();
  const wsState = useWebSocketState();
  const [symbol, setSymbol] = useState(defaultSymbol);
  const [timeframe, setTimeframe] = useState(defaultTimeframe);
  const [candles, setCandles] = useState<Candle[]>([]);
  const [ma20Line, setMa20Line] = useState<number[]>([]);
  const [bbands, setBbands] = useState<{ upper: number[]; basis: number[]; lower: number[] } | undefined>(undefined);
  const [srZones, setSrZones] = useState<SRZone[]>([]);
  const [markers, setMarkers] = useState<ChartMarker[]>([]);
  const [priceTrend, setPriceTrend] = useState<'UP' | 'DOWN' | 'NEUTRAL'>('NEUTRAL');
  const [dataMode, setDataMode] = useState<'API' | 'MOCK' | 'ERROR'>(mode === 'DEMO' ? 'MOCK' : 'API');
  const [loadError, setLoadError] = useState('');

  // 1. Fetch initial historical data on mount or when symbol/timeframe changes
  const loadHistory = useCallback(async () => {
    if (mode === 'DEMO') {
      const dto = fetchMarketDataDTO(symbol, timeframe, 200);
      setCandles(dto.candles);
      setMa20Line(dto.ma20Line);
      setBbands(dto.bbands);
      setSrZones(dto.srZones);
      setMarkers(dto.markers);
      setDataMode('MOCK');
      setLoadError('');
      return;
    }
    const to = Date.now();
    const from = to - 366 * 24 * 60 * 60 * 1000;
    try {
      const apiCandles = await fetchCandles(symbol, timeframe, from, to, 500);
      if (apiCandles.length < 20) throw new Error('insufficient candles');
      const closes = apiCandles.map(c => c.close);
      const period = 20;
      const multiplier = 2;
      const ma: number[] = [];
      const upper: number[] = [];
      const basis: number[] = [];
      const lower: number[] = [];

      for (let i = 0; i < closes.length; i++) {
        if (i < period - 1) {
          ma.push(Number.NaN);
          upper.push(Number.NaN);
          basis.push(Number.NaN);
          lower.push(Number.NaN);
          continue;
        }
        let sum = 0;
        for (let j = 0; j < period; j++) {
          sum += closes[i - j];
        }
        const avg = sum / period;
        ma.push(avg);
        basis.push(avg);

        let varianceSum = 0;
        for (let j = 0; j < period; j++) {
          varianceSum += Math.pow(closes[i - j] - avg, 2);
        }
        const stdDev = Math.sqrt(varianceSum / period);
        upper.push(avg + multiplier * stdDev);
        lower.push(avg - multiplier * stdDev);
      }

      setCandles(apiCandles);
      setMa20Line(ma);
      setBbands({ upper, basis, lower });
      setSrZones([]);
      setMarkers([]);
      setDataMode('API');
      setLoadError('');
    } catch (error) {
      setCandles([]);
      setMa20Line([]);
      setBbands(undefined);
      setSrZones([]);
      setMarkers([]);
      setDataMode('ERROR');
      setLoadError(`Không có historical data cho ${symbol}/${timeframe}: ${String(error)}`);
    }
  }, [mode, symbol, timeframe]);

  useEffect(() => {
    void Promise.resolve().then(loadHistory);
  }, [loadHistory]);

  function handleRealtimeUpdate(candle: Candle) {
    setCandles((prev) => {
      if (prev.length === 0) return [candle];
      const last = prev[prev.length - 1];

      if (candle.close > last.close) setPriceTrend('UP');
      else if (candle.close < last.close) setPriceTrend('DOWN');

      if (candle.openTime === last.openTime) {
        return [...prev.slice(0, -1), candle];
      } else if (candle.openTime > last.openTime) {
        return [...prev, candle];
      }
      return prev;
    });
  }



  // 2. Realtime WebSocket Subscription — only once real history has loaded,
  // so a stray tick can't seed the chart with a single candle (autoScale
  // then zooms to fit that one point, which looks broken) while history is
  // still loading or failed to load.
  useWebSocketSubscription('CANDLE_UPDATE', (candle: Candle) => {
    if (dataMode === 'API' && candle.symbol === symbol && candle.timeframe === timeframe) {
      handleRealtimeUpdate(candle);
    }
  });

  useEffect(() => {
    if (mode !== 'LIVE') return;
    wsManager.subscribeCandles(symbol, timeframe);
    return () => wsManager.unsubscribeCandles(symbol, timeframe);
  }, [mode, symbol, timeframe]);

  // 3. Fallback Mock tick interval (Runs ONLY if WebSocket is not connected)
  useEffect(() => {
    if (mode !== 'DEMO') return;

    const interval = setInterval(() => {
      setCandles((prev) => {
        if (prev.length === 0) return prev;
        const last = prev[prev.length - 1];
        const nextCandle = generateNextTick(last);

        // Detect price trend
        if (nextCandle.close > last.close) setPriceTrend('UP');
        else if (nextCandle.close < last.close) setPriceTrend('DOWN');

        const newCandles = [...prev];
        if (nextCandle.openTime === last.openTime) {
          // Update tick
          newCandles[newCandles.length - 1] = nextCandle;
        } else {
          // Append new candle
          newCandles.push(nextCandle);
        }

        // Recalculate indicators reactively for mock ticks smoothly
        // This simulates receiving refreshed indicator overlays from the server
        const closes = newCandles.map((c) => c.close);
        const period = 20;

        // Simple MA
        const ma: number[] = [];
        const upper: number[] = [];
        const basis: number[] = [];
        const lower: number[] = [];

        for (let i = 0; i < closes.length; i++) {
          if (i < period - 1) {
            ma.push(NaN);
            upper.push(NaN);
            basis.push(NaN);
            lower.push(NaN);
            continue;
          }
          let sum = 0;
          for (let j = 0; j < period; j++) {
            sum += closes[i - j];
          }
          const avg = sum / period;
          ma.push(avg);
          basis.push(avg);

          // Standard deviation for BBands
          let varianceSum = 0;
          for (let j = 0; j < period; j++) {
            varianceSum += Math.pow(closes[i - j] - avg, 2);
          }
          const stdDev = Math.sqrt(varianceSum / period);
          upper.push(avg + 2 * stdDev);
          lower.push(avg - 2 * stdDev);
        }

        setMa20Line(ma);
        setBbands({ upper, basis, lower });

        return newCandles;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [mode, symbol, timeframe]);



  const activeExperiment = useExperimentStore((s) => s.activeExperiment);
  const globalMarkers = useExperimentStore((s) => s.activeMarkers);

  // Get properties for header
  const latestCandle = candles[candles.length - 1];
  const currentPrice = latestCandle ? latestCandle.close : 0;

  // Use global markers from loaded experiment if available, otherwise fallback to local mock markers
  const effectiveMarkers = globalMarkers.length > 0 ? globalMarkers : markers;

  // Find last trade signal
  const lastSignal = [...effectiveMarkers]
    .reverse()
    .find(m => m.text.includes('BUY') || m.text.includes('SELL'));

  return (
    <div style={cardContainerStyle}>
      {/* Chart Header */}
      <div style={cardHeaderStyle}>
        <div style={leftHeaderStyle}>
          {/* Symbol Selector */}
          <select
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            style={selectStyle}
          >
            {markets.map(market => <option key={market.symbol} value={market.symbol}>{market.baseAsset}/{market.quoteAsset}</option>)}
          </select>

          {/* Timeframe Selector */}
          <div style={tabsStyle}>
            {['5m', '15m', '1h', '4h'].map((tf) => (
              <button
                key={tf}
                onClick={() => setTimeframe(tf)}
                style={timeframe === tf ? activeTabBtnStyle : tabBtnStyle}
              >
                {tf}
              </button>
            ))}
          </div>
        </div>

        {/* Live Info & Control Buttons */}
        <div style={rightHeaderStyle}>
          {/* Active Strategy Loaded Badge */}
          {activeExperiment && (
            <span style={strategyLoadedBadgeStyle} title={`Loaded strategy #${activeExperiment.id} (${activeExperiment.policy})`}>
              Strategy: {formatExperimentTitle(activeExperiment)}
            </span>
          )}

          {/* Last Signal Badge */}
          {lastSignal && (
            <span style={lastSignal.text.startsWith('BUY') ? buyBadgeStyle : sellBadgeStyle}>
              Last: {lastSignal.text}
            </span>
          )}

          {/* Price display with direction color */}
          <div style={getPriceStyle(priceTrend)}>
            {currentPrice ? currentPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : 'Loading...'}
          </div>

          {/* Live vs Mock indicator dot */}
          <div style={indicatorAreaStyle} title={mode === 'DEMO' ? 'Offline demo data' : wsState === 'CONNECTED' ? 'Binance realtime stream' : 'Historical API only'}>
            <span style={mode === 'DEMO' ? mockDotStyle : wsState === 'CONNECTED' ? liveDotStyle : errorDotStyle} />
            <span style={indicatorTextStyle}>{mode === 'DEMO' ? 'DEMO' : wsState === 'CONNECTED' ? `LIVE/${dataMode}` : dataMode === 'API' ? 'API/OFFLINE' : 'ERROR'}</span>
          </div>

          {/* Maximize and reload action button */}
          <button onClick={loadHistory} style={actionBtnStyle} title="Reload historical data">
            ↻
          </button>
          <button onClick={() => onToggleMaximize(id)} style={actionBtnStyle}>
            {isMaximized ? '−' : '□'}
          </button>
        </div>
      </div>

      {/* Chart Canvas Panel */}
      <div style={chartBodyStyle}>
        {candles.length > 0 ? (
          <TradingChart
            candles={candles}
            ma20Line={ma20Line}
            bbands={bbands}
            srZones={srZones}
            markers={effectiveMarkers}
          />
        ) : (
          <div style={loadingStyle}>{loadError || 'Loading historical data...'}</div>
        )}
      </div>
    </div>
  );
}

// ==========================================
// CSS STYLING
// ==========================================
const cardContainerStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  display: 'flex',
  flexDirection: 'column',
  height: '100%',
  width: '100%',
  overflow: 'hidden',
  boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)',
};

const cardHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '0.5rem 1rem',
  backgroundColor: '#ffffff',
  borderBottom: '1px solid #e2e8f0',
  flexWrap: 'wrap',
  gap: '0.5rem',
};

const leftHeaderStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.75rem',
};

const rightHeaderStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.75rem',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.25rem 0.5rem',
  fontWeight: '600',
  fontSize: '0.85rem',
  cursor: 'pointer',
  outline: 'none',
};

const tabsStyle: React.CSSProperties = {
  display: 'flex',
  backgroundColor: '#ffffff',
  borderRadius: '4px',
  padding: '2px',
  border: '1px solid #cbd5e1',
};

const tabBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: 'none',
  padding: '0.2rem 0.4rem',
  borderRadius: '3px',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'pointer',
  transition: 'all 0.1s',
};

const activeTabBtnStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6',
  color: '#ffffff',
  border: 'none',
  padding: '0.2rem 0.4rem',
  borderRadius: '3px',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'pointer',
};

const getPriceStyle = (trend: 'UP' | 'DOWN' | 'NEUTRAL'): React.CSSProperties => {
  let color = '#0f172a';
  if (trend === 'UP') color = '#10b981'; // Green
  else if (trend === 'DOWN') color = '#ef4444'; // Red

  return {
    color,
    fontWeight: '700',
    fontSize: '1rem',
    minWidth: '90px',
    textAlign: 'right',
    transition: 'color 0.15s ease',
  };
};

const strategyLoadedBadgeStyle: React.CSSProperties = {
  backgroundColor: 'rgba(37, 99, 235, 0.12)',
  border: '1px solid #2563eb',
  color: '#2563eb',
  fontSize: '0.7rem',
  fontWeight: '700',
  padding: '0.15rem 0.45rem',
  borderRadius: '4px',
};

const buyBadgeStyle: React.CSSProperties = {
  backgroundColor: 'rgba(16, 185, 129, 0.15)',
  border: '1px solid #10b981',
  color: '#10b981',
  fontSize: '0.7rem',
  fontWeight: '700',
  padding: '0.15rem 0.4rem',
  borderRadius: '4px',
};

const sellBadgeStyle: React.CSSProperties = {
  backgroundColor: 'rgba(239, 68, 68, 0.15)',
  border: '1px solid #ef4444',
  color: '#ef4444',
  fontSize: '0.7rem',
  fontWeight: '700',
  padding: '0.15rem 0.4rem',
  borderRadius: '4px',
};

const indicatorAreaStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.25rem',
  backgroundColor: '#ffffff',
  padding: '0.15rem 0.4rem',
  borderRadius: '4px',
  border: '1px solid #cbd5e1',
};

const liveDotStyle: React.CSSProperties = {
  width: '6px',
  height: '6px',
  backgroundColor: '#10b981',
  borderRadius: '50%',
  display: 'inline-block',
  boxShadow: '0 0 6px #10b981',
};

const mockDotStyle: React.CSSProperties = {
  width: '6px',
  height: '6px',
  backgroundColor: '#f59e0b',
  borderRadius: '50%',
  display: 'inline-block',
  boxShadow: '0 0 6px #f59e0b',
};

const errorDotStyle: React.CSSProperties = {
  ...mockDotStyle,
  backgroundColor: '#ef4444',
  boxShadow: '0 0 6px #ef4444',
};

const indicatorTextStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  color: '#94a3b8',
};

const actionBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#94a3b8',
  cursor: 'pointer',
  fontSize: '0.9rem',
  padding: '0.25rem',
  borderRadius: '4px',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  transition: 'background-color 0.2s',
  outline: 'none',
  width: '24px',
  height: '24px',
};

const chartBodyStyle: React.CSSProperties = {
  flexGrow: 1,
  position: 'relative',
  minHeight: '260px',
  backgroundColor: '#ffffff',
};

const loadingStyle: React.CSSProperties = {
  position: 'absolute',
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  color: '#64748b',
  fontSize: '0.9rem',
};
