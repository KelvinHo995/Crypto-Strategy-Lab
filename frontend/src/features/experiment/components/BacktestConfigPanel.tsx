import React, { useState } from 'react';

interface BacktestConfigPanelProps {
  onRunBacktest: (config: {
    symbol: string;
    timeframe: string;
    fromDate: string;
    toDate: string;
    capital: number;
    fee: number;
  }) => void;
  isLoading: boolean;
}

export function BacktestConfigPanel({
  onRunBacktest,
  isLoading,
}: BacktestConfigPanelProps) {
  const [symbol, setSymbol] = useState('BTCUSDT');
  const [timeframe, setTimeframe] = useState('5m');
  const [fromDate, setFromDate] = useState('2024-01-01');
  const [toDate, setToDate] = useState('2024-12-31');
  const [capital, setCapital] = useState(10000);
  const [fee, setFee] = useState(0.1); // 0.1% standard exchange fee

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onRunBacktest({
      symbol,
      timeframe,
      fromDate,
      toDate,
      capital,
      fee,
    });
  };

  return (
    <div style={panelContainerStyle}>
      <h3 style={titleStyle}>Run Simulation</h3>
      <form onSubmit={handleSubmit} style={formStyle}>
        <div style={formGridStyle}>
          {/* Pair selector */}
          <div style={formGroupStyle}>
            <label style={labelStyle}>Trading Pair</label>
            <select
              value={symbol}
              onChange={(e) => setSymbol(e.target.value)}
              style={selectStyle}
              disabled={isLoading}
            >
              <option value="BTCUSDT">BTC/USDT</option>
              <option value="ETHUSDT">ETH/USDT</option>
              <option value="SOLUSDT">SOL/USDT</option>
              <option value="BNBUSDT">BNB/USDT</option>
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
              <option value="1m">1m</option>
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
              onChange={(e) => setFee(parseFloat(e.target.value))}
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
            '▷ Kích hoạt Backtest'
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
  backgroundColor: '#0f172a',
  border: '1px solid #1e293b',
  borderRadius: '8px',
  padding: '1.25rem',
  boxSizing: 'border-box',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#94a3b8',
  margin: '0 0 1rem 0',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #1e293b',
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
  backgroundColor: '#1e293b',
  color: '#ffffff',
  border: '1px solid #334155',
  borderRadius: '6px',
  padding: '0.5rem',
  fontSize: '0.8rem',
  outline: 'none',
  cursor: 'pointer',
};

const inputStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  color: '#ffffff',
  border: '1px solid #334155',
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
  backgroundColor: '#1e293b',
  color: '#475569',
  border: '1px solid #334155',
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
