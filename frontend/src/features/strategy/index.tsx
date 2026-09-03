import { useCallback, useEffect, useRef, useState } from 'react';
import { SingleStrategyList } from './components/SingleStrategyList';
import { CompositeStrategyBuilder } from './components/CompositeStrategyBuilder';
import { LoopDiscoveryPanel, type DiscoveryLoopConfig } from './components/LoopDiscoveryPanel';
import { MiniLeaderboard } from './components/MiniLeaderboard';
import {
  DEFAULT_SINGLE_STRATEGIES,
  DEFAULT_DISCOVERY_STATS,
  MINI_LEADERBOARD_DATA,
  type SingleStrategyInstance,
  type DiscoveryStats,
} from './services/mockStrategyData';
import { ApiError, fetchMarkets, fetchStrategies, startSearch, startSearchLoop } from '../../shared/api';
import { useWebSocketSubscription } from '../../shared/hooks';
import type { WSSearchProgressPayload } from '../../types/websocket';
import type { ExperimentResult, StrategyInstance } from '../../types/backtest';
import type { MarketInfo } from '../../types/candle';
import { DEFAULT_MARKETS } from '../market/services/marketCatalog';
import { useAppMode } from '../../shared/auth';
import { applyProgress } from './services/discoveryProgress';

export function StrategyDiscoveryPage() {
  const mode = useAppMode();
  const [markets, setMarkets] = useState<MarketInfo[]>(DEFAULT_MARKETS);
  const [symbol, setSymbol] = useState('BTCUSDT');
  const [instances, setInstances] = useState<SingleStrategyInstance[]>(DEFAULT_SINGLE_STRATEGIES);
  const [stats, setStats] = useState<DiscoveryStats>(DEFAULT_DISCOVERY_STATS);
  const [liveLeaderboard, setLiveLeaderboard] = useState<typeof MINI_LEADERBOARD_DATA>([]);
  const miniLeaderboard = mode === 'DEMO' ? MINI_LEADERBOARD_DATA : liveLeaderboard;
  const activeSearchId = useRef<string | null>(null);

  useEffect(() => {
    if (mode !== 'LIVE') return;
    fetchStrategies().then(names => {
      setInstances(current => current.filter(instance => names.includes(instance.type)));
    }).catch(() => undefined);
    fetchMarkets().then(setMarkets).catch(() => setMarkets(DEFAULT_MARKETS));
  }, [mode]);

  const handleProgress = useCallback((progress: WSSearchProgressPayload) => {
    setStats(current => applyProgress(current, progress, activeSearchId.current));
  }, []);

  const handleLeaderboard = useCallback((results: ExperimentResult[]) => {
    if (mode !== 'LIVE' || !Array.isArray(results)) return;
    const completed = results.filter(result => result.status === 'COMPLETED').slice(0, 5);
    setLiveLeaderboard(completed.map((result, index) => ({
      rank: index + 1,
      name: result.instances.map(i => i.type).join(' + ') || result.candidateId,
      profit: result.totalProfit,
      winrate: result.winRate,
    })));
    if (completed[0]) {
      const best = completed[0];
      setStats(current => ({
        ...current,
        bestStrategy: {
          name: best.instances.map(i => i.type).join(' + ') || best.candidateId,
          profit: best.totalProfit,
          winrate: best.winRate,
          mdd: best.mdd,
        },
      }));
    }
  }, [mode]);

  useWebSocketSubscription<WSSearchProgressPayload>('SEARCH_PROGRESS', handleProgress);
  useWebSocketSubscription<ExperimentResult[]>('LEADERBOARD_UPDATE', handleLeaderboard);

  const handleCreateInstance = (newInstance: SingleStrategyInstance) => {
    setInstances((prev) => [...prev, newInstance]);
  };

  const handleStartBacktest = async (config: {
    instances: StrategyInstance[];
    policy: 'majority' | 'weighted';
  }) => {
    if (config.instances.length === 0) return;
    const activeNames = config.instances.map(inst => inst.type).join(' + ');
    try {
      await startSearch({
        pair: symbol, timeframe: '1h', from: Date.now() - 180 * 86400000, to: Date.now(), capital: 10000,
        instances: config.instances, policy: config.policy,
      });
      alert(`Đã gửi backtest thật: ${activeNames}. Theo dõi tiến độ qua WebSocket.`);
    } catch (error) {
      alert(`Không thể gửi backtest: ${String(error)}. Cần migration và backfill trước.`);
    }
  };

  const handleStartLoop = async (config: DiscoveryLoopConfig) => {
    if (mode !== 'LIVE') {
      setStats({ ...DEFAULT_DISCOVERY_STATS, status: 'FAILED', statusMessage: 'Đăng nhập LIVE mode để chạy Search Loop thật.' });
      return;
    }
    if (config.maxCandidates < 2 || config.maxCandidates > 200) {
      setStats(current => ({ ...current, status: 'FAILED', statusMessage: 'Candidates phải nằm trong khoảng 2–200.' }));
      return;
    }
    if (config.maxDurationSeconds < 60 || config.maxDurationSeconds > 3600) {
      setStats(current => ({ ...current, status: 'FAILED', statusMessage: 'Max duration phải nằm trong khoảng 60–3600 giây.' }));
      return;
    }
    if (config.noImprovementLimit < 0) {
      setStats(current => ({ ...current, status: 'FAILED', statusMessage: 'No-improvement limit không được âm.' }));
      return;
    }

    activeSearchId.current = null;
    setStats({
      ...DEFAULT_DISCOVERY_STATS,
      totalIterations: config.maxCandidates,
      status: 'RUNNING',
      statusMessage: 'Đã gửi Search Loop, đang chờ kết quả worker qua WebSocket…',
    });
    const to = Date.now();
    try {
      const response = await startSearchLoop({
        pair: symbol,
        timeframe: config.timeframe,
        from: to - 180 * 86400000,
        to,
        capital: 10000,
        maxCandidates: config.maxCandidates,
        maxDurationSeconds: config.maxDurationSeconds,
        noImprovementLimit: config.noImprovementLimit,
      });
      activeSearchId.current = response.searchId;
      setStats(current => ({
        ...current,
        searchId: response.searchId,
        totalIterations: response.maxCandidates,
        statusMessage: current.status === 'RUNNING'
          ? 'Search Loop đang chạy; số liệu bên dưới đến trực tiếp từ backend.'
          : current.statusMessage,
      }));
    } catch (error) {
      activeSearchId.current = null;
      const message = error instanceof ApiError && error.status === 422
        ? 'Không đủ 21 candle trong khoảng đã chọn. Hãy chạy backfill cho market/timeframe này rồi thử lại.'
        : `Không thể bắt đầu Search Loop: ${error instanceof Error ? error.message : String(error)}`;
      setStats(current => ({ ...current, status: 'FAILED', statusMessage: message }));
    }
  };

  const handleResetLoop = () => {
    activeSearchId.current = null;
    setStats(DEFAULT_DISCOVERY_STATS);
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
          onStart={handleStartLoop}
          onReset={handleResetLoop}
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
