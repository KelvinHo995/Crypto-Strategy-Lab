import { useState } from 'react';
import { BacktestConfigPanel } from './components/BacktestConfigPanel';
import { ExperimentLeaderboard } from './components/ExperimentLeaderboard';
import { ProvenanceModal } from './components/ProvenanceModal';
import { PerformanceSummaryCard } from './components/PerformanceSummaryCard';
import { TradeHistoryTable } from './components/TradeHistoryTable';
import { MOCK_EXPERIMENTS, generateMockTrades } from './services/mockExperimentData';
import type { ExperimentResult, Trade } from '../../types/backtest';

export function ExperimentDashboard() {
  const [experiments, setExperiments] = useState<ExperimentResult[]>(MOCK_EXPERIMENTS);
  const [selectedExpForMeta, setSelectedExpForMeta] = useState<ExperimentResult | null>(null);
  const [activeExp, setActiveExp] = useState<ExperimentResult | null>(MOCK_EXPERIMENTS[0]); // Default load first
  const [activeTrades, setActiveTrades] = useState<Trade[]>(
    MOCK_EXPERIMENTS[0] ? generateMockTrades(MOCK_EXPERIMENTS[0].id, MOCK_EXPERIMENTS[0].tradeCount) : []
  );
  
  const [isLoading, setIsLoading] = useState(false);

  const handleRunBacktest = (config: {
    symbol: string;
    timeframe: string;
    fromDate: string;
    toDate: string;
    capital: number;
    fee: number;
  }) => {
    setIsLoading(true);

    // Simulate backend backtest run
    setTimeout(() => {
      const newId = `exp-${Date.now().toString().slice(-4)}`;
      const returnPct = Number((Math.random() * 80 - 25).toFixed(2)); // Random -25% to +55%
      const mdd = Number((Math.random() * 25 + 3).toFixed(2)); // Random 3% to 28%
      const tradeCount = Math.floor(Math.random() * 80) + 15;
      const wins = Math.floor(tradeCount * (Math.random() * 0.3 + 0.45)); // 45% - 75% winrate
      const losses = tradeCount - wins;
      const winRate = Number(((wins / tradeCount) * 100).toFixed(2));
      const totalProfit = Number((config.capital * (returnPct / 100)).toFixed(2));

      const newExp: ExperimentResult = {
        id: newId,
        candidateId: `cand-${Date.now().toString().slice(-3)}`,
        strategies: ['MA', 'RSI'], // Constituent strategy placeholder
        params: { maWindow: 20, rsiPeriod: 14, capital: config.capital, fee: config.fee },
        policy: 'weighted',
        strategyVersions: { MA: 'v1.2.0', RSI: 'v2.0.1' },
        datasetPeriod: `${config.fromDate} to ${config.toDate}`,
        return: returnPct,
        mdd,
        tradeCount,
        winRate,
        wins,
        losses,
        totalProfit,
        status: 'COMPLETED',
        createdAt: Date.now(),
      };

      setExperiments((prev) => [newExp, ...prev]);
      setActiveExp(newExp);
      setActiveTrades(generateMockTrades(newId, tradeCount));
      setIsLoading(false);
      
      alert(`🎉 Backtest #${newId} Completed successfully!\nReturn: ${returnPct > 0 ? '+' : ''}${returnPct}%\nTrades: ${tradeCount}`);
    }, 2000);
  };

  const handleSelectExperimentForMeta = (exp: ExperimentResult) => {
    setSelectedExpForMeta(exp);
  };

  const handleLoadToChart = (exp: ExperimentResult) => {
    setActiveExp(exp);
    setActiveTrades(generateMockTrades(exp.id, exp.tradeCount));
    alert(`📈 Loaded Strategy parameters from #${exp.id} to Chart Canvas and Performance Summary.`);
  };

  const handleReplicate = (exp: ExperimentResult) => {
    setSelectedExpForMeta(null);
    alert(`⚙️ Replicated Strategy combination [${exp.strategies.join(' + ')}] into Builder state!`);
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
        <h3 style={sectionTitleStyle}>🏆 Strategy Experiment Leaderboard</h3>
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
  backgroundColor: '#0f172a',
  border: '1px solid #1e293b',
  borderRadius: '8px',
  padding: '1.25rem',
  boxSizing: 'border-box',
};

const summaryHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.5rem',
  marginBottom: '0.5rem',
};

const activeIdStyle: React.CSSProperties = {
  color: '#06b6d4',
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
  color: '#e2e8f0',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const tradesSectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
};
