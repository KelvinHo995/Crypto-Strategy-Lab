export interface NewsCrawlerHeaderProps {
  onCrawlStart: () => void;
  isCrawling: boolean;
  activeSource: 'website' | 'rss' | 'html';
  onSourceChange: (source: 'website' | 'rss' | 'html') => void;
  activeAsset: string;
  onAssetChange: (asset: string) => void;
  refreshInterval: string;
  onRefreshIntervalChange: (interval: string) => void;
  onOpenSourceConfig: () => void;
}

export function NewsCrawlerHeader({
  onCrawlStart,
  isCrawling,
  activeSource,
  onSourceChange,
  activeAsset,
  onAssetChange,
  refreshInterval,
  onRefreshIntervalChange,
  onOpenSourceConfig,
}: NewsCrawlerHeaderProps) {
  return (
    <div style={containerStyle}>
      {/* 1. Source selector Tabs */}
      <div style={leftSectionStyle}>
        <div style={tabsGroupStyle}>
          <button
            type="button"
            onClick={() => onSourceChange('website')}
            style={activeSource === 'website' ? activeTabStyle : tabStyle}
            disabled={isCrawling}
          >
            Website Scraper
          </button>
          <button
            type="button"
            onClick={() => onSourceChange('rss')}
            style={activeSource === 'rss' ? activeTabStyle : tabStyle}
            disabled={isCrawling}
          >
            RSS Feeds
          </button>
          <button
            type="button"
            onClick={() => onSourceChange('html')}
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
            onChange={(e) => onAssetChange(e.target.value)}
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
            onChange={(e) => onRefreshIntervalChange(e.target.value)}
            style={selectStyle}
            disabled={isCrawling}
          >
            <option value="1m">1 Min</option>
            <option value="2m">2 Mins</option>
            <option value="5m">5 Mins</option>
            <option value="off">Off</option>
          </select>
        </div>

        {/* Source Configuration button */}
        <button
          type="button"
          onClick={onOpenSourceConfig}
          style={configBtnStyle}
          disabled={isCrawling}
          title="Open Scraper & Feed URLs configuration"
        >
          ⚙ Source Config
        </button>

        {/* Action button */}
        <button
          type="button"
          onClick={onCrawlStart}
          disabled={isCrawling}
          style={isCrawling ? activeCrawlBtnStyle : crawlBtnStyle}
        >
          {isCrawling ? (
            <span style={crawlingFlexStyle}>
              <span style={spinnerStyle} /> Analyzing through live services...
            </span>
          ) : (
            'Analyze sample article'
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
  backgroundColor: '#f1f5f9',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '2px',
};

const tabStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#64748b',
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
  backgroundColor: '#ffffff',
  color: '#2563eb',
  border: '1px solid #cbd5e1',
  padding: '0.4rem 0.85rem',
  borderRadius: '4px',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  outline: 'none',
  boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
};

const rightSectionStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.85rem',
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
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.35rem 0.5rem',
  fontSize: '0.8rem',
  cursor: 'pointer',
  outline: 'none',
};

const configBtnStyle: React.CSSProperties = {
  backgroundColor: '#f8fafc',
  color: '#334155',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.4rem 0.75rem',
  fontSize: '0.8rem',
  fontWeight: '600',
  cursor: 'pointer',
  outline: 'none',
  transition: 'all 0.15s ease',
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
