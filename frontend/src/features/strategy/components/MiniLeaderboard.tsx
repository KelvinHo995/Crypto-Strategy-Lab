import type { MiniLeaderboardItem } from '../services/mockStrategyData';

interface MiniLeaderboardProps {
  items: MiniLeaderboardItem[];
}

export function MiniLeaderboard({ items }: MiniLeaderboardProps) {
  return (
    <div style={containerStyle}>
      <h4 style={titleStyle}>Top-5 Discovered Candidates</h4>
      <table style={tableStyle}>
        <thead>
          <tr>
            <th style={thStyle}>Rank</th>
            <th style={thLeftStyle}>Strategy Combination</th>
            <th style={thRightStyle}>Profit USDT</th>
            <th style={thRightStyle}>Winrate</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.rank} style={item.rank === 1 ? topRowStyle : trStyle}>
              <td style={tdRankStyle(item.rank)}>{item.rank}</td>
              <td style={tdNameStyle}>{item.name}</td>
              <td style={tdProfitStyle}>{item.profit.toLocaleString()} USDT</td>
              <td style={tdRightStyle}>{item.winrate}%</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const containerStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  border: '1px solid #1e293b',
  borderRadius: '8px',
  padding: '1rem',
  boxSizing: 'border-box',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#94a3b8',
  margin: '0 0 0.75rem 0',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const tableStyle: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  fontSize: '0.75rem',
};

const trStyle: React.CSSProperties = {
  borderBottom: '1px solid #1e293b',
};

const topRowStyle: React.CSSProperties = {
  borderBottom: '1px solid #1e293b',
  backgroundColor: 'rgba(6, 182, 212, 0.03)', // Light cyan background for #1
};

const thStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'center',
  padding: '0.35rem',
  fontWeight: '600',
  borderBottom: '1px solid #1e293b',
};

const thLeftStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'left',
  padding: '0.35rem',
  fontWeight: '600',
  borderBottom: '1px solid #1e293b',
};

const thRightStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'right',
  padding: '0.35rem',
  fontWeight: '600',
  borderBottom: '1px solid #1e293b',
};

const tdRankStyle = (rank: number): React.CSSProperties => {
  let color = '#94a3b8';
  let fontWeight = '500';

  if (rank === 1) {
    color = '#f59e0b'; // Gold
    fontWeight = '700';
  } else if (rank === 2) {
    color = '#cbd5e1'; // Silver
    fontWeight = '700';
  } else if (rank === 3) {
    color = '#b45309'; // Bronze
    fontWeight = '700';
  }

  return {
    color,
    fontWeight,
    textAlign: 'center',
    padding: '0.5rem 0.35rem',
    width: '35px',
  };
};

const tdNameStyle: React.CSSProperties = {
  color: '#e2e8f0',
  fontWeight: '500',
  padding: '0.5rem 0.35rem',
};

const tdProfitStyle: React.CSSProperties = {
  color: '#10b981', // Green
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.5rem 0.35rem',
};

const tdRightStyle: React.CSSProperties = {
  color: '#cbd5e1',
  textAlign: 'right',
  padding: '0.5rem 0.35rem',
};
