import { useState } from 'react';

interface NewsCrawlerHeaderProps {
  onCrawlStart: () => void;
  isCrawling: boolean;
}

export function NewsCrawlerHeader({
  onCrawlStart,
  isCrawling,
}: NewsCrawlerHeaderProps) {
  const [activeSource, setActiveSource] = useState<'website' | 'rss' | 'html'>('website');
  const [activeAsset, setActiveAsset] = useState<string>('ALL');
  const [refreshInterval, setRefreshInterval] = useState<string>('2m');

  return (
    <div style={containerStyle}>
      {/* 1. Source selector Tabs */}
      <div style={leftSectionStyle}>
        <div style={tabsGroupStyle}>
          <button
            onClick={() => setActiveSource('website')}
            style={activeSource === 'website' ? activeTabStyle : tabStyle}
            disabled={isCrawling}
          >
            Website Scraper
          </button>
          <button
            onClick={() => setActiveSource('rss')}
            style={activeSource === 'rss' ? activeTabStyle : tabStyle}
            disabled={isCrawling}
          >
            RSS Feeds
          </button>
          <button
            onClick={() => setActiveSource('html')}
            style={activeSource === 'html' ? activeTabStyle : tabStyle}
            disabled={isCrawling}
          >
            Raw HTML Import
          </button>
        </div>
      </div>

      {/* 2. Asset filters & Refresh controllers */}
      <div style={rightSectionStyle}>
        {/* Asset Filter */}
        <div style={groupStyle}>
          <span style={labelStyle}>Asset:</span>
          <select
            value={activeAsset}
            onChange={(e) => setActiveAsset(e.target.value)}
            style={selectStyle}
            disabled={isCrawling}
          >
            <option value="ALL">All Crypto</option>
            <option value="BTC">BTC Only</option>
            <option value="ETH">ETH Only</option>
            <option value="SOL">SOL Only</option>
            <option value="BNB">BNB Only</option>
          </select>
        </div>

        {/* Auto Refresh timer */}
        <div style={groupStyle}>
          <span style={labelStyle}>Auto Refresh:</span>
          <select
            value={refreshInterval}
            onChange={(e) => setRefreshInterval(e.target.value)}
            style={selectStyle}
            disabled={isCrawling}
          >
            <option value="1m">1 Min</option>
            <option value="2m">2 Mins</option>
            <option value="5m">5 Mins</option>
            <option value="off">Off</option>
          </select>
        </div>

        {/* Action buttons */}
        <button style={configBtnStyle} disabled={isCrawling}>
          Source Config
        </button>

        <button
          onClick={onCrawlStart}
          disabled={isCrawling}
          style={isCrawling ? activeCrawlBtnStyle : crawlBtnStyle}
        >
          {isCrawling ? (
            <span style={crawlingFlexStyle}>
              <span style={spinnerStyle} /> Crawling Live Feeds...
            </span>
          ) : (
            'Bắt đầu crawl'
          )}
        </button>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const containerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '0.75rem 1.25rem',
  flexWrap: 'wrap',
  gap: '1rem',
};

const leftSectionStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
};

const tabsGroupStyle: React.CSSProperties = {
  display: 'flex',
  backgroundColor: '#ffffff',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '2px',
};

const tabStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: 'none',
  padding: '0.4rem 0.85rem',
  borderRadius: '4px',
  fontSize: '0.8rem',
  fontWeight: '600',
  cursor: 'pointer',
  transition: 'all 0.15s',
  outline: 'none',
};

const activeTabStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#2563eb',
  border: 'none',
  padding: '0.4rem 0.85rem',
  borderRadius: '4px',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  outline: 'none',
};

const rightSectionStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '1rem',
  flexWrap: 'wrap',
};

const groupStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.35rem',
};

const labelStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  fontWeight: '600',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#ffffff',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.35rem 0.5rem',
  fontSize: '0.8rem',
  cursor: 'pointer',
  outline: 'none',
};

const configBtnStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#cbd5e1',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.4rem 0.75rem',
  fontSize: '0.8rem',
  fontWeight: '600',
  cursor: 'pointer',
  outline: 'none',
};

const crawlBtnStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.4rem 1rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  outline: 'none',
  boxShadow: '0 2px 4px rgba(59, 130, 246, 0.2)',
};

const activeCrawlBtnStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#475569',
  border: '1px solid #e2e8f0',
  borderRadius: '4px',
  padding: '0.4rem 1rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'not-allowed',
  outline: 'none',
};

const crawlingFlexStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
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
