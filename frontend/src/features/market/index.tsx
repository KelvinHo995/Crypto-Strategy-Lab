import { useEffect, useState } from 'react';
import { MarketGrid } from './components/MarketGrid';
import { MarketStatusBar } from './components/MarketStatusBar';
import { RecentTicksPanel } from './components/RecentTicksPanel';
import { fetchMarkets } from '../../shared/api';
import { useAppMode } from '../../shared/auth';
import type { MarketInfo } from '../../types/candle';
import { DEFAULT_MARKETS } from './services/marketCatalog';

export function MarketDashboard() {
  const mode = useAppMode();
  const [markets, setMarkets] = useState<MarketInfo[]>(DEFAULT_MARKETS);
  const [catalogError, setCatalogError] = useState('');

  useEffect(() => {
    if (mode !== 'LIVE') return;
    fetchMarkets().then(items => {
      setMarkets(items);
      setCatalogError('');
    }).catch(error => setCatalogError(`Không tải được market catalog: ${String(error)}`));
  }, [mode]);

  return (
    <div style={dashboardContainerStyle}>
      {catalogError && <div style={errorStyle}>{catalogError}</div>}
      {/* Cột trái: Lưới biểu đồ (Multi-Charts Grid) */}
      <div style={leftColumnStyle}>
        <MarketGrid markets={markets} />
      </div>

      {/* Cột phải: Sidebar Panel (Trạng thái kết nối & Legend & Recent Trades) */}
      <div style={rightColumnStyle}>
        <MarketStatusBar />
        <RecentTicksPanel markets={markets} />
      </div>
    </div>
  );
}

const errorStyle: React.CSSProperties = {
  width: '100%', color: '#b91c1c', backgroundColor: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', padding: '0.65rem 0.8rem', fontSize: '0.8rem',
};

// ==========================================
// STYLING PRESET
// ==========================================
const dashboardContainerStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1rem',
  width: '100%',
  alignItems: 'stretch',
  flexWrap: 'wrap', // Responsive wrapping for smaller screens
};

const leftColumnStyle: React.CSSProperties = {
  flex: '1 1 700px', // Flexible growth, min-width 700px before wrapping
  display: 'flex',
  flexDirection: 'column',
};

const rightColumnStyle: React.CSSProperties = {
  width: '280px',
  flexShrink: 0,
  display: 'flex',
  flexDirection: 'column',
  gap: '1rem',
};
