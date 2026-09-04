import React, { useEffect, useState } from 'react';
import { fetchMarkets } from '../../../shared/api';
import { useAppMode } from '../../../shared/auth';
import { DEFAULT_MARKETS } from '../../market/services/marketCatalog';
import type { MarketInfo } from '../../../types/candle';
import type { StrategyInstance } from '../../../types/backtest';
import { useExperimentStore } from '../../../shared/stores/useExperimentStore';

interface BacktestConfigPanelProps {
  onRunBacktest: (config: {
    symbol: string;
    timeframe: string;
    fromDate: string;
    toDate: string;
    capital: number;
    fee: number;
    slippage: number;
    instances: StrategyInstance[];
    policy?: 'majority' | 'weighted';
  }) => void;
  isLoading: boolean;
}

function dateInputValue(offsetDays = 0): string {
  const date = new Date(Date.now() + offsetDays * 86400000);
  return date.toISOString().slice(0, 10);
}

const QUICK_STRATEGIES: Record<string, { label: string; instances: StrategyInstance[]; policy: 'majority' | 'weighted' }> = {
  MA: {
    label: 'MA Crossover (20/50)',
    instances: [{ type: 'MA', params: { maShortWindow: 20, maLongWindow: 50 }, weight: 1 }],
    policy: 'majority',
  },
  RSI: {
    label: 'RSI Oscillator (14, 70/30)',
    instances: [{ type: 'RSI', params: { rsiPeriod: 14, rsiOverbought: 70, rsiOversold: 30 }, weight: 1 }],
    policy: 'majority',
  },
  Bollinger: {
    label: 'Bollinger Bands (20, 2.0)',
    instances: [{ type: 'Bollinger', params: { bollingerPeriod: 20, bollingerStdDev: 2 }, weight: 1 }],
    policy: 'majority',
  },
  SR: {
    label: 'Support / Resistance Pivot (20)',
    instances: [{ type: 'SR', params: { srWindow: 20, srTolerance: 0.005 }, weight: 1 }],
    policy: 'majority',
  },
  SMC: {
    label: 'Smart Money Concepts (SMC 10)',
    instances: [{ type: 'SMC', params: { smcLookback: 10 }, weight: 1 }],
    policy: 'majority',
  },
  Sentiment: {
    label: 'News Sentiment Filter (FinBERT 0.7)',
    instances: [{ type: 'Sentiment', params: { sentimentThreshold: 0.7 }, weight: 1 }],
    policy: 'majority',
  },
};

