import { useState, useEffect } from 'react';

export interface SourceConfigItem {
  id: string;
  name: string;
  url: string;
  enabled: boolean;
  type: 'website' | 'rss';
}

const STORAGE_KEY = 'crypto-strategy-lab-source-config';

const DEFAULT_SOURCES: SourceConfigItem[] = [
  { id: 'src-1', name: 'CoinDesk Web Scraper', url: 'https://www.coindesk.com', enabled: true, type: 'website' },
  { id: 'src-2', name: 'Cointelegraph Web Scraper', url: 'https://cointelegraph.com', enabled: true, type: 'website' },
  { id: 'src-3', name: 'The Block News Scraper', url: 'https://www.theblock.co', enabled: true, type: 'website' },
  { id: 'src-4', name: 'Decrypt Media Scraper', url: 'https://decrypt.co', enabled: false, type: 'website' },
  { id: 'rss-1', name: 'CoinDesk RSS Feed', url: 'https://www.coindesk.com/arc/outboundfeeds/rss/', enabled: true, type: 'rss' },
  { id: 'rss-2', name: 'Cointelegraph RSS Feed', url: 'https://cointelegraph.com/rss', enabled: true, type: 'rss' },
  { id: 'rss-3', name: 'Bitcoin Magazine RSS', url: 'https://bitcoinmagazine.com/feed', enabled: false, type: 'rss' },
];

interface SourceConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSaveSuccess?: () => void;
}

export function SourceConfigModal({
  isOpen,
  onClose,
  onSaveSuccess,
}: SourceConfigModalProps) {
  const [sources, setSources] = useState<SourceConfigItem[]>(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      return saved ? JSON.parse(saved) : DEFAULT_SOURCES;
    } catch {
      return DEFAULT_SOURCES;
    }
  });

  const [timeoutSecs, setTimeoutSecs] = useState<number>(10);
  const [maxArticles, setMaxArticles] = useState<number>(20);
  const [savedNotice, setSavedNotice] = useState(false);

  useEffect(() => {
    if (isOpen) {
      try {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (saved) setSources(JSON.parse(saved));
      } catch {
        // Fallback
      }
      setSavedNotice(false);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const toggleSource = (id: string) => {
    setSources((prev) =>
      prev.map((s) => (s.id === id ? { ...s, enabled: !s.enabled } : s))
    );
  };

  const handleUrlChange = (id: string, newUrl: string) => {
    setSources((prev) =>
      prev.map((s) => (s.id === id ? { ...s, url: newUrl } : s))
    );
  };

  const handleSave = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(sources));
      setSavedNotice(true);
      onSaveSuccess?.();
      setTimeout(() => {
        setSavedNotice(false);
        onClose();
      }, 700);
    } catch (err) {
      alert(`Could not save configuration: ${String(err)}`);
    }
  };

  const websiteSources = sources.filter((s) => s.type === 'website');
  const rssSources = sources.filter((s) => s.type === 'rss');

  return (
    <div style={overlayStyle} onClick={onClose}>
      <div style={modalStyle} onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div style={headerStyle}>
          <div>
            <h3 style={titleStyle}>News Scraper & Feed Source Configuration</h3>
            <span style={subtitleStyle}>Configure active web scrapers, RSS feeds, and crawling parameters</span>
          </div>
          <button type="button" onClick={onClose} style={closeBtnStyle} aria-label="Close modal">
            ✕
          </button>
        </div>

        {/* Content Body */}
        <div style={bodyStyle}>
          {savedNotice && (
            <div style={noticeStyle}>
              ✓ Cấu hình nguồn cào tin đã được lưu thành công!
            </div>
          )}

          {/* Section 1: Web Scraper Sources */}
          <div style={sectionStyle}>
            <h4 style={sectionTitleStyle}>Web HTML Scrapers</h4>
            <div style={sourceListStyle}>
              {websiteSources.map((source) => (
                <div key={source.id} style={sourceRowStyle}>
                  <label style={checkboxLabelStyle}>
                    <input
                      type="checkbox"
                      checked={source.enabled}
                      onChange={() => toggleSource(source.id)}
                      style={checkboxStyle}
                    />
                    <span style={sourceNameStyle}>{source.name}</span>
                  </label>
                  <input
                    type="url"
                    value={source.url}
                    onChange={(e) => handleUrlChange(source.id, e.target.value)}
                    style={urlInputStyle}
                    placeholder="https://..."
                  />
                  <span style={source.enabled ? activeBadgeStyle : inactiveBadgeStyle}>
                    {source.enabled ? 'ACTIVE' : 'OFF'}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Section 2: RSS Feeds */}
          <div style={sectionStyle}>
            <h4 style={sectionTitleStyle}>RSS / Atom Feed Endpoints</h4>
            <div style={sourceListStyle}>
              {rssSources.map((source) => (
                <div key={source.id} style={sourceRowStyle}>
                  <label style={checkboxLabelStyle}>
                    <input
                      type="checkbox"
                      checked={source.enabled}
                      onChange={() => toggleSource(source.id)}
                      style={checkboxStyle}
                    />
                    <span style={sourceNameStyle}>{source.name}</span>
                  </label>
                  <input
                    type="url"
                    value={source.url}
                    onChange={(e) => handleUrlChange(source.id, e.target.value)}
                    style={urlInputStyle}
                    placeholder="https://.../rss"
                  />
                  <span style={source.enabled ? activeBadgeStyle : inactiveBadgeStyle}>
                    {source.enabled ? 'ACTIVE' : 'OFF'}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Section 3: Crawler Request Settings */}
          <div style={sectionStyle}>
            <h4 style={sectionTitleStyle}>Crawler Networking & Safety Limits</h4>
            <div style={settingsGridStyle}>
              <div style={settingItemStyle}>
                <label style={settingLabelStyle}>Request Timeout (seconds):</label>
                <input
                  type="number"
                  min="3"
                  max="60"
                  value={timeoutSecs}
                  onChange={(e) => setTimeoutSecs(parseInt(e.target.value) || 10)}
                  style={numberInputStyle}
                />
              </div>
              <div style={settingItemStyle}>
                <label style={settingLabelStyle}>Max Articles Per Cycle:</label>
                <input
                  type="number"
                  min="5"
                  max="100"
                  value={maxArticles}
                  onChange={(e) => setMaxArticles(parseInt(e.target.value) || 20)}
                  style={numberInputStyle}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div style={footerStyle}>
          <button type="button" onClick={onClose} style={cancelBtnStyle}>
            Cancel
          </button>
          <button type="button" onClick={handleSave} style={saveBtnStyle}>
            Save Configuration
          </button>
        </div>
      </div>
    </div>
  );
}

// ==========================================
// STYLING
// ==========================================
const overlayStyle: React.CSSProperties = {
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
  zIndex: 1100,
  padding: '1rem',
};

const modalStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  borderRadius: '12px',
  border: '1px solid #cbd5e1',
  boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.25)',
  width: '100%',
  maxWidth: '680px',
  maxHeight: '90vh',
  display: 'flex',
  flexDirection: 'column',
  overflow: 'hidden',
  color: '#0f172a',
};

const headerStyle: React.CSSProperties = {
  padding: '1.25rem 1.5rem',
  borderBottom: '1px solid #e2e8f0',
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'flex-start',
  backgroundColor: '#f8fafc',
};

const titleStyle: React.CSSProperties = {
  margin: 0,
  fontSize: '1.05rem',
  fontWeight: '700',
  color: '#0f172a',
};

const subtitleStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  marginTop: '0.2rem',
  display: 'block',
};

