import { useState } from 'react';
import { ChartCard } from './ChartCard';
import type { MarketInfo } from '../../../types/candle';

const CHART_CONFIGS = [
  { id: 1, defaultSymbol: 'BTCUSDT', defaultTimeframe: '5m' },
  { id: 2, defaultSymbol: 'BTCUSDT', defaultTimeframe: '15m' },
  { id: 3, defaultSymbol: 'BTCUSDT', defaultTimeframe: '1h' },
  { id: 4, defaultSymbol: 'BTCUSDT', defaultTimeframe: '4h' },
];

export function MarketGrid({ markets }: { markets: MarketInfo[] }) {
  const [layout, setLayout] = useState<1 | 2 | 4>(4);
  const [maximizedId, setMaximizedId] = useState<number | null>(null);

  const handleToggleMaximize = (id: number) => {
    if (maximizedId === id) {
      setMaximizedId(null); // Restore split view
    } else {
      setMaximizedId(id); // Maximize this chart
    }
  };

  // Filter charts to render based on layout and maximized state
  const getVisibleCharts = () => {
    if (maximizedId !== null) {
      return CHART_CONFIGS.filter((c) => c.id === maximizedId);
    }
    // Return subset based on active layout selection
    return CHART_CONFIGS.slice(0, layout);
  };

  const visibleCharts = getVisibleCharts();

  // Determine CSS Grid styles based on visibility and layout
  const getGridStyle = (): React.CSSProperties => {
    if (maximizedId !== null || visibleCharts.length === 1) {
      return {
        display: 'grid',
        gridTemplateColumns: '1fr',
        height: 'calc(100vh - 180px)',
        minHeight: '450px',
        gap: '1rem',
      };
    }

    if (visibleCharts.length === 2) {
      return {
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))',
        height: 'calc(100vh - 180px)',
        minHeight: '450px',
        gap: '1rem',
      };
    }

    // Default 4 charts layout (2x2 grid)
    return {
      display: 'grid',
      gridTemplateColumns: 'repeat(2, 1fr)',
      gridTemplateRows: 'repeat(2, 1fr)',
      height: 'calc(100vh - 180px)',
      minHeight: '560px',
      gap: '1rem',
    };
  };

  return (
    <div style={containerStyle}>
      {/* Grid Controls Toolbar */}
      <div style={toolbarStyle}>
        <div style={toolbarTitleStyle}>
          <span style={titleIconStyle}></span> Multi-Timeframe Dashboard
        </div>
        
        {maximizedId === null && (
          <div style={layoutButtonGroupStyle}>
            <button
              onClick={() => setLayout(1)}
              style={layout === 1 ? activeButtonStyle : buttonStyle}
              title="1 Chart Layout"
            >
              [ 1 ]
            </button>
            <button
              onClick={() => setLayout(2)}
              style={layout === 2 ? activeButtonStyle : buttonStyle}
              title="2 Charts Split"
            >
              [ 1x2 ]
            </button>
            <button
              onClick={() => setLayout(4)}
              style={layout === 4 ? activeButtonStyle : buttonStyle}
              title="4 Charts Grid (2x2)"
            >
              [ 2x2 ]
            </button>
          </div>
        )}

        {maximizedId !== null && (
          <button onClick={() => setMaximizedId(null)} style={restoreButtonStyle}>
            Restore Grid View
          </button>
        )}
      </div>

      {/* Grid Canvas */}
      <div style={getGridStyle()}>
        {visibleCharts.map((chart) => (
          <ChartCard
            key={chart.id}
            id={chart.id}
            defaultSymbol={chart.defaultSymbol}
            defaultTimeframe={chart.defaultTimeframe}
            isMaximized={maximizedId === chart.id}
            onToggleMaximize={handleToggleMaximize}
            markets={markets}
          />
        ))}
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const containerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
  width: '100%',
  height: '100%',
};

const toolbarStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '0.5rem 1rem',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
};

const toolbarTitleStyle: React.CSSProperties = {
  fontSize: '0.9rem',
  fontWeight: '600',
  color: '#0f172a',
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
};

const titleIconStyle: React.CSSProperties = {
  color: '#2563eb',
};

const layoutButtonGroupStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.35rem',
  backgroundColor: '#e2e8f0',
  padding: '3px',
  borderRadius: '6px',
};

const buttonStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: 'none',
  padding: '0.25rem 0.6rem',
  borderRadius: '4px',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'pointer',
  transition: 'all 0.1s',
  outline: 'none',
};

const activeButtonStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  color: '#2563eb',
  border: 'none',
  padding: '0.25rem 0.6rem',
  borderRadius: '4px',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 1px 3px rgba(0,0,0,0.2)',
  outline: 'none',
};

const restoreButtonStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.35rem 0.75rem',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'pointer',
  outline: 'none',
  display: 'flex',
  alignItems: 'center',
  gap: '0.25rem',
  boxShadow: '0 2px 4px rgba(59, 130, 246, 0.2)',
};