export function BacktestConfigPanel({
  onRunBacktest,
  isLoading,
}: BacktestConfigPanelProps) {
  const mode = useAppMode();
  const builderInstances = useExperimentStore((state) => state.builderInstances);
  const builderPolicy = useExperimentStore((state) => state.builderPolicy);

  const [markets, setMarkets] = useState<MarketInfo[]>(DEFAULT_MARKETS);
  const [symbol, setSymbol] = useState('BTCUSDT');
  const [timeframe, setTimeframe] = useState('5m');
  const [fromDate, setFromDate] = useState(() => dateInputValue(-90));
  const [toDate, setToDate] = useState(() => dateInputValue());
  const [capital, setCapital] = useState(10000);
  const [fee, setFee] = useState(0.1); // 0.1% standard exchange fee
  const [slippage, setSlippage] = useState(5); // 5 bps standard slippage
  const [selectedStrategyKey, setSelectedStrategyKey] = useState<string>('builder');

  useEffect(() => {
    if (mode !== 'LIVE') return;
    fetchMarkets().then(setMarkets).catch(() => setMarkets(DEFAULT_MARKETS));
  }, [mode]);

  const builderSummary = builderInstances.map((i) => i.type).join(' + ') || 'Composite Strategy';

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    let instances: StrategyInstance[];
    let policy: 'majority' | 'weighted' = 'majority';

    if (selectedStrategyKey === 'builder') {
      instances = builderInstances.length > 0 ? builderInstances : [{ type: 'MA', params: { maShortWindow: 20, maLongWindow: 50 }, weight: 1 }];
      policy = builderPolicy;
    } else if (QUICK_STRATEGIES[selectedStrategyKey]) {
      instances = QUICK_STRATEGIES[selectedStrategyKey].instances;
      policy = QUICK_STRATEGIES[selectedStrategyKey].policy;
    } else {
      instances = [{ type: 'MA', params: { maShortWindow: 20, maLongWindow: 50 }, weight: 1 }];
    }

    onRunBacktest({
      symbol,
      timeframe,
      fromDate,
      toDate,
      capital,
      fee,
      slippage,
      instances,
      policy,
    });
  };

  return (
    <div style={panelContainerStyle}>
      <h3 style={titleStyle}>Run Simulation</h3>
      <form onSubmit={handleSubmit} style={formStyle}>
        <div style={formGridStyle}>
          {/* Strategy / Preset Selector */}
          <div style={{ ...formGroupStyle, gridColumn: '1 / -1' }}>
            <label style={labelStyle}>Strategy Configuration</label>
            <select
              value={selectedStrategyKey}
              onChange={(e) => setSelectedStrategyKey(e.target.value)}
              style={{ ...selectStyle, fontWeight: '600' }}
              disabled={isLoading}
            >
              <option value="builder">★ Tổ hợp từ Strategy Builder ({builderSummary})</option>
              <optgroup label="Single Indicator Strategies">
                {Object.entries(QUICK_STRATEGIES).map(([key, item]) => (
                  <option key={key} value={key}>
                    {item.label}
                  </option>
                ))}
              </optgroup>
            </select>
          </div>

          {/* Pair selector */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>Trading Pair</label>
            <select
              value={symbol}
              onChange={(e) => setSymbol(e.target.value)}
              style={selectStyle}
              disabled={isLoading}
            >
              {markets.map(market => <option key={market.symbol} value={market.symbol}>{market.baseAsset}/{market.quoteAsset}</option>)}
            </select>
          </div>

          {/* Timeframe */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>Timeframe</label>
            <select
              value={timeframe}
              onChange={(e) => setTimeframe(e.target.value)}
              style={selectStyle}
              disabled={isLoading}
            >
              <option value="5m">5m</option>
              <option value="15m">15m</option>
              <option value="1h">1h</option>
              <option value="4h">4h</option>
            </select>
          </div>

          {/* Date range - From */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>From Date</label>
            <input
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
              style={inputStyle}
              disabled={isLoading}
              required
            />
          </div>

          {/* Date range - To */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>To Date</label>
            <input
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
              style={inputStyle}
              disabled={isLoading}
              required
            />
          </div>

          {/* Capital */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>Starting Capital (USD)</label>
            <input
              type="number"
              value={capital}
              min="100"
              max="10000000"
              onChange={(e) => setCapital(parseFloat(e.target.value))}
              style={inputStyle}
              disabled={isLoading}
              required
            />
          </div>

          {/* Trading fee */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>Maker/Taker Fee (%)</label>
            <input
              type="number"
              value={fee}
              min="0"
              max="2"
              step="0.01"
              onChange={(e) => setFee(parseFloat(e.target.value) || 0)}
              style={inputStyle}
              disabled={isLoading}
              required
            />
          </div>

          {/* Slippage */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>Slippage (bps)</label>
            <input
              type="number"
              value={slippage}
              min="0"
              max="200"
              step="1"
              onChange={(e) => setSlippage(parseFloat(e.target.value) || 0)}
              style={inputStyle}
              disabled={isLoading}
              required
            />
          </div>
        </div>

        {/* Submit button */}
        <button type="submit" disabled={isLoading} style={isLoading ? activeSubmitBtnStyle : submitBtnStyle}>
          {isLoading ? (
            <span style={spinnerContainerStyle}>
              <span style={spinnerStyle} /> Processing Backtest...
            </span>
          ) : (
            'Kích hoạt Backtest'
          )}
        </button>
      </form>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const panelContainerStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1.25rem',
  boxSizing: 'border-box',
  height: '100%',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#94a3b8',
  margin: '0 0 1rem 0',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
};

const formStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1.25rem',
};

const formGridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
  gap: '1rem',
};

const formGroupStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.35rem',
};

const labelStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  fontWeight: '600',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.5rem',
  fontSize: '0.8rem',
  outline: 'none',
  cursor: 'pointer',
};

const inputStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.45rem',
  fontSize: '0.8rem',
  outline: 'none',
};

const submitBtnStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6', // Blue-500
  color: '#ffffff',
  border: 'none',
  borderRadius: '6px',
  padding: '0.65rem',
  fontSize: '0.85rem',
  fontWeight: '700',
  cursor: 'pointer',
  transition: 'all 0.2s',
  boxShadow: '0 2px 6px rgba(59, 130, 246, 0.2)',
};

const activeSubmitBtnStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#475569',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.65rem',
  fontSize: '0.85rem',
  fontWeight: '700',
  cursor: 'not-allowed',
};

const spinnerContainerStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  gap: '0.5rem',
};

const spinnerStyle: React.CSSProperties = {
  width: '12px',
  height: '12px',
  border: '2px solid rgba(255,255,255,0.2)',
  borderTop: '2px solid #3b82f6',
  borderRadius: '50%',
  animation: 'spin 1s linear infinite',
  display: 'inline-block',
};
