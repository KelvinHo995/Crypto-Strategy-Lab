import React from 'react';

export type NewsSourceTab = 'Website Scraper' | 'RSS Feeds' | 'Raw HTML Import';
export type AssetFilter = 'ALL' | 'BTC' | 'ETH' | 'SOL' | 'BNB' | 'XRP';

interface NewsCrawlerHeaderProps {
  activeSource: NewsSourceTab;
  onSelectSource: (source: NewsSourceTab) => void;
  activeAsset: AssetFilter;
  onSelectAsset: (asset: AssetFilter) => void;
  autoRefreshInterval: number; // in ms, 0 = off
  onSelectAutoRefresh: (interval: number) => void;
  onOpenSourceModal: () => void;
  onManualRefresh: () => void;
  isRefreshing?: boolean;
}

const SOURCES: NewsSourceTab[] = ['RSS Feeds', 'Website Scraper', 'Raw HTML Import'];
const ASSETS: AssetFilter[] = ['ALL', 'BTC', 'ETH', 'SOL', 'BNB', 'XRP'];

export function NewsCrawlerHeader({
  activeSource,
  onSelectSource,
  activeAsset,
  onSelectAsset,
  autoRefreshInterval,
  onSelectAutoRefresh,
  onOpenSourceModal,
  onManualRefresh,
  isRefreshing = false,
}: NewsCrawlerHeaderProps) {
  return (
    <div style={containerStyle}>
      {/* Top Controls Row */}
      <div style={topRowStyle}>
        {/* Source Switcher Tabs */}
        <div style={sourceTabsStyle}>
          {SOURCES.map((src) => (
            <button
              key={src}
              onClick={() => onSelectSource(src)}
              style={activeSource === src ? activeSourceTabBtnStyle : sourceTabBtnStyle}
            >
              {src === 'RSS Feeds' ? '📡 RSS Feeds' : src === 'Website Scraper' ? '🌐 Web Scraper' : '📄 Raw Ingest'}
            </button>
          ))}
        </div>

        {/* Right Tools: Auto Refresh & Source Config */}
        <div style={rightToolsStyle}>
          {/* Auto Refresh dropdown */}
          <div style={toolItemStyle}>
            <label style={toolLabelStyle}>Auto Refresh:</label>
            <select
              value={autoRefreshInterval}
              onChange={(e) => onSelectAutoRefresh(parseInt(e.target.value))}
              style={selectStyle}
            >
              <option value={0}>Off</option>
              <option value={60000}>1m</option>
              <option value={120000}>2m</option>
              <option value={300000}>5m</option>
            </select>
          </div>

          {/* Refresh Action */}
          <button
            onClick={onManualRefresh}
            disabled={isRefreshing}
            style={actionBtnStyle}
            title="Refresh latest news observations"
          >
            {isRefreshing ? '↻ Loading…' : '↻ Fetch'}
          </button>

          {/* Source Config Button */}
          <button
            onClick={onOpenSourceModal}
            style={configBtnStyle}
            title="Configure scraper & RSS sources"
          >
            ⚙ Source Config
          </button>
        </div>
      </div>

      {/* Bottom Filter Row: Asset Filter Pills */}
      <div style={filterRowStyle}>
        <span style={filterLabelStyle}>Filter Asset:</span>
        <div style={assetPillsStyle}>
          {ASSETS.map((asset) => (
            <button
              key={asset}
              onClick={() => onSelectAsset(asset)}
              style={activeAsset === asset ? activeAssetPillStyle : assetPillStyle}
            >
              {asset}
            </button>
          ))}
        </div>
        <span style={hintStyle}>
          Live multi-source articles normalized, cleansed & sentiment-scored via LLM FinBERT.
        </span>
      </div>
    </div>
  );
}

// ==========================================
// STYLING
// ==========================================
const containerStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '0.85rem 1rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
};

const topRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  flexWrap: 'wrap',
  gap: '0.75rem',
  borderBottom: '1px solid #f1f5f9',
  paddingBottom: '0.65rem',
};

const sourceTabsStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.4rem',
  backgroundColor: '#f1f5f9',
  padding: '3px',
  borderRadius: '6px',
};

const sourceTabBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#64748b',
  fontSize: '0.78rem',
  fontWeight: '600',
  padding: '0.35rem 0.65rem',
  borderRadius: '4px',
  cursor: 'pointer',
  transition: 'all 0.15s',
};

const activeSourceTabBtnStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #cbd5e1',
  color: '#0f172a',
  fontSize: '0.78rem',
  fontWeight: '700',
  padding: '0.35rem 0.65rem',
  borderRadius: '4px',
  cursor: 'pointer',
  boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
};

const rightToolsStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.6rem',
};

const toolItemStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.35rem',
};

const toolLabelStyle: React.CSSProperties = {
  fontSize: '0.72rem',
  color: '#64748b',
  fontWeight: '600',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#f8fafc',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.25rem 0.5rem',
  fontSize: '0.75rem',
  outline: 'none',
  cursor: 'pointer',
};

const actionBtnStyle: React.CSSProperties = {
  backgroundColor: '#f8fafc',
  color: '#3b82f6',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.25rem 0.6rem',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'pointer',
};

const configBtnStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.3rem 0.75rem',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'pointer',
  display: 'flex',
  alignItems: 'center',
  gap: '0.35rem',
};

const filterRowStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.75rem',
  flexWrap: 'wrap',
};

const filterLabelStyle: React.CSSProperties = {
  fontSize: '0.72rem',
  fontWeight: '700',
  color: '#475569',
  textTransform: 'uppercase',
  letterSpacing: '0.03em',
};

const assetPillsStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.35rem',
};

const assetPillStyle: React.CSSProperties = {
  backgroundColor: '#f1f5f9',
  border: '1px solid #e2e8f0',
  color: '#64748b',
  fontSize: '0.72rem',
  fontWeight: '700',
  padding: '0.2rem 0.55rem',
  borderRadius: '12px',
  cursor: 'pointer',
};

const activeAssetPillStyle: React.CSSProperties = {
  backgroundColor: '#2563eb',
  border: '1px solid #2563eb',
  color: '#ffffff',
  fontSize: '0.72rem',
  fontWeight: '700',
  padding: '0.2rem 0.55rem',
  borderRadius: '12px',
  cursor: 'pointer',
};

const hintStyle: React.CSSProperties = {
  marginLeft: 'auto',
  fontSize: '0.7rem',
  color: '#94a3b8',
};
