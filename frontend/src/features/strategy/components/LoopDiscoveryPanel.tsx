import { useState, useEffect } from 'react';
import type { DiscoveryStats } from '../services/mockStrategyData';

interface LoopDiscoveryPanelProps {
  stats: DiscoveryStats;
  onStatusChange: (status: 'IDLE' | 'RUNNING' | 'PAUSED' | 'COMPLETED') => void;
  onUpdateIteration: (iter: number, tested: number) => void;
}

export function LoopDiscoveryPanel({
  stats,
  onStatusChange,
  onUpdateIteration,
}: LoopDiscoveryPanelProps) {
  const [searchMethod, setSearchMethod] = useState<'random' | 'domain' | 'genetic'>('random');

  // Simulated Loop Discovery updates when running
  useEffect(() => {
    if (stats.status !== 'RUNNING') return;

    const interval = setInterval(() => {
      if (stats.iteration >= stats.totalIterations) {
        onStatusChange('COMPLETED');
        clearInterval(interval);
        return;
      }
      
      const step = Math.floor(Math.random() * 3) + 1; // Step up 1 to 3 iterations
      const nextIter = Math.min(stats.iteration + step, stats.totalIterations);
      const nextTested = stats.testedCandidates + step * 50; // 50 backtests per iteration

      onUpdateIteration(nextIter, nextTested);
    }, 1500);

    return () => clearInterval(interval);
  }, [stats.status, stats.iteration, stats.totalIterations]);

  const handleStart = () => {
    if (stats.status === 'COMPLETED') {
      onUpdateIteration(0, 0); // Reset
    }
    onStatusChange('RUNNING');
  };

  const handlePause = () => {
    onStatusChange('PAUSED');
  };

  const handleReset = () => {
    onStatusChange('IDLE');
    onUpdateIteration(0, 0);
  };

  const progressPercent = (stats.iteration / stats.totalIterations) * 100;

  return (
    <div style={panelContainerStyle}>
      <h3 style={titleStyle}>Loop Discovery Engine</h3>

      {/* Visual Flow diagram of Discovery Loop */}
      <div style={flowDiagramStyle}>
        <div style={flowStepStyle} title="Generate Strategy variant parameters">Gen</div>
        <div style={flowArrowStyle}>➔</div>
        <div style={flowStepStyle} title="Run Backtest simulation">Backtest</div>
        <div style={flowArrowStyle}>➔</div>
        <div style={flowStepStyle} title="Evaluate metrics (PnL, Winrate)">Eval</div>
        <div style={flowArrowStyle}>➔</div>
        <div style={flowStepStyle} title="Rank and update Leaderboard">Rank</div>
      </div>

      {/* Discovery Search Algorithm Selector */}
      <div style={sectionStyle}>
        <h4 style={sectionHeaderStyle}>Discovery Method</h4>
        <div style={radioGroupStyle}>
          <label style={radioLabelStyle}>
            <input
              type="radio"
              name="method"
              checked={searchMethod === 'random'}
              onChange={() => setSearchMethod('random')}
              style={radioStyle}
            />
            <div>
              <span style={radioTitleStyle}>Random Search</span>
              <span style={radioDescStyle}>Brute-force parameter sweeps (Standard)</span>
            </div>
          </label>

          <label style={radioLabelStyle}>
            <input
              type="radio"
              name="method"
              checked={searchMethod === 'domain'}
              onChange={() => setSearchMethod('domain')}
              style={radioStyle}
            />
            <div>
              <span style={radioTitleStyle}>Domain-guided</span>
              <span style={radioDescStyle}>Searches within structural boundaries</span>
            </div>
          </label>

          <label style={radioLabelStyle}>
            <input
              type="radio"
              name="method"
              checked={searchMethod === 'genetic'}
              onChange={() => setSearchMethod('genetic')}
              style={radioStyle}
            />
            <div>
              <span style={radioTitleStyle}>Genetic Algorithm</span>
              <span style={radioDescStyle}>Evolves variants over multi-generations</span>
            </div>
          </label>
        </div>
      </div>

      {/* Discovery Progress Statistics */}
      <div style={progressSectionStyle}>
        <div style={progressHeaderStyle}>
          <h4 style={sectionHeaderStyle}>Execution Progress</h4>
          <span style={statusBadgeStyle(stats.status)}>{stats.status}</span>
        </div>

        {/* Dynamic Progress Bar */}
        <div style={progressBarContainerStyle}>
          <div style={{ ...progressBarFillStyle, width: `${progressPercent}%` }} />
        </div>

        <div style={statsGridStyle}>
          <div style={statBoxStyle}>
            <span style={statLabelStyle}>Iterations</span>
            <span style={statValueStyle}>{stats.iteration} / {stats.totalIterations}</span>
          </div>
          <div style={statBoxStyle}>
            <span style={statLabelStyle}>Candidates Checked</span>
            <span style={statValueStyle}>{stats.testedCandidates.toLocaleString()}</span>
          </div>
        </div>

        {/* Best Strategy So Far details */}
        {stats.iteration > 0 && (
          <div style={bestStrategyCardStyle}>
            <div style={bestHeaderStyle}>
              ⭐ Best Candidate So Far
            </div>
            <div style={bestNameStyle}>{stats.bestStrategy.name}</div>
            <div style={bestStatsGridStyle}>
              <div style={bestStatItemStyle}>
                <span style={bestLabelStyle}>Return:</span>
                <span style={bestProfitStyle}>+{stats.bestStrategy.profit.toLocaleString()} USDT</span>
              </div>
              <div style={bestStatItemStyle}>
                <span style={bestLabelStyle}>Winrate:</span>
                <span style={bestValStyle}>{stats.bestStrategy.winrate}%</span>
              </div>
              <div style={bestStatItemStyle}>
                <span style={bestLabelStyle}>MDD:</span>
                <span style={bestValStyle}>{stats.bestStrategy.mdd}%</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Control Buttons */}
      <div style={controlsContainerStyle}>
        {stats.status === 'RUNNING' ? (
          <button onClick={handlePause} style={pauseButtonStyle}>
            ⏸ Pause Discovery
          </button>
        ) : (
          <button onClick={handleStart} style={startButtonStyle}>
            {stats.status === 'COMPLETED' ? '🔁 Restart Loop' : '▶ Start Loop Discovery'}
          </button>
        )}
        <button onClick={handleReset} style={resetButtonStyle}>
          Reset
        </button>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const panelContainerStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  border: '1px solid #1e293b',
  borderRadius: '8px',
  padding: '1rem',
  height: '100%',
  boxSizing: 'border-box',
  display: 'flex',
  flexDirection: 'column',
  gap: '1rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.9rem',
  fontWeight: '700',
  color: '#e2e8f0',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.5rem',
};

const flowDiagramStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  backgroundColor: '#070a13',
  padding: '0.5rem 0.75rem',
  borderRadius: '6px',
  border: '1px solid #1e293b',
};

const flowStepStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  fontWeight: '700',
  color: '#06b6d4',
  backgroundColor: '#0f172a',
  border: '1px solid #334155',
  padding: '0.2rem 0.4rem',
  borderRadius: '4px',
  cursor: 'help',
};

const flowArrowStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#475569',
};

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const sectionHeaderStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#94a3b8',
  margin: 0,
  textTransform: 'uppercase',
};

const radioGroupStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const radioLabelStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'flex-start',
  gap: '0.5rem',
  backgroundColor: '#1e293b',
  border: '1px solid #334155',
  borderRadius: '6px',
  padding: '0.5rem',
  cursor: 'pointer',
};

const radioStyle: React.CSSProperties = {
  marginTop: '3px',
  cursor: 'pointer',
};

const radioTitleStyle: React.CSSProperties = {
  display: 'block',
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#f8fafc',
};

const radioDescStyle: React.CSSProperties = {
  display: 'block',
  fontSize: '0.65rem',
  color: '#64748b',
};

const progressSectionStyle: React.CSSProperties = {
  flexGrow: 1,
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
  borderTop: '1px solid #1e293b',
  paddingTop: '0.75rem',
};

const progressHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
};

const statusBadgeStyle = (status: string): React.CSSProperties => {
  let color = '#94a3b8';
  let bgColor = '#1e293b';

  if (status === 'RUNNING') {
    color = '#10b981';
    bgColor = 'rgba(16, 185, 129, 0.1)';
  } else if (status === 'PAUSED') {
    color = '#f59e0b';
    bgColor = 'rgba(245, 158, 11, 0.1)';
  } else if (status === 'COMPLETED') {
    color = '#3b82f6';
    bgColor = 'rgba(59, 130, 246, 0.1)';
  }

  return {
    fontSize: '0.65rem',
    fontWeight: '700',
    color,
    backgroundColor: bgColor,
    padding: '0.15rem 0.4rem',
    borderRadius: '4px',
    border: `1px solid ${color}`,
    textTransform: 'uppercase',
  };
};

