interface PerformanceSummaryCardProps {
  profit: number;
  returnPct: number;
  winRate: number;
  mdd: number;
  tradeCount: number;
  wins: number;
  losses: number;
}

export function PerformanceSummaryCard({
  profit,
  returnPct,
  winRate,
  mdd,
  tradeCount,
  wins,
  losses,
}: PerformanceSummaryCardProps) {
  const isProfit = returnPct >= 0;

  return (
    <div style={containerStyle}>
      {/* 1. Net Profit Card */}
      <div style={cardStyle}>
        <span style={labelStyle}>Net Profit / Return</span>
        <span style={isProfit ? valueProfitStyle : valueLossStyle}>
          {isProfit ? `+$${profit.toFixed(2)}` : `-$${Math.abs(profit).toFixed(2)}`}
        </span>
        <span style={isProfit ? badgeProfitStyle : badgeLossStyle}>
          {isProfit ? `+${returnPct.toFixed(2)}%` : `${returnPct.toFixed(2)}%`}
        </span>
      </div>

      {/* 2. Win Rate Card */}
      <div style={cardStyle}>
        <span style={labelStyle}>Win Rate</span>
        <span style={{ ...valueStyle, color: '#0f172a' }}>
          {winRate.toFixed(2)}%
        </span>
        <span style={subLabelStyle}>
          {wins} Wins / {losses} Losses
        </span>
      </div>

      {/* 3. Max Drawdown Card */}
      <div style={cardStyle}>
        <span style={labelStyle}>Max Drawdown (MDD)</span>
        <span style={mdd >= 20 ? valueWarnStyle : valueStyle}>
          {mdd.toFixed(2)}%
        </span>
        <span style={subLabelStyle}>
          {mdd >= 20 ? 'High Risk Exposure' : 'Normal Risk Level'}
        </span>
      </div>

      {/* 4. Total Trades Card */}
      <div style={cardStyle}>
        <span style={labelStyle}>Total Trades Executed</span>
        <span style={{ ...valueStyle, color: '#2563eb' }}>
          {tradeCount}
        </span>
        <span style={subLabelStyle}>
          Average 1.6 trades/day
        </span>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const containerStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
  gap: '1rem',
  width: '100%',
};

const cardStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.25rem',
  boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
};

const labelStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  fontWeight: '700',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const valueStyle: React.CSSProperties = {
  fontSize: '1.4rem',
  fontWeight: '800',
  color: '#cbd5e1',
};

const valueProfitStyle: React.CSSProperties = {
  fontSize: '1.4rem',
  fontWeight: '800',
  color: '#10b981', // Green
};

const valueLossStyle: React.CSSProperties = {
  fontSize: '1.4rem',
  fontWeight: '800',
  color: '#ef4444', // Red
};

const valueWarnStyle: React.CSSProperties = {
  fontSize: '1.4rem',
  fontWeight: '800',
  color: '#f59e0b', // Yellow
};

const badgeProfitStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  fontWeight: '700',
  color: '#10b981',
  backgroundColor: 'rgba(16, 185, 129, 0.1)',
  padding: '0.1rem 0.3rem',
  borderRadius: '4px',
  alignSelf: 'flex-start',
  marginTop: '0.25rem',
};

const badgeLossStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  fontWeight: '700',
  color: '#ef4444',
  backgroundColor: 'rgba(239, 68, 68, 0.1)',
  padding: '0.1rem 0.3rem',
  borderRadius: '4px',
  alignSelf: 'flex-start',
  marginTop: '0.25rem',
};

const subLabelStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  marginTop: '0.25rem',
};
