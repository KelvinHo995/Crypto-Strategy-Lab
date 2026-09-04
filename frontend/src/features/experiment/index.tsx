import { useCallback, useEffect, useRef, useState } from 'react';
import { BacktestConfigPanel } from './components/BacktestConfigPanel';
import { ExperimentLeaderboard } from './components/ExperimentLeaderboard';
import { ProvenanceModal } from './components/ProvenanceModal';
import { PerformanceSummaryCard } from './components/PerformanceSummaryCard';
import { TradeHistoryTable } from './components/TradeHistoryTable';
import { MOCK_EXPERIMENTS, generateMockTrades } from './services/mockExperimentData';
import { fetchExperiment, fetchExperiments, startSearch } from '../../shared/api';
import { useWebSocketSubscription } from '../../shared/hooks';
import type { WSSearchProgressPayload } from '../../types/websocket';
import type { ExperimentResult, StrategyInstance } from '../../types/backtest';
import { useExperimentStore, formatExperimentTitle } from '../../shared/stores/useExperimentStore';

// Matches the worker's own worst-case retry/backoff window (up to 3 attempts,
// backoff climbing toward 30s each) — a shorter timeout would give up on a
// backtest that's genuinely still retrying, not stuck.
const RUN_TIMEOUT_MS = 90_000;

