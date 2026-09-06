import { useCallback, useEffect, useRef, useState } from 'react';
import { BacktestConfigPanel } from './components/BacktestConfigPanel';
import { ExperimentLeaderboard } from './components/ExperimentLeaderboard';
import { ProvenanceModal } from './components/ProvenanceModal';
import { PerformanceSummaryCard } from './components/PerformanceSummaryCard';
import { TradeHistoryTable } from './components/TradeHistoryTable';
import { BacktestChart } from './components/BacktestChart';
import { MOCK_EXPERIMENTS, generateMockTrades } from './services/mockExperimentData';
import { fetchExperiment, fetchExperiments, fetchTrades, startSearch } from '../../shared/api';
import { useWebSocketSubscription } from '../../shared/hooks';
import { useAppMode } from '../../shared/auth';
import type { WSSearchProgressPayload } from '../../types/websocket';
import type { ExperimentResult, Trade } from '../../types/backtest';
import { useExperimentStore, formatExperimentTitle } from '../../shared/stores/useExperimentStore';

// Matches the worker's own worst-case retry/backoff window (up to 3 attempts,
// backoff climbing toward 30s each) — a shorter timeout would give up on a
// backtest that's genuinely still retrying, not stuck.
const RUN_TIMEOUT_MS = 90_000;

