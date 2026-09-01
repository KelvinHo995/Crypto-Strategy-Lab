import { useEffect, useState } from 'react';
import { SingleStrategyList } from './components/SingleStrategyList';
import { CompositeStrategyBuilder } from './components/CompositeStrategyBuilder';
import { LoopDiscoveryPanel } from './components/LoopDiscoveryPanel';
import { MiniLeaderboard } from './components/MiniLeaderboard';
import {
  DEFAULT_SINGLE_STRATEGIES,
  DEFAULT_DISCOVERY_STATS,
  MINI_LEADERBOARD_DATA,
  type SingleStrategyInstance,
  type DiscoveryStats,
} from './services/mockStrategyData';
import { fetchMarkets, fetchStrategies, startSearch } from '../../shared/api';
import { useWebSocketSubscription } from '../../shared/hooks';
import type { WSSearchProgressPayload } from '../../types/websocket';
import type { MarketInfo } from '../../types/candle';
import { DEFAULT_MARKETS } from '../market/services/marketCatalog';
import { useAppMode } from '../../shared/auth';

export function StrategyDiscoveryPage() {
  const mode = useAppMode();
  const [markets, setMarkets] = useState<MarketInfo[]>(DEFAULT_MARKETS);
  const [symbol, setSymbol] = useState('BTCUSDT');
  const [instances, setInstances] = useState<SingleStrategyInstance[]>(DEFAULT_SINGLE_STRATEGIES);
  const [stats, setStats] = useState<DiscoveryStats>(DEFAULT_DISCOVERY_STATS);
  const [miniLeaderboard] = useState(MINI_LEADERBOARD_DATA);

  useEffect(() => {
    if (mode !== 'LIVE') return;
    fetchStrategies().then(names => {
      setInstances(current => current.filter(instance => names.includes(instance.type)));
    }).catch(() => undefined);
    fetchMarkets().then(setMarkets).catch(() => setMarkets(DEFAULT_MARKETS));
  }, [mode]);

  useWebSocketSubscription<WSSearchProgressPayload>('SEARCH_PROGRESS', progress => {
    setStats(current => ({
      ...current,
      iteration: progress.tested,
      totalIterations: progress.total,
      testedCandidates: progress.tested,
      status: progress.tested >= progress.total ? 'COMPLETED' : 'RUNNING',
    }));
  });

  const handleCreateInstance = (newInstance: SingleStrategyInstance) => {
    setInstances((prev) => [...prev, newInstance]);
  };

  const handleStatusChange = (status: 'IDLE' | 'RUNNING' | 'PAUSED' | 'COMPLETED') => {
    setStats((prev) => ({
      ...prev,
      status,
    }));
  };

  const handleUpdateIteration = (iteration: number, testedCandidates: number) => {
    setStats((prev) => ({
      ...prev,
      iteration,
      testedCandidates,
    }));
  };

  const handleStartBacktest = async (config: {
    strategies: string[];
    weights: Record<string, number>;
    policy: 'majority' | 'weighted';
  }) => {
    // Generate description list
    const activeNames = config.strategies
      .map(id => instances.find(inst => inst.id === id)?.name || id)
      .join(' + ');

    const strategyNames = [...new Set(config.strategies.map(id => instances.find(inst => inst.id === id)?.type).filter((name): name is string => Boolean(name)))];
    if (strategyNames.length === 0) return;
    setStats(current => ({ ...current, iteration: 0, totalIterations: 1, testedCandidates: 0, status: 'RUNNING' }));
    try {
      await startSearch({ pair: symbol, timeframe: '1h', from: Date.now() - 180 * 86400000, to: Date.now(), capital: 10000, strategies: strategyNames });
      alert(`Đã gửi backtest thật: ${activeNames}. Theo dõi tiến độ qua WebSocket.`);
    } catch (error) {
      setStats(current => ({ ...current, status: 'IDLE' }));
      alert(`Không thể gửi backtest: ${String(error)}. Cần migration và backfill trước.`);
    }
  };

  return (
    <div style={pageContainerStyle}>
      <div style={marketToolbarStyle}>
        <span>Discovery market</span>
        <select value={symbol} onChange={event => setSymbol(event.target.value)} style={marketSelectStyle}>
          {markets.map(market => <option key={market.symbol} value={market.symbol}>{market.baseAsset}/{market.quoteAsset}</option>)}
        </select>
      </div>
      {/* Cột 1: Danh sách Strategy đơn */}
      <div style={col1Style}>
        <SingleStrategyList
          instances={instances}
          onCreateInstance={handleCreateInstance}
        />
      </div>

      {/* Cột 2: Giao diện xây dựng Strategy kết hợp */}
      <div style={col2Style}>
        <CompositeStrategyBuilder
          singleInstances={instances}
          onStartBacktest={handleStartBacktest}
        />
      </div>

      {/* Cột 3: Quản lý vòng lặp khám phá & BXH rút gọn */}
      <div style={col3Style}>
        <LoopDiscoveryPanel
          stats={stats}
          onStatusChange={handleStatusChange}
          onUpdateIteration={handleUpdateIteration}
        />
        <MiniLeaderboard items={miniLeaderboard} />
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const pageContainerStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1rem',
  width: '100%',
  alignItems: 'stretch',
  flexWrap: 'wrap', // Wrap column on smaller screens
};

const marketToolbarStyle: React.CSSProperties = { width: '100%', display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: '0.6rem', color: '#475569', fontSize: '0.78rem' };
const marketSelectStyle: React.CSSProperties = { border: '1px solid #cbd5e1', borderRadius: '6px', background: '#fff', color: '#0f172a', padding: '0.35rem 0.5rem' };

const col1Style: React.CSSProperties = {
  width: '280px',
  flexShrink: 0,
  display: 'flex',
  flexDirection: 'column',
};

const col2Style: React.CSSProperties = {
  flex: '1 1 350px',
  display: 'flex',
  flexDirection: 'column',
};

const col3Style: React.CSSProperties = {
  width: '340px',
  flexShrink: 0,
  display: 'flex',
  flexDirection: 'column',
  gap: '1rem',
};
