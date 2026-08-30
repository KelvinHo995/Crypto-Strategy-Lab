import { useCallback, useEffect, useState } from 'react';
import { BacktestConfigPanel } from './components/BacktestConfigPanel';
import { ExperimentLeaderboard } from './components/ExperimentLeaderboard';
import { ProvenanceModal } from './components/ProvenanceModal';
import { PerformanceSummaryCard } from './components/PerformanceSummaryCard';
import { TradeHistoryTable } from './components/TradeHistoryTable';
import { MOCK_EXPERIMENTS, generateMockTrades } from './services/mockExperimentData';
import type { ExperimentResult, Trade } from '../../types/backtest';
import { fetchExperiments, startSearch } from '../../shared/api';
import { useWebSocketSubscription } from '../../shared/hooks';

export function ExperimentDashboard() {
  const [experiments, setExperiments] = useState<ExperimentResult[]>(MOCK_EXPERIMENTS);
  const [selectedExpForMeta, setSelectedExpForMeta] = useState<ExperimentResult | null>(null);
  const [activeExp, setActiveExp] = useState<ExperimentResult | null>(MOCK_EXPERIMENTS[0]); // Default load first
  const [activeTrades, setActiveTrades] = useState<Trade[]>(
    MOCK_EXPERIMENTS[0] ? generateMockTrades(MOCK_EXPERIMENTS[0].id, MOCK_EXPERIMENTS[0].tradeCount) : []
  );
  
  const [isLoading, setIsLoading] = useState(false);

  const applyExperiments = useCallback((items: ExperimentResult[]) => {
    if (items.length === 0) return;
    setExperiments(items);
    setActiveExp(current => current ? items.find(item => item.id === current.id) ?? items[0] : items[0]);
  }, []);

  useEffect(() => {
    fetchExperiments().then(applyExperiments).catch(() => undefined);
  }, [applyExperiments]);

  useWebSocketSubscription<ExperimentResult[]>('LEADERBOARD_UPDATE', applyExperiments);

  const handleRunBacktest = async (config: {
    symbol: string;
    timeframe: string;
    fromDate: string;
    toDate: string;
    capital: number;
    fee: number;
  }) => {
    setIsLoading(true);
    try {
      const started = await startSearch({
        pair: config.symbol,
        timeframe: config.timeframe,
        from: new Date(config.fromDate).getTime(),
        to: new Date(config.toDate).getTime(),
        capital: config.capital,
        strategies: ['MA'],
      });
      for (let attempt = 0; attempt < 15; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 500));
        const items = await fetchExperiments();
        applyExperiments(items);
        const result = items.find(item => item.id === started.searchId);
        if (result?.status === 'COMPLETED' || result?.status === 'FAILED') {
          if (result.status === 'COMPLETED') {
            setActiveExp(result);
            setActiveTrades(generateMockTrades(result.id, result.tradeCount));
          }
          break;
        }
      }
    } catch (error) {
      alert(`Không thể chạy backtest thật: ${String(error)}. Hãy chạy migration/backfill trước.`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSelectExperimentForMeta = (exp: ExperimentResult) => {
    setSelectedExpForMeta(exp);
  };

  const handleLoadToChart = (exp: ExperimentResult) => {
    setActiveExp(exp);
    setActiveTrades(generateMockTrades(exp.id, exp.tradeCount));
    alert(`Loaded Strategy parameters from #${exp.id} to Chart Canvas and Performance Summary.`);
  };

  const handleReplicate = (exp: ExperimentResult) => {
    setSelectedExpForMeta(null);
    alert(`Replicated Strategy combination [${exp.strategies.join(' + ')}] into Builder state!`);
    // In production, this would sync with a global strategy builder state/store
  };

  const handleHoverTrade = (trade: Trade | null) => {
    if (trade) {
      console.log(`Hovering Trade: Exit Price: $${trade.exitPrice}, Profit: $${trade.profit}`);
      // This hook is wired up to dispatch highlighting coordinates to the active chart series in production
    }
  };

  const handleClickTrade = (trade: Trade) => {
    console.log(`Clicked Trade: Entry Price: $${trade.entryPrice}, Exit Price: $${trade.exitPrice}`);
    // Zoom/Scroll the active chart timeframe to this trade's entryTime
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
              <span style={activeStrategiesStyle}>{activeExp.strategies.join(' + ')}</span>
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
          <TradeHistoryTable
            trades={activeTrades}
            onHoverTrade={handleHoverTrade}
            onClickTrade={handleClickTrade}
          />
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