// Fixture trades are only legal in explicit offline DEMO mode. LIVE mode
// always exposes API failures instead of silently replacing them with data
// that looks like a real run.
async function loadTradesFor(exp: ExperimentResult, mode: 'DEMO' | 'LIVE'): Promise<Trade[]> {
  if (mode === 'DEMO') {
    return generateMockTrades(exp.id, exp.tradeCount || 30);
  }
  return fetchTrades(exp.id);
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function ExperimentDashboard() {
  const mode = useAppMode();
  const [experiments, setExperiments] = useState<ExperimentResult[]>([]);
  const [leaderboardState, setLeaderboardState] = useState<'loading' | 'ready' | 'empty' | 'error'>('loading');
  const [leaderboardError, setLeaderboardError] = useState('');
  const [leaderboardReload, setLeaderboardReload] = useState(0);
  const [tradeState, setTradeState] = useState<'idle' | 'loading' | 'ready' | 'error'>('idle');
  const [tradeError, setTradeError] = useState('');
  const [selectedExpForMeta, setSelectedExpForMeta] = useState<ExperimentResult | null>(null);

  const activeExp = useExperimentStore((state) => state.activeExperiment);
  const activeTrades = useExperimentStore((state) => state.activeTrades);
  const setActiveExp = useExperimentStore((state) => state.setActiveExperiment);
  const setActiveTrades = useExperimentStore((state) => state.setActiveTrades);

  const [isLoading, setIsLoading] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const toastTimeout = useRef<number | null>(null);
  const activeSearchId = useRef<string | null>(null);
  const runTimeout = useRef<number | null>(null);
  const pollTimeout = useRef<number | null>(null);

  // Which single trade the chart below highlights — per spec (Trade Detail),
  // clicking a row highlights just that trade's own entry/exit, not every
  // trade in the run at once. Defaults to the first trade so the chart isn't
  // blank before the user clicks anything.
  const [selectedTrade, setSelectedTrade] = useState<Trade | null>(null);

  const showToast = useCallback((message: string) => {
    if (toastTimeout.current !== null) window.clearTimeout(toastTimeout.current);
    setToast(message);
    toastTimeout.current = window.setTimeout(() => setToast(null), 4000);
  }, []);

  const applyExperiments = useCallback((items: ExperimentResult[]) => {
    if (mode !== 'LIVE') return;
    setExperiments(items);
    setLeaderboardState(items.length > 0 ? 'ready' : 'empty');
    setLeaderboardError('');
    if (items.length === 0) {
      setActiveExp(null);
      setActiveTrades([]);
      return;
    }
    const current = useExperimentStore.getState().activeExperiment;
    if (!current || MOCK_EXPERIMENTS.some(mock => mock.id === current.id)) {
      setActiveExp(items[0]);
    }
  }, [mode, setActiveExp, setActiveTrades]);

  useEffect(() => {
    let cancelled = false;
    void Promise.resolve().then(async () => {
      if (cancelled) return;
      setSelectedExpForMeta(null);
      setActiveTrades([]);
      setTradeState('idle');
      setTradeError('');

      if (mode === 'DEMO') {
        setExperiments(MOCK_EXPERIMENTS);
        setLeaderboardState('ready');
        setLeaderboardError('');
        setActiveExp(MOCK_EXPERIMENTS[0]);
        return;
      }

      setExperiments([]);
      setActiveExp(null);
      setLeaderboardState('loading');
      setLeaderboardError('');
      try {
        const items = await fetchExperiments();
        if (!cancelled) applyExperiments(items);
      } catch (error) {
        if (cancelled) return;
        setExperiments([]);
        setLeaderboardState('error');
        setLeaderboardError(errorMessage(error));
      }
    });
    return () => { cancelled = true; };
  }, [applyExperiments, leaderboardReload, mode, setActiveExp, setActiveTrades]);

  useEffect(() => {
    let cancelled = false;
    void Promise.resolve().then(async () => {
      if (cancelled) return;
      setActiveTrades([]);
      setTradeError('');
      if (!activeExp) {
        setTradeState('idle');
        return;
      }
      setTradeState('loading');
      try {
        const trades = await loadTradesFor(activeExp, mode);
        if (cancelled) return;
        setActiveTrades(trades);
        setTradeState('ready');
      } catch (error) {
        if (cancelled) return;
        setActiveTrades([]);
        setTradeState('error');
        setTradeError(errorMessage(error));
      }
    });
    return () => { cancelled = true; };
  }, [activeExp, mode, setActiveTrades]);

  useEffect(() => {
    void Promise.resolve().then(() => setSelectedTrade(activeTrades[0] ?? null));
  }, [activeTrades]);

  const stopTrackingRun = useCallback(() => {
    if (runTimeout.current !== null) {
      window.clearTimeout(runTimeout.current);
      runTimeout.current = null;
    }
    if (pollTimeout.current !== null) {
      window.clearTimeout(pollTimeout.current);
      pollTimeout.current = null;
    }
    activeSearchId.current = null;
    setIsLoading(false);
  }, []);

  // WebSocket is the fast path. The REST read verifies the persisted terminal
  // status, preventing a FAILED candidate from being interpreted as COMPLETED
  // merely because tested == total.
  const finishRun = useCallback(async (id: string): Promise<boolean> => {
    try {
      const result = await fetchExperiment(id);
      if (result.status === 'PENDING' || result.status === 'RUNNING') return false;
      stopTrackingRun();
      setExperiments(current => current.some(item => item.id === result.id)
        ? current.map(item => item.id === result.id ? result : item)
        : [result, ...current]);
      setLeaderboardState('ready');
      if (result.status !== 'COMPLETED') {
        showToast(`Backtest ${result.id} failed. Open provenance or server logs for details.`);
        return true;
      }
      setActiveExp(result);
      showToast(`Backtest ${result.id} completed successfully.`);
      return true;
    } catch (error) {
      showToast(`Could not verify backtest status: ${errorMessage(error)}`);
      return false;
    }
  }, [setActiveExp, showToast, stopTrackingRun]);

  const scheduleStatusPolling = useCallback((id: string) => {
    if (pollTimeout.current !== null) window.clearTimeout(pollTimeout.current);
    const poll = async () => {
      if (activeSearchId.current !== id) return;
      const terminal = await finishRun(id);
      if (!terminal && activeSearchId.current === id) {
        pollTimeout.current = window.setTimeout(poll, 1500);
      }
    };
    pollTimeout.current = window.setTimeout(poll, 1000);
  }, [finishRun]);

  const handleProgress = useCallback((progress: WSSearchProgressPayload) => {
    if (!activeSearchId.current || progress.searchId !== activeSearchId.current) return;
    if (progress.status === 'FAILED' || progress.status === 'STOPPED' || progress.tested >= progress.total) {
      void finishRun(activeSearchId.current);
    }
  }, [finishRun]);

  useEffect(() => () => {
    if (toastTimeout.current !== null) window.clearTimeout(toastTimeout.current);
    if (runTimeout.current !== null) window.clearTimeout(runTimeout.current);
    if (pollTimeout.current !== null) window.clearTimeout(pollTimeout.current);
  }, []);

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
    if (mode === 'DEMO') {
      setActiveExp(MOCK_EXPERIMENTS[0]);
      showToast('Offline demo loaded a deterministic sample backtest.');
      return;
    }
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
      scheduleStatusPolling(started.searchId);
      runTimeout.current = window.setTimeout(() => {
        if (activeSearchId.current !== started.searchId) return;
        stopTrackingRun();
        showToast('Backtest is taking longer than 90 seconds. It may still finish; refresh the leaderboard shortly.');
      }, RUN_TIMEOUT_MS);
    } catch (error) {
      setIsLoading(false);
      setLeaderboardError(`Could not start backtest: ${errorMessage(error)}`);
      setLeaderboardState('error');
      showToast('Backtest could not start. Check migration, backfill and server connectivity.');
    }
  };

  const handleSelectExperimentForMeta = (exp: ExperimentResult) => {
    setSelectedExpForMeta(exp);
  };

  const handleLoadToChart = (exp: ExperimentResult) => {
    setActiveExp(exp);
    showToast(`Loading #${exp.id} (${formatExperimentTitle(exp)}) with its persisted trade history.`);
  };

  const handleReplicate = (exp: ExperimentResult) => {
    setSelectedExpForMeta(null);
    showToast(`Replicated strategy combination [${formatExperimentTitle(exp)}] into Builder state.`);
    // In production, this would sync with a global strategy builder state/store
  };

  return (
    <div style={dashboardContainerStyle}>
      {toast && <div style={toastStyle}>{toast}</div>}
      {mode === 'LIVE' && leaderboardState === 'error' && (
        <div style={errorPanelStyle} role="alert">
          <div><strong>Could not load live experiments.</strong> {leaderboardError}</div>
          <button type="button" onClick={() => setLeaderboardReload(value => value + 1)} style={retryButtonStyle}>Retry</button>
        </div>
      )}
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
        {leaderboardState === 'loading' && <p style={statePanelStyle}>Loading live experiments…</p>}
        {leaderboardState === 'empty' && <p style={statePanelStyle}>No live backtests yet. Run one above to create the first result.</p>}
        {experiments.length > 0 && (
          <ExperimentLeaderboard
            experiments={experiments}
            onSelectExperiment={handleSelectExperimentForMeta}
            onLoadToChart={handleLoadToChart}
          />
        )}
      </div>

      {/* 3. Bottom Section: Trade Chart + History log */}
      {activeExp && (
        <div style={tradesSectionStyle}>
          {tradeState === 'loading' ? (
            <p style={statePanelStyle}>Loading persisted trades…</p>
          ) : tradeState === 'error' ? (
            <p style={errorPanelStyle} role="alert">Could not load trade history: {tradeError}</p>
          ) : activeTrades.length > 0 ? (
            <>
              <BacktestChart experiment={activeExp} trades={activeTrades} highlightedTrade={selectedTrade} />
              <TradeHistoryTable
                trades={activeTrades}
                selectedTrade={selectedTrade}
                onClickTrade={setSelectedTrade}
              />
            </>
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
  gap: '1rem',
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

const statePanelStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  color: '#64748b',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1rem',
  margin: 0,
};

const errorPanelStyle: React.CSSProperties = {
  ...statePanelStyle,
  color: '#b91c1c',
  backgroundColor: '#fef2f2',
  borderColor: '#fecaca',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  gap: '1rem',
};

const retryButtonStyle: React.CSSProperties = {
  color: '#ffffff',
  backgroundColor: '#b91c1c',
  border: 0,
  borderRadius: '6px',
  padding: '0.4rem 0.75rem',
  cursor: 'pointer',
  fontWeight: 700,
};
