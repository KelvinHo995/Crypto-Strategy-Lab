import { MarketGrid } from './components/MarketGrid';
import { MarketStatusBar } from './components/MarketStatusBar';
import { RecentTicksPanel } from './components/RecentTicksPanel';

export function MarketDashboard() {
  return (
    <div style={dashboardContainerStyle}>
      {/* Cột trái: Lưới biểu đồ (Multi-Charts Grid) */}
      <div style={leftColumnStyle}>
        <MarketGrid />
      </div>

      {/* Cột phải: Sidebar Panel (Trạng thái kết nối & Legend & Recent Trades) */}
      <div style={rightColumnStyle}>
        <MarketStatusBar />
        <RecentTicksPanel />
      </div>
    </div>
  );
}

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
