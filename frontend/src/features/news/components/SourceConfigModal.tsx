import { useState } from 'react';

export interface CrawlerSource {
  id: string;
  name: string;
  type: 'RSS' | 'SCRAPER' | 'API';
  url: string;
  enabled: boolean;
  frequency: string; // e.g., '5m', '15m'
  lastStatus: 'OK' | 'ERROR' | 'IDLE';
}

const DEFAULT_SOURCES: CrawlerSource[] = [
  {
    id: 'src-1',
    name: 'CoinDesk Main RSS',
    type: 'RSS',
    url: 'https://www.coindesk.com/arc/outboundfeeds/rss/',
    enabled: true,
    frequency: '5m',
    lastStatus: 'OK',
  },
  {
    id: 'src-2',
    name: 'The Block Crypto News',
    type: 'RSS',
    url: 'https://www.theblock.co/rss.xml',
    enabled: true,
    frequency: '5m',
    lastStatus: 'OK',
  },
  {
    id: 'src-3',
    name: 'Cointelegraph News Wire',
    type: 'RSS',
    url: 'https://cointelegraph.com/rss',
    enabled: true,
    frequency: '5m',
    lastStatus: 'OK',
  },
  {
    id: 'src-4',
    name: 'Decrypt Media Feed',
    type: 'RSS',
    url: 'https://decrypt.co/feed',
    enabled: true,
    frequency: '10m',
    lastStatus: 'OK',
  },
  {
    id: 'src-5',
    name: 'Binance Announcements Scraper',
    type: 'SCRAPER',
    url: 'https://www.binance.com/en/support/announcement',
    enabled: false,
    frequency: '15m',
    lastStatus: 'IDLE',
  },
];

interface SourceConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SourceConfigModal({ isOpen, onClose }: SourceConfigModalProps) {
  const [sources, setSources] = useState<CrawlerSource[]>(DEFAULT_SOURCES);
  const [newName, setNewName] = useState('');
  const [newUrl, setNewUrl] = useState('');
  const [newType, setNewType] = useState<'RSS' | 'SCRAPER'>('RSS');

  if (!isOpen) return null;

  const handleToggle = (id: string) => {
    setSources((prev) =>
      prev.map((s) => (s.id === id ? { ...s, enabled: !s.enabled } : s))
    );
  };

  const handleAddSource = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim() || !newUrl.trim()) return;

    const newSource: CrawlerSource = {
      id: `src-${Date.now()}`,
      name: newName.trim(),
      type: newType,
      url: newUrl.trim(),
      enabled: true,
      frequency: '5m',
      lastStatus: 'OK',
    };

    setSources((prev) => [...prev, newSource]);
    setNewName('');
    setNewUrl('');
  };

  const handleDelete = (id: string) => {
    setSources((prev) => prev.filter((s) => s.id !== id));
  };

  return (
    <div style={overlayStyle}>
      <div style={modalStyle}>
        {/* Header */}
        <div style={headerStyle}>
          <div>
            <h3 style={titleStyle}>⚙ Crawler Source Configuration</h3>
            <span style={subtitleStyle}>Manage active RSS Feeds & Web Scraper ingestion pipelines</span>
          </div>
          <button onClick={onClose} style={closeBtnStyle}>✕</button>
        </div>

        {/* Source List */}
        <div style={bodyStyle}>
          <div style={listStyle}>
            {sources.map((src) => (
              <div key={src.id} style={itemStyle}>
                <div style={infoColStyle}>
                  <div style={nameRowStyle}>
                    <span style={nameStyle}>{src.name}</span>
                    <span style={typeBadgeStyle}>{src.type}</span>
                    <span style={statusBadgeStyle(src.lastStatus)}>
                      {src.lastStatus}
                    </span>
                  </div>
                  <span style={urlStyle}>{src.url}</span>
                </div>

                <div style={actionColStyle}>
                  <label style={toggleSwitchLabel}>
                    <input
                      type="checkbox"
                      checked={src.enabled}
                      onChange={() => handleToggle(src.id)}
                      style={checkboxStyle}
                    />
                    <span style={toggleTextStyle}>
                      {src.enabled ? 'Active' : 'Paused'}
                    </span>
                  </label>
                  <button
                    onClick={() => handleDelete(src.id)}
                    style={deleteBtnStyle}
                    title="Remove source"
                  >
                    🗑
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Add Source Form */}
          <form onSubmit={handleAddSource} style={addFormStyle}>
            <h4 style={formTitleStyle}>+ Add New Ingestion Endpoint</h4>
            <div style={formGridStyle}>
              <input
                type="text"
                placeholder="Source Name (e.g. Bankless RSS)"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                style={inputStyle}
                required
              />
              <select
                value={newType}
                onChange={(e) => setNewType(e.target.value as 'RSS' | 'SCRAPER')}
                style={selectStyle}
              >
                <option value="RSS">RSS Feed</option>
                <option value="SCRAPER">Web Scraper</option>
              </select>
              <input
                type="url"
                placeholder="https://example.com/feed.xml"
                value={newUrl}
                onChange={(e) => setNewUrl(e.target.value)}
                style={{ ...inputStyle, gridColumn: '1 / -1' }}
                required
              />
            </div>
            <button type="submit" style={addBtnStyle}>
              Add Source Endpoint
            </button>
          </form>
        </div>

        {/* Footer */}
        <div style={footerStyle}>
          <button onClick={onClose} style={doneBtnStyle}>
            Close & Save
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
  backgroundColor: 'rgba(0,0,0,0.65)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 99999,
  fontFamily: 'system-ui, sans-serif',
};

const modalStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #cbd5e1',
  borderRadius: '10px',
  width: '95%',
  maxWidth: '640px',
  padding: '1.5rem',
  boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.3)',
  color: '#0f172a',
};

const headerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'flex-start',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.75rem',
  marginBottom: '1rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '1.05rem',
  fontWeight: '700',
  color: '#0f172a',
  margin: 0,
};

const subtitleStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
};

const closeBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#94a3b8',
  fontSize: '1.1rem',
  cursor: 'pointer',
};

const bodyStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1.25rem',
  maxHeight: '420px',
  overflowY: 'auto',
  paddingRight: '4px',
};

const listStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.6rem',
};

const itemStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  backgroundColor: '#f8fafc',
  border: '1px solid #e2e8f0',
  borderRadius: '6px',
  padding: '0.65rem 0.85rem',
};

const infoColStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.2rem',
  maxWidth: '70%',
};

const nameRowStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
};

const nameStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#0f172a',
};

const typeBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  backgroundColor: '#e2e8f0',
  color: '#3b82f6',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
};

const statusBadgeStyle = (status: 'OK' | 'ERROR' | 'IDLE'): React.CSSProperties => ({
  fontSize: '0.6rem',
  fontWeight: '700',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
  backgroundColor: status === 'OK' ? 'rgba(16, 185, 129, 0.15)' : status === 'ERROR' ? 'rgba(239, 68, 68, 0.15)' : 'rgba(148, 163, 184, 0.15)',
  color: status === 'OK' ? '#10b981' : status === 'ERROR' ? '#ef4444' : '#64748b',
});

const urlStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  fontFamily: 'monospace',
  overflow: 'hidden',
  textOverflow: 'ellipsis',
  whiteSpace: 'nowrap',
};

const actionColStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.75rem',
};

const toggleSwitchLabel: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.35rem',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'pointer',
};

const checkboxStyle: React.CSSProperties = {
  cursor: 'pointer',
};

const toggleTextStyle: React.CSSProperties = {
  color: '#475569',
};

const deleteBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  cursor: 'pointer',
  fontSize: '0.9rem',
  color: '#ef4444',
  padding: '0.2rem',
};

const addFormStyle: React.CSSProperties = {
  backgroundColor: '#f1f5f9',
  border: '1px dashed #cbd5e1',
  borderRadius: '6px',
  padding: '0.85rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.6rem',
};

const formTitleStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#1e293b',
  margin: 0,
};

const formGridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: '1fr 120px',
  gap: '0.5rem',
};

const inputStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.4rem 0.6rem',
  fontSize: '0.78rem',
  outline: 'none',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#0f172a',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.4rem',
  fontSize: '0.78rem',
  outline: 'none',
  cursor: 'pointer',
};

const addBtnStyle: React.CSSProperties = {
  backgroundColor: '#2563eb',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.45rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
};

const footerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'flex-end',
  borderTop: '1px solid #e2e8f0',
  paddingTop: '0.85rem',
  marginTop: '0.85rem',
};

const doneBtnStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  color: '#ffffff',
  border: 'none',
  borderRadius: '6px',
  padding: '0.5rem 1.25rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
};
