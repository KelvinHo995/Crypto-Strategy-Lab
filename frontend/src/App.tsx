import { useState, useEffect } from 'react';
import { ErrorBoundary, WebSocketStateBanner } from './shared/components/index';
import { wsManager } from './shared/ws';
import { MarketDashboard } from './features/market/index';
import { StrategyDiscoveryPage } from './features/strategy/index';
import { ExperimentDashboard } from './features/experiment/index';
import { NewsCrawlerDashboard } from './features/news/index';


function App() {
  const [activeTab, setActiveTab] = useState<'charts' | 'leaderboard' | 'builder' | 'news'>('charts');

  useEffect(() => {
    // Initiate WebSocket connection on app startup
    wsManager.connect();
    return () => {
      // Disconnect when app unmounts
      wsManager.disconnect();
    };
  }, []);

  return (
    <ErrorBoundary>
      {/* WebSocket Status Indicator Banner */}
      <WebSocketStateBanner />

      <div style={appContainerStyle}>
        {/* Navigation Sidebar / Header */}
        <header style={headerStyle}>
          <div style={logoAreaStyle}>
            <div style={logoIconStyle}>⚡</div>
            <h1 style={logoTitleStyle}>Crypto Strategy Lab</h1>
            <span style={badgeStyle}>v1.0.0</span>
          </div>

          <nav style={navStyle}>
            <button
              style={activeTab === 'charts' ? activeNavBtnStyle : navBtnStyle}
              onClick={() => setActiveTab('charts')}
            >
              Multi-Charts
            </button>
            <button
              style={activeTab === 'builder' ? activeNavBtnStyle : navBtnStyle}
              onClick={() => setActiveTab('builder')}
            >
              Strategy Builder
            </button>
            <button
              style={activeTab === 'leaderboard' ? activeNavBtnStyle : navBtnStyle}
              onClick={() => setActiveTab('leaderboard')}
            >
              Leaderboard
            </button>
            <button
              style={activeTab === 'news' ? activeNavBtnStyle : navBtnStyle}
              onClick={() => setActiveTab('news')}
            >
              Sentiment Feed
            </button>
          </nav>
        </header>

        {/* Main Work Area */}
        <main style={mainStyle}>
          {activeTab === 'charts' && <MarketDashboard />}

          {activeTab === 'builder' && <StrategyDiscoveryPage />}

          {activeTab === 'leaderboard' && <ExperimentDashboard />}

          {activeTab === 'news' && <NewsCrawlerDashboard />}
        </main>

        {/* Footer */}
        <footer style={footerStyle}>
          <p>© 2026 Crypto Strategy Lab — Monorepo Architecture Board</p>
        </footer>
      </div>
    </ErrorBoundary>
  );
}

// ==========================================
// STYLING PRESET (Premium dark/cyan theme)
// ==========================================
const appContainerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  minHeight: '100vh',
  backgroundColor: '#0c0f17', // Midnight blue
  color: '#e2e8f0',
  fontFamily: "'Inter', system-ui, -apple-system, sans-serif",
};

const headerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '1.25rem 2rem',
  borderBottom: '1px solid #1e293b',
  backgroundColor: '#0f172a',
  flexWrap: 'wrap',
  gap: '1rem',
};

const logoAreaStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.75rem',
  flexGrow: 1,
};

const logoIconStyle: React.CSSProperties = {
  fontSize: '1.5rem',
  background: 'linear-gradient(135deg, #06b6d4, #3b82f6)',
  padding: '0.25rem 0.5rem',
  borderRadius: '6px',
};

const logoTitleStyle: React.CSSProperties = {
  fontSize: '1.25rem',
  fontWeight: '700',
  letterSpacing: '-0.025em',
  margin: 0,
  background: 'linear-gradient(to right, #ffffff, #94a3b8)',
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
};

const badgeStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  backgroundColor: '#1e293b',
  padding: '0.2rem 0.5rem',
  borderRadius: '9999px',
  color: '#06b6d4',
  fontWeight: '600',
  border: '1px solid #334155',
};

const navStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.5rem',
};

const navBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: 'none',
  padding: '0.5rem 1rem',
  borderRadius: '6px',
  cursor: 'pointer',
  fontWeight: '500',
  fontSize: '0.9rem',
  transition: 'all 0.2s',
};

const activeNavBtnStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  color: '#06b6d4', // Cyan
  border: 'none',
  padding: '0.5rem 1rem',
  borderRadius: '6px',
  cursor: 'pointer',
  fontWeight: '600',
  fontSize: '0.9rem',
  boxShadow: '0 0 10px rgba(6, 182, 212, 0.1)',
};

const mainStyle: React.CSSProperties = {
  flexGrow: 1,
  padding: '2rem',
  maxWidth: '1200px',
  width: '100%',
  margin: '0 auto',
  boxSizing: 'border-box',
};

const footerStyle: React.CSSProperties = {
  textAlign: 'center',
  padding: '1.5rem',
  borderTop: '1px solid #1e293b',
  backgroundColor: '#0b0f19',
  color: '#475569',
  fontSize: '0.8rem',
};

export default App;