const closeBtnStyle: React.CSSProperties = {
  background: 'transparent',
  border: 'none',
  fontSize: '1.1rem',
  color: '#64748b',
  cursor: 'pointer',
  padding: '0.2rem 0.5rem',
  borderRadius: '4px',
};

const bodyStyle: React.CSSProperties = {
  padding: '1.25rem 1.5rem',
  overflowY: 'auto',
  display: 'flex',
  flexDirection: 'column',
  gap: '1.25rem',
};

const noticeStyle: React.CSSProperties = {
  backgroundColor: '#ecfdf5',
  border: '1px solid #a7f3d0',
  color: '#047857',
  padding: '0.6rem 0.85rem',
  borderRadius: '6px',
  fontSize: '0.8rem',
  fontWeight: '600',
};

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.6rem',
};

const sectionTitleStyle: React.CSSProperties = {
  margin: 0,
  fontSize: '0.8rem',
  fontWeight: '700',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  color: '#475569',
};

const sourceListStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const sourceRowStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.75rem',
  backgroundColor: '#f8fafc',
  padding: '0.5rem 0.75rem',
  borderRadius: '6px',
  border: '1px solid #e2e8f0',
};

const checkboxLabelStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
  width: '200px',
  cursor: 'pointer',
};

const checkboxStyle: React.CSSProperties = {
  cursor: 'pointer',
};

const sourceNameStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '600',
  color: '#1e293b',
};

const urlInputStyle: React.CSSProperties = {
  flex: 1,
  backgroundColor: '#ffffff',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.35rem 0.6rem',
  fontSize: '0.75rem',
  fontFamily: 'monospace',
};

const activeBadgeStyle: React.CSSProperties = {
  backgroundColor: 'rgba(16, 185, 129, 0.15)',
  color: '#059669',
  border: '1px solid #10b981',
  fontSize: '0.65rem',
  fontWeight: '700',
  padding: '0.15rem 0.4rem',
  borderRadius: '4px',
};

const inactiveBadgeStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#64748b',
  border: '1px solid #cbd5e1',
  fontSize: '0.65rem',
  fontWeight: '600',
  padding: '0.15rem 0.4rem',
  borderRadius: '4px',
};

const settingsGridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
  gap: '1rem',
};

const settingItemStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.35rem',
};

const settingLabelStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  fontWeight: '600',
};

const numberInputStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.4rem 0.6rem',
  fontSize: '0.8rem',
  width: '120px',
};

const footerStyle: React.CSSProperties = {
  padding: '1rem 1.5rem',
  borderTop: '1px solid #e2e8f0',
  display: 'flex',
  justifyContent: 'flex-end',
  gap: '0.75rem',
  backgroundColor: '#f8fafc',
};

const cancelBtnStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#475569',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.5rem 1rem',
  fontSize: '0.85rem',
  fontWeight: '600',
  cursor: 'pointer',
};

const saveBtnStyle: React.CSSProperties = {
  backgroundColor: '#2563eb',
  color: '#ffffff',
  border: 'none',
  borderRadius: '6px',
  padding: '0.5rem 1.25rem',
  fontSize: '0.85rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 4px rgba(37, 99, 235, 0.2)',
};
