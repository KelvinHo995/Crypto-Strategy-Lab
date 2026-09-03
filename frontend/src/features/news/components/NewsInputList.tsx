import { useState } from 'react';
import type { NewsItem } from '../../../types/news';

interface NewsInputListProps {
  news: NewsItem[];
  selectedArticleId?: string | null;
  onSelectArticle?: (id: string) => void;
}

export function NewsInputList({
  news,
  selectedArticleId,
  onSelectArticle,
}: NewsInputListProps) {
  const [now] = useState(() => Date.now());
  const [isViewAllModalOpen, setIsViewAllModalOpen] = useState(false);
  const [modalArticle, setModalArticle] = useState<NewsItem | null>(null);

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

  const getAssetInfo = (title: string, content = ''): { label: string; style: React.CSSProperties } => {
    const text = `${title} ${content}`.toUpperCase();
    let label = 'BTC';
    let bgColor = 'rgba(245, 158, 11, 0.15)';
    let color = '#f59e0b';

    if (text.includes('ETH') || text.includes('ETHEREUM')) {
      label = 'ETH';
      bgColor = 'rgba(99, 102, 241, 0.15)';
      color = '#6366f1';
    } else if (text.includes('SOL') || text.includes('SOLANA')) {
      label = 'SOL';
      bgColor = 'rgba(6, 182, 212, 0.15)';
      color = '#2563eb';
    } else if (text.includes('BNB') || text.includes('BINANCE')) {
      label = 'BNB';
      bgColor = 'rgba(234, 179, 8, 0.15)';
      color = '#eab308';
    } else if (text.includes('XRP') || text.includes('RIPPLE')) {
      label = 'XRP';
      bgColor = 'rgba(14, 165, 233, 0.15)';
      color = '#0ea5e9';
    } else if (text.includes('ADA') || text.includes('CARDANO')) {
      label = 'ADA';
      bgColor = 'rgba(59, 130, 246, 0.15)';
      color = '#3b82f6';
    } else if (text.includes('DOGE')) {
      label = 'DOGE';
      bgColor = 'rgba(217, 119, 6, 0.15)';
      color = '#d97706';
    }

    return {
      label,
      style: {
        fontSize: '0.6rem',
        fontWeight: '700',
        backgroundColor: bgColor,
        color,
        padding: '0.1rem 0.35rem',
        borderRadius: '4px',
        marginRight: '0.45rem',
        display: 'inline-block',
        verticalAlign: 'middle',
      },
    };
  };

  const handleCardClick = (item: NewsItem) => {
    onSelectArticle?.(item.id);
  };

  const handleCardDoubleClick = (item: NewsItem) => {
    setModalArticle(item);
    setIsViewAllModalOpen(true);
  };

  return (
    <div style={panelContainerStyle}>
      {/* Header */}
      <div style={headerStyle}>
        <h4 style={titleStyle}>Input Feed News ({news.length})</h4>
        <span style={timeStyle}>Updated: {new Date().toTimeString().split(' ')[0]}</span>
      </div>

      {/* News List */}
      <div style={listStyle}>
        {news.length === 0 ? (
          <div style={emptyStyle}>Không có tin bài nào khớp với bộ lọc đã chọn.</div>
        ) : (
          news.map((item) => {
            const isSelected = item.id === selectedArticleId;
            const assetInfo = getAssetInfo(item.title, item.content);
            return (
              <div
                key={item.id}
                onClick={() => handleCardClick(item)}
                onDoubleClick={() => handleCardDoubleClick(item)}
                style={isSelected ? activeCardStyle : cardStyle}
                title="Click để trích xuất sang Cột 2 (Double click để xem toàn văn)"
              >
                {/* Publisher & Time */}
                <div style={metaRowStyle}>
                  <span style={sourceStyle}>{item.source}</span>
                  <span style={dateStyle}>{getFormattedTime(item.publishedAt)}</span>
                </div>

                {/* Title */}
                <h5 style={newsTitleStyle}>
                  <span style={assetInfo.style}>{assetInfo.label}</span>
                  {item.title}
                </h5>

                {/* Summary */}
                <p style={summaryStyle}>{item.content}</p>

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

      {/* Footer Link */}
      <button
        type="button"
        onClick={() => {
          setModalArticle(null);
          setIsViewAllModalOpen(true);
        }}
        style={viewAllStyle}
      >
        View All Market News ({news.length}) →
      </button>

      {/* Full Article / View All Modal */}
      {isViewAllModalOpen && (
        <div style={modalOverlayStyle} onClick={() => setIsViewAllModalOpen(false)}>
          <div style={modalContentStyle} onClick={(e) => e.stopPropagation()}>
            <div style={modalHeaderStyle}>
              <h3 style={modalTitleStyle}>
                {modalArticle ? 'Article Full View & Metadata' : 'All Market News Feed Archive'}
              </h3>
              <button
                type="button"
                onClick={() => setIsViewAllModalOpen(false)}
                style={modalCloseBtnStyle}
              >
                ✕
              </button>
            </div>
            <div style={modalBodyStyle}>
              {modalArticle ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', color: '#64748b' }}>
                    <strong>Nguồn: {modalArticle.source}</strong>
                    <span>Thời gian: {new Date(modalArticle.publishedAt).toLocaleString()}</span>
                  </div>
                  <h2 style={{ fontSize: '1.15rem', color: '#0f172a', margin: '0.5rem 0' }}>{modalArticle.title}</h2>
                  <p style={{ fontSize: '0.9rem', lineHeight: '1.6', color: '#334155' }}>{modalArticle.content}</p>
                  {modalArticle.url && (
                    <a href={modalArticle.url} target="_blank" rel="noreferrer" style={{ color: '#2563eb', fontSize: '0.8rem' }}>
                      Mở link gốc ({modalArticle.url}) ↗
                    </a>
                  )}
                  {modalArticle.sentiment && (
                    <div style={{ padding: '0.75rem', backgroundColor: '#f8fafc', borderRadius: '6px', border: '1px solid #e2e8f0', marginTop: '0.5rem' }}>
                      <strong>Kết quả Sentiment:</strong> {modalArticle.sentiment.sentiment} ({Math.round(modalArticle.sentiment.score * 100)}%) - Model: {modalArticle.sentiment.model.name} ({modalArticle.sentiment.model.version})
                    </div>
                  )}
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  {news.map((item, idx) => (
                    <div
                      key={item.id}
                      onClick={() => {
                        onSelectArticle?.(item.id);
                        setModalArticle(item);
                      }}
                      style={{ padding: '0.6rem', border: '1px solid #e2e8f0', borderRadius: '6px', cursor: 'pointer' }}
                    >
                      <div style={{ fontSize: '0.7rem', color: '#64748b' }}>#{idx + 1} • {item.source} • {new Date(item.publishedAt).toLocaleTimeString()}</div>
                      <div style={{ fontWeight: '700', fontSize: '0.85rem', color: '#0f172a' }}>{item.title}</div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
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
  marginBottom: '1rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#475569',
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

const emptyStyle: React.CSSProperties = {
  textAlign: 'center',
  padding: '2rem 1rem',
  color: '#94a3b8',
  fontSize: '0.8rem',
};

const cardStyle: React.CSSProperties = {
  backgroundColor: '#f8fafc',
  border: '1px solid #e2e8f0',
  borderRadius: '6px',
  padding: '0.75rem',
  cursor: 'pointer',
  transition: 'all 0.15s ease',
};

const activeCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(59, 130, 246, 0.08)',
  border: '1px solid #3b82f6',
  borderRadius: '6px',
  padding: '0.75rem',
  cursor: 'pointer',
  boxShadow: '0 0 0 1px #3b82f6',
  transition: 'all 0.15s ease',
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
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#0f172a',
  margin: '0 0 0.4rem 0',
  lineHeight: '1.3',
};

const summaryStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#475569',
  margin: '0 0 0.6rem 0',
  lineHeight: '1.4',
};

const sentimentRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderTop: '1px dashed #cbd5e1',
  paddingTop: '0.5rem',
};

const badgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
  border: '1px solid transparent',
};

const sourceBadgeStyle: React.CSSProperties = { fontSize: '0.58rem', fontWeight: 800, padding: '0.1rem 0.3rem', borderRadius: '4px' };
const liveSourceStyle: React.CSSProperties = { color: '#047857', background: '#d1fae5' };
const demoSourceStyle: React.CSSProperties = { color: '#92400e', background: '#fef3c7' };

const modelMetaStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
  fontFamily: 'monospace',
};

const viewAllStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#2563eb',
  cursor: 'pointer',
  fontSize: '0.8rem',
  fontWeight: '600',
  textAlign: 'center',
  paddingTop: '0.75rem',
  borderTop: '1px solid #e2e8f0',
  marginTop: '0.5rem',
  width: '100%',
};

const modalOverlayStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
  backgroundColor: 'rgba(15, 23, 42, 0.65)',
  backdropFilter: 'blur(4px)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 1200,
  padding: '1rem',
};

const modalContentStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  borderRadius: '12px',
  border: '1px solid #cbd5e1',
  boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.25)',
  width: '100%',
  maxWidth: '640px',
  maxHeight: '85vh',
  display: 'flex',
  flexDirection: 'column',
  overflow: 'hidden',
};

const modalHeaderStyle: React.CSSProperties = {
  padding: '1rem 1.25rem',
  borderBottom: '1px solid #e2e8f0',
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  backgroundColor: '#f8fafc',
};

const modalTitleStyle: React.CSSProperties = {
  margin: 0,
  fontSize: '1rem',
  fontWeight: '700',
  color: '#0f172a',
};

const modalCloseBtnStyle: React.CSSProperties = {
  background: 'transparent',
  border: 'none',
  fontSize: '1.1rem',
  color: '#64748b',
  cursor: 'pointer',
};

const modalBodyStyle: React.CSSProperties = {
  padding: '1.25rem',
  overflowY: 'auto',
};
