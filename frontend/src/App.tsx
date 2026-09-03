import { lazy, Suspense, useEffect, useState, type ComponentType } from 'react';
import { BarChart3, ChartCandlestick, FlaskConical, Newspaper, Radio, Trophy } from 'lucide-react';
import { ErrorBoundary, WebSocketStateBanner } from './shared/components/index';
import { AuthGate } from './shared/components/AuthGate';
import { useAppMode } from './shared/auth';
import { wsManager } from './shared/ws';
import { ThemeProvider, ThemeToggle } from './shared/theme';

const MarketDashboard = lazy(() => import('./features/market/index').then((module) => ({ default: module.MarketDashboard })));
const StrategyDiscoveryPage = lazy(() => import('./features/strategy/index').then((module) => ({ default: module.StrategyDiscoveryPage })));
const ExperimentDashboard = lazy(() => import('./features/experiment/index').then((module) => ({ default: module.ExperimentDashboard })));
const NewsCrawlerDashboard = lazy(() => import('./features/news/index').then((module) => ({ default: module.NewsCrawlerDashboard })));

type Tab = 'charts' | 'leaderboard' | 'builder' | 'news';
const navigation: Array<{ id: Tab; label: string; caption: string; icon: ComponentType<{ size?: number }> }> = [
  { id: 'charts', label: 'Market', caption: 'Realtime charts', icon: ChartCandlestick },
  { id: 'builder', label: 'Strategies', caption: 'Build & discover', icon: FlaskConical },
  { id: 'leaderboard', label: 'Backtests', caption: 'Results & ranking', icon: Trophy },
  { id: 'news', label: 'Market news', caption: 'Sentiment feed', icon: Newspaper },
];

function AuthenticatedApp() {
  const mode = useAppMode();
  const [activeTab, setActiveTab] = useState<Tab>('charts');
  const activeItem = navigation.find((item) => item.id === activeTab) ?? navigation[0];
  useEffect(() => {
    if (mode !== 'LIVE') return;
    wsManager.connect();
    return () => wsManager.disconnect();
  }, [mode]);

  return (
    <ErrorBoundary>
      <div className="app-shell">
        <aside className="app-sidebar">
          <div className="brand">
            <div className="brand-mark"><BarChart3 size={22} /></div>
            <div><strong>Crypto Strategy</strong><span>Research Lab</span></div>
          </div>
          <nav className="sidebar-nav" aria-label="Primary navigation">
            <span className="nav-eyebrow">Workspace</span>
            {navigation.map(({ id, label, caption, icon: Icon }) => (
              <button key={id} className={activeTab === id ? 'nav-item active' : 'nav-item'} onClick={() => setActiveTab(id)}>
                <Icon size={19} /><span><strong>{label}</strong><small>{caption}</small></span>
              </button>
            ))}
          </nav>
          <div className="sidebar-meta">
            <div className="market-status"><span className="status-dot" /> Market data</div>
            <span>Binance API + WebSocket</span><small>Version 1.0.0</small>
          </div>
        </aside>
        <section className="app-workspace">
          <header className="workspace-header">
            <div><span className="page-kicker">Trading workspace</span><h1>{activeItem.label}</h1></div>
            <div className="workspace-actions">
              <div className="live-pill"><Radio size={15} /> {mode === 'LIVE' ? 'Live infrastructure' : 'Offline demo'}</div>
              <ThemeToggle />
            </div>
          </header>
          <WebSocketStateBanner />
          <main className="app-main">
            <Suspense fallback={<div className="screen-loading">Loading workspace…</div>}>
              {activeTab === 'charts' && <MarketDashboard />}
              {activeTab === 'builder' && <StrategyDiscoveryPage />}
              {activeTab === 'leaderboard' && <ExperimentDashboard />}
              {activeTab === 'news' && <NewsCrawlerDashboard />}
            </Suspense>
          </main>
        </section>
      </div>
    </ErrorBoundary>
  );
}
function App() {
  return <ThemeProvider><AuthGate><AuthenticatedApp /></AuthGate></ThemeProvider>;
}
export default App;
