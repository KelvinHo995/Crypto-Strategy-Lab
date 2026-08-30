import { useState } from 'react';
import type { NewsItem } from '../../../types/news';

interface NewsInputListProps {
  news: NewsItem[];
}

export function NewsInputList({ news }: NewsInputListProps) {
  const [now] = useState(() => Date.now());
  const getFormattedTime = (timestamp: number): string => {
    const diffMins = Math.floor((now - timestamp) / (60 * 1000));
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    return new Date(timestamp).toLocaleDateString();
  };

  const getSentimentStyle = (sentiment: string) => {
    switch (sentiment) {
      case 'POSITIVE':
        return { color: '#10b981', backgroundColor: 'rgba(16, 185, 129, 0.1)', borderColor: '#10b981' };
      case 'NEGATIVE':
        return { color: '#ef4444', backgroundColor: 'rgba(239, 68, 68, 0.1)', borderColor: '#ef4444' };
      default:
        return { color: '#94a3b8', backgroundColor: 'rgba(148, 163, 184, 0.1)', borderColor: '#475569' };
    }
  };

  const getAssetBadgeStyle = (title: string): React.CSSProperties => {
    let bgColor = 'rgba(245, 158, 11, 0.15)';
    let color = '#f59e0b';

    if (title.toUpperCase().includes('ETH') || title.toUpperCase().includes('ETHEREUM')) {
      bgColor = 'rgba(99, 102, 241, 0.15)';
      color = '#6366f1';
    } else if (title.toUpperCase().includes('SOL') || title.toUpperCase().includes('SOLANA')) {
      bgColor = 'rgba(6, 182, 212, 0.15)';
      color = '#06b6d4';
    } else if (title.toUpperCase().includes('BNB')) {
      bgColor = 'rgba(234, 179, 8, 0.15)';
      color = '#eab308';
    }

    return {
      fontSize: '0.6rem',
      fontWeight: '700',
      backgroundColor: bgColor,
      color,
      padding: '0.1rem 0.3rem',
      borderRadius: '4px',
      marginRight: '0.5rem',
      display: 'inline-block',
      verticalAlign: 'middle',
    };
  };

  return (
    <div style={panelContainerStyle}>
      {/* Header */}
      <div style={headerStyle}>
        <h4 style={titleStyle}>Input Feed News</h4>
        <span style={timeStyle}>Updated: {new Date().toTimeString().split(' ')[0]}</span>
      </div>

      {/* News List */}
      <div style={listStyle}>
        {news.map((item) => (
          <div key={item.id} style={cardStyle}>
            {/* Publisher & Time */}
            <div style={metaRowStyle}>
              <span style={sourceStyle}>{item.source}</span>
              <span style={dateStyle}>{getFormattedTime(item.publishedAt)}</span>
            </div>

            {/* Title */}
            <h5 style={newsTitleStyle}>
              <span style={getAssetBadgeStyle(item.title)}>
                {item.title.toUpperCase().includes('ETH') ? 'ETH' : item.title.toUpperCase().includes('SOL') ? 'SOL' : 'BTC'}
              </span>
              {item.title}
            </h5>

            {/* Summary */}
            <p style={summaryStyle}>{item.content}</p>

            {/* Sentiment DTO Analytics Badge */}
            {item.sentiment && (
              <div style={sentimentRowStyle}>
                <span style={{ ...badgeStyle, ...getSentimentStyle(item.sentiment.sentiment) }}>
                  {item.sentiment.sentiment} ({(item.sentiment.score * 100).toFixed(0)}%)
                </span>
                <span style={modelMetaStyle} title="MLOps Traceability Model Information">
                  🤖 {item.sentiment.model.name} ({item.sentiment.model.version})
                </span>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Footer Link */}
      <button style={viewAllStyle}>
        View All Market News →
      </button>
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
  padding: '1rem',
  height: '100%',
  boxSizing: 'border-box',
  display: 'flex',
  flexDirection: 'column',
};

const headerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.5rem',
  marginBottom: '1rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#cbd5e1',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const timeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
};

const listStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
  overflowY: 'auto',
  maxHeight: 'calc(100vh - 280px)',
  paddingRight: '4px',
  flexGrow: 1,
};

const cardStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  border: '1px solid #334155',
  borderRadius: '6px',
  padding: '0.75rem',
};

const metaRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.7rem',
  marginBottom: '0.35rem',
};

const sourceStyle: React.CSSProperties = {
  color: '#3b82f6',
  fontWeight: '700',
};

const dateStyle: React.CSSProperties = {
  color: '#64748b',
};

const newsTitleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#f8fafc',
  margin: '0 0 0.4rem 0',
  lineHeight: '1.3',
};

const summaryStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#94a3b8',
  margin: '0 0 0.6rem 0',
  lineHeight: '1.4',
};

const sentimentRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderTop: '1px dashed #334155',
  paddingTop: '0.5rem',
};

const badgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
  border: '1px solid transparent',
};

const modelMetaStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
  fontFamily: 'monospace',
};

const viewAllStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#06b6d4',
  cursor: 'pointer',
  fontSize: '0.8rem',
  fontWeight: '600',
  textAlign: 'center',
  paddingTop: '1rem',
  borderTop: '1px solid #1e293b',
  marginTop: '0.5rem',
  width: '100%',
};
