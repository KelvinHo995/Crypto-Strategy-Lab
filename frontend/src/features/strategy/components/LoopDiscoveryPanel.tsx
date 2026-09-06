import { useState } from 'react';
import type { DiscoveryStats } from '../services/mockStrategyData';

export interface DiscoveryLoopConfig {
  timeframe: '5m' | '15m' | '1h' | '4h';
  maxCandidates: number;
  maxDurationSeconds: number;
  noImprovementLimit: number;
}

interface LoopDiscoveryPanelProps {
  stats: DiscoveryStats;
  onStart: (config: DiscoveryLoopConfig) => void | Promise<void>;
  onReset: () => void;
}

export function LoopDiscoveryPanel({
  stats,
  onStart,
  onReset,
}: LoopDiscoveryPanelProps) {
  const [config, setConfig] = useState<DiscoveryLoopConfig>({
    timeframe: '1h',
    maxCandidates: 10,
    maxDurationSeconds: 300,
    noImprovementLimit: 5,
  });

  const handleStart = () => {
    void onStart(config);
  };

  const progressPercent = stats.totalIterations > 0
    ? Math.min(100, (stats.iteration / stats.totalIterations) * 100)
    : 0;
  const formatMoney = (value: number) => `${value >= 0 ? '+' : '-'}${Math.abs(value).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} USDT`;
  const formatPercent = (value: number) => `${value.toFixed(2)}%`;

  return (
    <div style={panelContainerStyle}>
      <h3 style={titleStyle}>Loop Discovery Engine</h3>
      <small style={liveContractStyle}>Random Search chạy qua API; tiến độ chỉ cập nhật từ WebSocket.</small>

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
              checked
              readOnly
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
              checked={false}
              disabled
              style={radioStyle}
            />
            <div>
              <span style={radioTitleStyle}>Domain-guided</span>
              <span style={radioDescStyle}>Chưa có generator ở backend</span>
            </div>
          </label>

          <label style={radioLabelStyle}>
            <input
              type="radio"
              name="method"
              checked={false}
              disabled
              style={radioStyle}
            />
            <div>
              <span style={radioTitleStyle}>Genetic Algorithm</span>
              <span style={radioDescStyle}>Chưa có generator ở backend</span>
            </div>
          </label>
        </div>
      </div>

      <div style={sectionStyle}>
        <h4 style={sectionHeaderStyle}>Loop limits</h4>
        <div style={settingsGridStyle}>
          <label style={fieldLabelStyle}>
            Timeframe
            <select
              value={config.timeframe}
              disabled={stats.status === 'RUNNING'}
              onChange={event => setConfig(current => ({ ...current, timeframe: event.target.value as DiscoveryLoopConfig['timeframe'] }))}
              style={fieldStyle}
            >
              {['5m', '15m', '1h', '4h'].map(value => <option key={value}>{value}</option>)}
            </select>
          </label>
          <label style={fieldLabelStyle}>
            Candidates (2–200)
            <input
              type="number"
              min={2}
              max={200}
              value={config.maxCandidates}
              disabled={stats.status === 'RUNNING'}
              onChange={event => setConfig(current => ({ ...current, maxCandidates: Number(event.target.value) }))}
              style={fieldStyle}
            />
          </label>
          <label style={fieldLabelStyle}>
            Max duration (60–3600s)
            <input
              type="number"
              min={60}
              max={3600}
              value={config.maxDurationSeconds}
              disabled={stats.status === 'RUNNING'}
              onChange={event => setConfig(current => ({ ...current, maxDurationSeconds: Number(event.target.value) }))}
              style={fieldStyle}
            />
          </label>
          <label style={fieldLabelStyle}>
            No improvement limit
            <input
              type="number"
              min={0}
              value={config.noImprovementLimit}
              disabled={stats.status === 'RUNNING'}
              onChange={event => setConfig(current => ({ ...current, noImprovementLimit: Number(event.target.value) }))}
              style={fieldStyle}
            />
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

        {stats.searchId && <div style={searchIdStyle}>Search ID: {stats.searchId}</div>}
        {stats.statusMessage && (
          <div role="status" style={messageStyle(stats.status)}>{stats.statusMessage}</div>
        )}

        {/* Best Strategy So Far details */}
        {stats.bestStrategy && (
          <div style={bestStrategyCardStyle}>
            <div style={bestHeaderStyle}>
              Best Candidate So Far
            </div>
            <div style={bestNameStyle}>{stats.bestStrategy.name}</div>
            <div style={bestStatsGridStyle}>
              <div style={bestStatItemStyle}>
                <span style={bestLabelStyle}>Return:</span>
                <span style={bestProfitStyle}>{formatMoney(stats.bestStrategy.profit)}</span>
              </div>
              <div style={bestStatItemStyle}>
                <span style={bestLabelStyle}>Winrate:</span>
                <span style={bestValStyle}>{formatPercent(stats.bestStrategy.winrate)}</span>
              </div>
              <div style={bestStatItemStyle}>
                <span style={bestLabelStyle}>MDD:</span>
                <span style={bestValStyle}>{formatPercent(stats.bestStrategy.mdd)}</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Control Buttons */}
      <div style={controlsContainerStyle}>
        <button
          onClick={handleStart}
          style={{ ...startButtonStyle, opacity: stats.status === 'RUNNING' ? 0.55 : 1 }}
          disabled={stats.status === 'RUNNING'}
        >
          {stats.status === 'RUNNING' ? 'Discovery đang chạy' : stats.status === 'IDLE' ? 'Start Loop Discovery' : 'Run new discovery'}
        </button>
        <button onClick={onReset} style={resetButtonStyle} disabled={stats.status === 'RUNNING'}>
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
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
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
  color: '#0f172a',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
};

const liveContractStyle: React.CSSProperties = {
  color: '#047857',
  background: '#ecfdf5',
  border: '1px solid #a7f3d0',
  borderRadius: '5px',
  padding: '0.4rem 0.5rem',
};

const flowDiagramStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  backgroundColor: '#ffffff',
  padding: '0.5rem 0.75rem',
  borderRadius: '6px',
  border: '1px solid #e2e8f0',
};

const flowStepStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  fontWeight: '700',
  color: '#2563eb',
  backgroundColor: '#ffffff',
  border: '1px solid #cbd5e1',
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

const settingsGridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
  gap: '0.5rem',
};

const fieldLabelStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.25rem',
  color: '#64748b',
  fontSize: '0.65rem',
  fontWeight: 600,
};

const fieldStyle: React.CSSProperties = {
  minWidth: 0,
  width: '100%',
  boxSizing: 'border-box',
  border: '1px solid #cbd5e1',
  borderRadius: '5px',
  background: '#fff',
  color: '#0f172a',
  padding: '0.35rem 0.4rem',
  fontSize: '0.72rem',
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
  backgroundColor: '#e2e8f0',
  border: '1px solid #cbd5e1',
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
  color: '#0f172a',
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
  borderTop: '1px solid #e2e8f0',
  paddingTop: '0.75rem',
};

const progressHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
};

const statusBadgeStyle = (status: string): React.CSSProperties => {
  let color = '#94a3b8';
  let bgColor = '#e2e8f0';

  if (status === 'RUNNING') {
    color = '#10b981';
    bgColor = 'rgba(16, 185, 129, 0.1)';
  } else if (status === 'STOPPED') {
    color = '#f59e0b';
    bgColor = 'rgba(245, 158, 11, 0.1)';
  } else if (status === 'COMPLETED') {
    color = '#3b82f6';
    bgColor = 'rgba(59, 130, 246, 0.1)';
  } else if (status === 'FAILED') {
    color = '#dc2626';
    bgColor = 'rgba(220, 38, 38, 0.08)';
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
  backgroundColor: '#e2e8f0',
  borderRadius: '9999px',
  overflow: 'hidden',
};

const progressBarFillStyle: React.CSSProperties = {
  height: '100%',
  backgroundColor: '#2563eb',
  backgroundImage: 'linear-gradient(90deg, #2563eb, #3b82f6)',
  borderRadius: '9999px',
  transition: 'width 0.3s ease',
};

const statsGridStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.75rem',
};

const statBoxStyle: React.CSSProperties = {
  flex: 1,
  backgroundColor: '#e2e8f0',
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
  color: '#0f172a',
  fontFamily: 'monospace',
};

const searchIdStyle: React.CSSProperties = {
  color: '#64748b',
  fontFamily: 'monospace',
  fontSize: '0.62rem',
  overflowWrap: 'anywhere',
};

const messageStyle = (status: DiscoveryStats['status']): React.CSSProperties => ({
  borderRadius: '5px',
  border: `1px solid ${status === 'FAILED' ? '#fecaca' : status === 'STOPPED' ? '#fde68a' : '#bfdbfe'}`,
  background: status === 'FAILED' ? '#fef2f2' : status === 'STOPPED' ? '#fffbeb' : '#eff6ff',
  color: status === 'FAILED' ? '#b91c1c' : status === 'STOPPED' ? '#92400e' : '#1d4ed8',
  padding: '0.45rem 0.5rem',
  fontSize: '0.68rem',
});

const bestStrategyCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(6, 182, 212, 0.05)',
  border: '1px solid rgba(6, 182, 212, 0.2)',
  borderRadius: '6px',
  padding: '0.5rem',
};

const bestHeaderStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  color: '#2563eb',
  textTransform: 'uppercase',
  marginBottom: '0.25rem',
};

const bestNameStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  fontWeight: '700',
  color: '#0f172a',
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
  borderTop: '1px solid #e2e8f0',
  paddingTop: '1rem',
};

const startButtonStyle: React.CSSProperties = {
  flexGrow: 2,
  backgroundColor: '#2563eb',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.5rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 6px rgba(6, 182, 212, 0.2)',
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