const progressBarContainerStyle: React.CSSProperties = {
  height: '6px',
  backgroundColor: '#1e293b',
  borderRadius: '9999px',
  overflow: 'hidden',
};

const progressBarFillStyle: React.CSSProperties = {
  height: '100%',
  backgroundColor: '#06b6d4',
  backgroundImage: 'linear-gradient(90deg, #06b6d4, #3b82f6)',
  borderRadius: '9999px',
  transition: 'width 0.3s ease',
};

const statsGridStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.75rem',
};

const statBoxStyle: React.CSSProperties = {
  flex: 1,
  backgroundColor: '#1e293b',
  borderRadius: '6px',
  padding: '0.4rem 0.5rem',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

const statLabelStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
};

const statValueStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#f8fafc',
  fontFamily: 'monospace',
};

const bestStrategyCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(6, 182, 212, 0.05)',
  border: '1px solid rgba(6, 182, 212, 0.2)',
  borderRadius: '6px',
  padding: '0.5rem',
};

const bestHeaderStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  color: '#06b6d4',
  textTransform: 'uppercase',
  marginBottom: '0.25rem',
};

const bestNameStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  fontWeight: '700',
  color: '#f8fafc',
  marginBottom: '0.25rem',
};

const bestStatsGridStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  gap: '0.5rem',
};

const bestStatItemStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.25rem',
  fontSize: '0.7rem',
};

const bestLabelStyle: React.CSSProperties = {
  color: '#64748b',
};

const bestValStyle: React.CSSProperties = {
  color: '#cbd5e1',
  fontWeight: '600',
};

const bestProfitStyle: React.CSSProperties = {
  color: '#10b981',
  fontWeight: '700',
};

const controlsContainerStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.5rem',
  borderTop: '1px solid #1e293b',
  paddingTop: '1rem',
};

const startButtonStyle: React.CSSProperties = {
  flexGrow: 2,
  backgroundColor: '#06b6d4',
  color: '#0f172a',
  border: 'none',
  borderRadius: '4px',
  padding: '0.5rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 6px rgba(6, 182, 212, 0.2)',
};

const pauseButtonStyle: React.CSSProperties = {
  flexGrow: 2,
  backgroundColor: '#f59e0b',
  color: '#0f172a',
  border: 'none',
  borderRadius: '4px',
  padding: '0.5rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
};

const resetButtonStyle: React.CSSProperties = {
  flexGrow: 1,
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: '1px solid #475569',
  borderRadius: '4px',
  padding: '0.5rem',
  fontSize: '0.8rem',
  fontWeight: '600',
  cursor: 'pointer',
};
