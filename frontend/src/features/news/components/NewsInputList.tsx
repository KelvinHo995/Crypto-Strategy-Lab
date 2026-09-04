import { useState } from 'react';
import type { NewsItem } from '../../../types/news';

interface NewsInputListProps {
  news: NewsItem[];
  selectedNewsId?: string;
  onSelectNews?: (item: NewsItem) => void;
}

export function NewsInputList({
  news,
  selectedNewsId,
  onSelectNews,
}: NewsInputListProps) {
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
        return { color: '#94a3b8', backgroundColor: 'rgba(148, 163, 184, 0.1)', borderColor: '#cbd5e1' };
    }
  };

  const detectCoinTag = (title: string, content = ''): { tag: string; color: string; bg: string } => {
    const text = (title + ' ' + content).toUpperCase();
    if (text.includes('ETH') || text.includes('ETHEREUM')) {
      return { tag: 'ETH', color: '#6366f1', bg: 'rgba(99, 102, 241, 0.15)' };
    }
    if (text.includes('SOL') || text.includes('SOLANA')) {
      return { tag: 'SOL', color: '#06b6d4', bg: 'rgba(6, 182, 212, 0.15)' };
    }
    if (text.includes('BNB') || text.includes('BINANCE')) {
      return { tag: 'BNB', color: '#eab308', bg: 'rgba(234, 179, 8, 0.15)' };
    }
    if (text.includes('XRP') || text.includes('RIPPLE')) {
      return { tag: 'XRP', color: '#0ea5e9', bg: 'rgba(14, 165, 233, 0.15)' };
    }
    if (text.includes('ADA') || text.includes('CARDANO')) {
      return { tag: 'ADA', color: '#3b82f6', bg: 'rgba(59, 130, 246, 0.15)' };
    }
    if (text.includes('AVAX') || text.includes('AVALANCHE')) {
      return { tag: 'AVAX', color: '#e11d48', bg: 'rgba(225, 29, 72, 0.15)' };
    }
    if (text.includes('DOGE') || text.includes('DOGECOIN')) {
      return { tag: 'DOGE', color: '#d97706', bg: 'rgba(217, 119, 6, 0.15)' };
    }
    return { tag: 'BTC', color: '#f59e0b', bg: 'rgba(245, 158, 11, 0.15)' };
  };

  return (
    <div style={panelContainerStyle}>
      {/* Header */}
      <div style={headerStyle}>
        <div>
          <h4 style={titleStyle}>Input Feed News</h4>
          <span style={{ fontSize: '0.68rem', color: '#64748b' }}>
            {news.length} article{news.length === 1 ? '' : 's'} in view
          </span>
        </div>
        <span style={timeStyle}>Updated: {new Date().toTimeString().split(' ')[0]}</span>
      </div>

      {/* News List */}
      <div style={listStyle}>
        {news.length === 0 ? (
          <div style={emptyStyle}>
            No articles match the selected source or asset filter.
          </div>
        ) : (
          news.map((item) => {
            const isSelected = selectedNewsId === item.id;
            const coin = detectCoinTag(item.title, item.content);

            return (
              <div
                key={item.id}
                onClick={() => onSelectNews?.(item)}
                style={isSelected ? selectedCardStyle : cardStyle}
              >
                {/* Publisher & Time */}
                <div style={metaRowStyle}>
                  <span style={sourceStyle}>{item.source}</span>
                  <span style={dateStyle}>{getFormattedTime(item.publishedAt)}</span>
                </div>

                {/* Title */}
                <h5 style={newsTitleStyle}>
                  <span
                    style={{
                      fontSize: '0.62rem',
                      fontWeight: '700',
                      backgroundColor: coin.bg,
                      color: coin.color,
                      padding: '0.1rem 0.35rem',
                      borderRadius: '4px',
                      marginRight: '0.45rem',
                      display: 'inline-block',
                      verticalAlign: 'middle',
                    }}
                  >
                    {coin.tag}
                  </span>
                  {item.title}
                </h5>

                {/* Summary / Excerpt */}
                {item.content && (
                  <p style={summaryStyle}>{item.content}</p>
                )}

                {/* Sentiment DTO Analytics Badge */}
                {item.sentiment && (
                  <div style={sentimentRowStyle}>
                    <div style={{ display: 'flex', gap: '0.35rem', alignItems: 'center' }}>
                      <span
                        style={{
                          ...sourceBadgeStyle,
                          ...(item.analysisSource === 'LIVE' ? liveSourceStyle : demoSourceStyle),
                        }}
                      >
                        {item.analysisSource === 'LIVE' ? 'LIVE' : 'DEMO'}
                      </span>
                      <span style={{ ...badgeStyle, ...getSentimentStyle(item.sentiment.sentiment) }}>
                        {item.sentiment.sentiment} ({(item.sentiment.score * 100).toFixed(0)}%)
                      </span>
                    </div>
                    <span style={modelMetaStyle} title="MLOps Traceability Model Information">
                      {item.sentiment.model.name} ({item.sentiment.model.version})
                    </span>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
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
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
  marginBottom: '0.85rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#0f172a',
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
  backgroundColor: '#f8fafc',
  border: '1px solid #e2e8f0',
  borderRadius: '6px',
  padding: '0.75rem',
  cursor: 'pointer',
  transition: 'all 0.15s ease',
};

const selectedCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(59, 130, 246, 0.05)',
  border: '1.5px solid #2563eb',
  borderRadius: '6px',
  padding: '0.75rem',
  cursor: 'pointer',
  boxShadow: '0 0 0 2px rgba(37, 99, 235, 0.1)',
  transition: 'all 0.15s ease',
};

const emptyStyle: React.CSSProperties = {
  padding: '2rem 1rem',
  textAlign: 'center',
  fontSize: '0.8rem',
  color: '#64748b',
};

const metaRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.7rem',
  marginBottom: '0.35rem',
};

const sourceStyle: React.CSSProperties = {
  color: '#2563eb',
  fontWeight: '700',
};

const dateStyle: React.CSSProperties = {
  color: '#64748b',
};

const newsTitleStyle: React.CSSProperties = {
  fontSize: '0.82rem',
  fontWeight: '700',
  color: '#0f172a',
  margin: '0 0 0.35rem 0',
  lineHeight: '1.3',
};

const summaryStyle: React.CSSProperties = {
  fontSize: '0.73rem',
  color: '#64748b',
  margin: '0 0 0.5rem 0',
  lineHeight: '1.4',
  display: '-webkit-box',
  WebkitLineClamp: 2,
  WebkitBoxOrient: 'vertical',
  overflow: 'hidden',
};

const sentimentRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderTop: '1px dashed #e2e8f0',
  paddingTop: '0.45rem',
  marginTop: '0.2rem',
};

const badgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
  border: '1px solid transparent',
};

const sourceBadgeStyle: React.CSSProperties = {
  fontSize: '0.58rem',
  fontWeight: 800,
  padding: '0.1rem 0.3rem',
  borderRadius: '4px',
};

const liveSourceStyle: React.CSSProperties = { color: '#047857', background: '#d1fae5' };
const demoSourceStyle: React.CSSProperties = { color: '#92400e', background: '#fef3c7' };

const modelMetaStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
  fontFamily: 'monospace',
};