export function ExperimentDashboard() {
  const [experiments, setExperiments] = useState<ExperimentResult[]>(MOCK_EXPERIMENTS);
  const [selectedExpForMeta, setSelectedExpForMeta] = useState<ExperimentResult | null>(null);

  const activeExp = useExperimentStore((state) => state.activeExperiment) ?? MOCK_EXPERIMENTS[0];
  const activeTrades = useExperimentStore((state) => state.activeTrades);
  const setActiveExp = useExperimentStore((state) => state.setActiveExperiment);
  const setActiveTrades = useExperimentStore((state) => state.setActiveTrades);
  const loadToChart = useExperimentStore((state) => state.loadToChart);
  const replicateToBuilder = useExperimentStore((state) => state.replicateToBuilder);

  const [isLoading, setIsLoading] = useState(false);
  const activeSearchId = useRef<string | null>(null);
  const runTimeout = useRef<number | null>(null);

  useEffect(() => {
    if (activeTrades.length === 0 && activeExp) {
      setActiveTrades(generateMockTrades(activeExp.id, activeExp.tradeCount || 30));
    }
  }, [activeExp, activeTrades.length, setActiveTrades]);

  const applyExperiments = useCallback((items: ExperimentResult[]) => {
    if (items.length === 0) return;
    setExperiments(items);
    if (!useExperimentStore.getState().activeExperiment) {
      setActiveExp(items[0]);
      setActiveTrades(generateMockTrades(items[0].id, items[0].tradeCount || 30));
    }
  }, [setActiveExp, setActiveTrades]);

  useEffect(() => {
    fetchExperiments().then(applyExperiments).catch(() => undefined);
  }, [applyExperiments]);

  // Fires once the run this component started reaches a terminal status —
  // driven by the real SEARCH_PROGRESS event, not a fixed-attempt REST poll,
  // so a slow (retrying) backtest is still tracked instead of silently
  // timing out after a few seconds.
  const finishRun = useCallback(async (id: string, status: 'COMPLETED' | 'FAILED' | 'STOPPED') => {
    if (runTimeout.current !== null) {
      window.clearTimeout(runTimeout.current);
      runTimeout.current = null;
    }
    activeSearchId.current = null;
    setIsLoading(false);
    if (status !== 'COMPLETED') {
      alert(`Backtest kết thúc: ${status}.`);
      return;
    }
    try {
      const result = await fetchExperiment(id);
      setActiveExp(result);
      setActiveTrades(generateMockTrades(result.id, result.tradeCount || 30));
      setExperiments(current => current.some(item => item.id === result.id)
        ? current.map(item => item.id === result.id ? result : item)
        : [result, ...current]);
    } catch {
      alert('Backtest đã hoàn tất nhưng không tải được kết quả — kiểm tra leaderboard.');
    }
  }, [setActiveExp, setActiveTrades]);

  const handleProgress = useCallback((progress: WSSearchProgressPayload) => {
    if (!activeSearchId.current || progress.searchId !== activeSearchId.current) return;
    if (progress.status === 'FAILED' || progress.status === 'STOPPED') {
      void finishRun(activeSearchId.current, progress.status);
    } else if (progress.tested >= progress.total) {
      void finishRun(activeSearchId.current, 'COMPLETED');
    }
  }, [finishRun]);

  useWebSocketSubscription<WSSearchProgressPayload>('SEARCH_PROGRESS', handleProgress);
  useWebSocketSubscription<ExperimentResult[]>('LEADERBOARD_UPDATE', applyExperiments);

  const handleRunBacktest = async (config: {
    symbol: string;
    timeframe: string;
    fromDate: string;
    toDate: string;
    capital: number;
    fee: number;
    slippage: number;
    instances: StrategyInstance[];
    policy?: 'majority' | 'weighted';
  }) => {
    setIsLoading(true);
    try {
      const started = await startSearch({
        pair: config.symbol,
        timeframe: config.timeframe,
        from: new Date(config.fromDate).getTime(),
        to: new Date(config.toDate).getTime(),
        capital: config.capital,
        fee: config.fee,
        slippage: config.slippage,
        instances: config.instances,
        policy: config.policy ?? 'majority',
      });
      activeSearchId.current = started.searchId;
      runTimeout.current = window.setTimeout(() => {
        if (activeSearchId.current !== started.searchId) return;
        activeSearchId.current = null;
        setIsLoading(false);
        alert('Backtest đang chạy lâu hơn bình thường (>90s) — có thể vẫn hoàn tất, kiểm tra leaderboard sau ít phút.');
      }, RUN_TIMEOUT_MS);
    } catch (error) {
      setIsLoading(false);
      alert(`Không thể chạy backtest thật: ${String(error)}. Hãy chạy migration/backfill trước.`);
    }
  };

  const handleSelectExperimentForMeta = (exp: ExperimentResult) => {
    setSelectedExpForMeta(exp);
  };

  const handleLoadToChart = (exp: ExperimentResult) => {
    const trades = generateMockTrades(exp.id, exp.tradeCount || 30);
    loadToChart(exp, trades);
  };

  const handleReplicate = (exp: ExperimentResult) => {
    setSelectedExpForMeta(null);
    const policy = exp.policy === 'weighted' ? 'weighted' : 'majority';
    replicateToBuilder(exp.instances, policy);
  };

  return (
    <div style={dashboardContainerStyle}>
      <div style={{color:'#f59e0b',fontSize:'0.75rem'}}>Metrics/leaderboard ưu tiên API; trade detail dùng demo vì backend MVP chưa lưu từng trade.</div>
      {/* 1. Top Section: Run Simulation & Performance Overview */}
      <div style={topSectionStyle}>
        <div style={configColStyle}>
          <BacktestConfigPanel onRunBacktest={handleRunBacktest} isLoading={isLoading} />
        </div>
        
        {activeExp && (
          <div style={summaryColStyle}>
            <div style={summaryHeaderStyle}>
              <h4>Active Simulation Summary: <span style={activeIdStyle}>{activeExp.id}</span></h4>
              <span style={activeStrategiesStyle}>{formatExperimentTitle(activeExp)}</span>
            </div>
            <PerformanceSummaryCard
              profit={activeExp.totalProfit}
              returnPct={activeExp.return}
              winRate={activeExp.winRate}
              mdd={activeExp.mdd}
              tradeCount={activeExp.tradeCount}
              wins={activeExp.wins}
              losses={activeExp.losses}
            />
          </div>
        )}
      </div>

      {/* 2. Middle Section: Leaderboard */}
      <div style={leaderboardSectionStyle}>
        <h3 style={sectionTitleStyle}>Strategy Experiment Leaderboard</h3>
        <ExperimentLeaderboard
          experiments={experiments}
          onSelectExperiment={handleSelectExperimentForMeta}
          onLoadToChart={handleLoadToChart}
        />
      </div>

      {/* 3. Bottom Section: Trade History log */}
      {activeTrades.length > 0 && (
        <div style={tradesSectionStyle}>
          <TradeHistoryTable trades={activeTrades} />
        </div>
      )}

      {/* Provenance Metadata Drawer */}
      <ProvenanceModal
        experiment={selectedExpForMeta}
        onClose={() => setSelectedExpForMeta(null)}
        onReplicate={handleReplicate}
      />
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const dashboardContainerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1.5rem',
  width: '100%',
  boxSizing: 'border-box',
};

const topSectionStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1.25rem',
  alignItems: 'stretch',
  flexWrap: 'wrap',
};

const configColStyle: React.CSSProperties = {
  flex: '1 1 350px',
};

const summaryColStyle: React.CSSProperties = {
  flex: '2 1 600px',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1.25rem',
  boxSizing: 'border-box',
};

const summaryHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
  marginBottom: '0.5rem',
};

const activeIdStyle: React.CSSProperties = {
  color: '#2563eb',
  fontFamily: 'monospace',
  fontWeight: '700',
};

const activeStrategiesStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '600',
  color: '#94a3b8',
};

const leaderboardSectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
};

const sectionTitleStyle: React.CSSProperties = {
  fontSize: '0.9rem',
  fontWeight: '700',
  color: '#0f172a',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const tradesSectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
};
