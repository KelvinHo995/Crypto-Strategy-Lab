import { useState } from 'react';
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

export function StrategyDiscoveryPage() {
  const [instances, setInstances] = useState<SingleStrategyInstance[]>(DEFAULT_SINGLE_STRATEGIES);
  const [stats, setStats] = useState<DiscoveryStats>(DEFAULT_DISCOVERY_STATS);
  const [miniLeaderboard] = useState(MINI_LEADERBOARD_DATA);

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

  const handleStartBacktest = (config: {
    strategies: string[];
    weights: Record<string, number>;
    policy: 'majority' | 'weighted';
  }) => {
    // Generate description list
    const activeNames = config.strategies
      .map(id => instances.find(inst => inst.id === id)?.name || id)
      .join(' + ');

    alert(`🚀 Backtesting initiated on Go Backend!\nStrategy: ${activeNames}\nPolicy: ${config.policy}\n\nActivating Loop Discovery Panel...`);
    
    // Reset and trigger simulated loop discovery panel
    setStats({
      iteration: 0,
      totalIterations: 200,
      testedCandidates: 0,
      status: 'RUNNING',
      bestStrategy: {
        name: activeNames + (config.policy === 'weighted' ? ' (Weighted)' : ' (Majority)'),
        profit: Number((Math.random() * 1500 + 800).toFixed(2)),
        winrate: Number((Math.random() * 15 + 55).toFixed(2)),
        mdd: Number((Math.random() * 8 + 6).toFixed(2)),
      }
    });
  };

  return (
    <div style={pageContainerStyle}>
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
