import { useCallback, useEffect, useRef, useState } from 'react';
import { BacktestConfigPanel } from './components/BacktestConfigPanel';
import { ExperimentLeaderboard } from './components/ExperimentLeaderboard';
import { ProvenanceModal } from './components/ProvenanceModal';
import { PerformanceSummaryCard } from './components/PerformanceSummaryCard';
import { TradeHistoryTable } from './components/TradeHistoryTable';
import { MOCK_EXPERIMENTS, generateMockTrades } from './services/mockExperimentData';
import { fetchExperiment, fetchExperiments, fetchTrades, startSearch } from '../../shared/api';
import { useWebSocketSubscription } from '../../shared/hooks';
import type { WSSearchProgressPayload } from '../../types/websocket';
import type { ExperimentResult, Trade } from '../../types/backtest';
import { useExperimentStore, formatExperimentTitle } from '../../shared/stores/useExperimentStore';

// Matches the worker's own worst-case retry/backoff window (up to 3 attempts,
// backoff climbing toward 30s each) — a shorter timeout would give up on a
// backtest that's genuinely still retrying, not stuck.
const RUN_TIMEOUT_MS = 90_000;

const isMockExperiment = (id: string) => MOCK_EXPERIMENTS.some(m => m.id === id);

// MOCK_EXPERIMENTS are fixture data end to end — generating consistent fake
// trades for those specific rows is fine, it's already clearly demo data.
// Anything else is a real experiment, so its trades come from the real
// per-trade history the worker now persists (empty if this run predates
// that, or genuinely has none — never backfilled with fake ones).
async function loadTradesFor(exp: ExperimentResult): Promise<Trade[]> {
  if (isMockExperiment(exp.id)) {
    return generateMockTrades(exp.id, exp.tradeCount || 30);
  }
  try {
    return await fetchTrades(exp.id);
  } catch {
    return [];
  }
}

export function ExperimentDashboard() {
  const [experiments, setExperiments] = useState<ExperimentResult[]>(MOCK_EXPERIMENTS);
  const [selectedExpForMeta, setSelectedExpForMeta] = useState<ExperimentResult | null>(null);

  const activeExp = useExperimentStore((state) => state.activeExperiment) ?? MOCK_EXPERIMENTS[0];
  const activeTrades = useExperimentStore((state) => state.activeTrades);
  const setActiveExp = useExperimentStore((state) => state.setActiveExperiment);
  const setActiveTrades = useExperimentStore((state) => state.setActiveTrades);
  const loadExperimentToChart = useExperimentStore((state) => state.loadExperimentToChart);

  const [isLoading, setIsLoading] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const toastTimeout = useRef<number | null>(null);
  const activeSearchId = useRef<string | null>(null);
  const runTimeout = useRef<number | null>(null);

  const showToast = useCallback((message: string) => {
    if (toastTimeout.current !== null) window.clearTimeout(toastTimeout.current);
    setToast(message);
    toastTimeout.current = window.setTimeout(() => setToast(null), 4000);
  }, []);

  useEffect(() => {
    if (activeTrades.length === 0 && activeExp) {
      loadTradesFor(activeExp).then(setActiveTrades);
    }
  }, [activeExp, activeTrades.length, setActiveTrades]);

  const applyExperiments = useCallback((items: ExperimentResult[]) => {
    if (items.length === 0) return;
    setExperiments(items);
    if (!useExperimentStore.getState().activeExperiment) {
      setActiveExp(items[0]);
      loadTradesFor(items[0]).then(setActiveTrades);
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
      loadTradesFor(result).then(setActiveTrades);
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
        instances: [{ type: 'MA' }],
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

  const handleLoadToChart = async (exp: ExperimentResult) => {
    const trades = await loadTradesFor(exp);
    loadExperimentToChart(exp, trades);
    showToast(`Loaded #${exp.id} (${formatExperimentTitle(exp)}) onto the chart — open the Market tab to see it.`);
  };

  const handleReplicate = (exp: ExperimentResult) => {
    setSelectedExpForMeta(null);
    showToast(`Replicated strategy combination [${formatExperimentTitle(exp)}] into Builder state.`);
    // In production, this would sync with a global strategy builder state/store
  };

  return (
    <div style={dashboardContainerStyle}>
      {toast && <div style={toastStyle}>{toast}</div>}
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
      {activeExp && (
        <div style={tradesSectionStyle}>
          {activeTrades.length > 0 ? (
            <TradeHistoryTable trades={activeTrades} />
          ) : activeExp.tradeCount > 0 ? (
            <p style={noTradesStyle}>
              This run reported {activeExp.tradeCount} trade{activeExp.tradeCount === 1 ? '' : 's'}, but predates
              per-trade history tracking — only the aggregate metrics above were kept. Re-run it to get real trade detail.
            </p>
          ) : (
            <p style={noTradesStyle}>This strategy never traded during the backtest window.</p>
          )}
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

const noTradesStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  color: '#64748b',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1rem',
  margin: 0,
};

const toastStyle: React.CSSProperties = {
  position: 'fixed',
  bottom: '1.5rem',
  right: '1.5rem',
  maxWidth: '360px',
  fontSize: '0.8rem',
  color: '#047857',
  backgroundColor: '#ecfdf5',
  border: '1px solid #a7f3d0',
  borderRadius: '8px',
  padding: '0.65rem 0.9rem',
  boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.2)',
  zIndex: 1000,
};
