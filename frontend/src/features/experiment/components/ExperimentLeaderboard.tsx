import { useState } from 'react';
import type { ExperimentResult } from '../../../types/backtest';

interface ExperimentLeaderboardProps {
  experiments: ExperimentResult[];
  onSelectExperiment: (exp: ExperimentResult) => void;
  onLoadToChart: (exp: ExperimentResult) => void;
}

type SortCriterion = 'return' | 'winRate' | 'mdd' | 'tradeCount' | 'overallScore';

export function isCompetitiveExperiment(exp: ExperimentResult): boolean {
  return exp.status === 'COMPLETED' && exp.tradeCount > 0 && exp.instances.length > 0 && Boolean(exp.pair && exp.timeframe);
}

export function calculateExperimentScore(exp: ExperimentResult): number {
  return 0.5 * exp.return + 0.3 * exp.winRate - 0.2 * exp.mdd;
}

function eligibilityLabel(exp: ExperimentResult): string {
  if (isCompetitiveExperiment(exp)) return 'RANKED';
  if (exp.status !== 'COMPLETED') return exp.status;
  if (exp.tradeCount === 0) return 'NO TRADES';
  return 'LEGACY';
}

export function ExperimentLeaderboard({
  experiments,
  onSelectExperiment,
  onLoadToChart,
}: ExperimentLeaderboardProps) {
  const [topK, setTopK] = useState<number>(10);
  const [sortBy, setSortBy] = useState<SortCriterion>('return');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [showHistory, setShowHistory] = useState(false);
  const hiddenHistoryCount = experiments.filter((exp) => !isCompetitiveExperiment(exp)).length;

  // Sort and filter logic
  const getProcessedData = () => {
    const filtered = experiments.filter((exp) => {
      const matchSearch = exp.instances.map(i => i.type).join(' + ').toLowerCase().includes(searchQuery.toLowerCase()) ||
                          exp.id.toLowerCase().includes(searchQuery.toLowerCase());
      return matchSearch && (showHistory || isCompetitiveExperiment(exp));
    });

    // Sort by selected criterion
    filtered.sort((a, b) => {
      if (isCompetitiveExperiment(a) !== isCompetitiveExperiment(b)) {
        return isCompetitiveExperiment(a) ? -1 : 1;
      }
      if (sortBy === 'overallScore') {
        return calculateExperimentScore(b) - calculateExperimentScore(a);
      }
      if (sortBy === 'mdd') {
        // Lower Drawdown is better (ascending order)
        return a.mdd - b.mdd;
      }
      // Standard descending sort
      const valA = a[sortBy] as number;
      const valB = b[sortBy] as number;
      return valB - valA;
    });

    // Slice to Top K
    return filtered.slice(0, topK);
  };

  const processedExperiments = getProcessedData();

  const getRankBadge = (rank: number) => {
    switch (rank) {
      case 1: return <span style={{ ...medalStyle, backgroundColor: '#f59e0b', color: '#000000' }} title="Gold Medal">1</span>;
      case 2: return <span style={{ ...medalStyle, backgroundColor: '#cbd5e1', color: '#000000' }} title="Silver Medal">2</span>;
      case 3: return <span style={{ ...medalStyle, backgroundColor: '#b45309', color: '#ffffff' }} title="Bronze Medal">3</span>;
      default: return <span style={rankTextStyle}>{rank}</span>;
    }
  };

  return (
    <div style={containerStyle}>
      {/* Top Toolbar */}
      <div style={toolbarStyle}>
        <div style={leftToolbarStyle}>
          {/* Search Filter */}
          <input
            type="text"
            placeholder="Filter strategies..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={searchStyle}
          />
          {hiddenHistoryCount > 0 && (
            <label style={historyToggleStyle}>
              <input type="checkbox" checked={showHistory} onChange={(event) => setShowHistory(event.target.checked)} />
              Show {hiddenHistoryCount} non-competitive historical run{hiddenHistoryCount === 1 ? '' : 's'}
            </label>
          )}
        </div>

        <div style={rightToolbarStyle}>
          {/* Top K Selector */}
          <div style={filterGroupStyle}>
            <label style={labelStyle}>Show</label>
            <select
              value={topK}
              onChange={(e) => setTopK(parseInt(e.target.value))}
              style={selectStyle}
            >
              <option value={5}>Top 5</option>
              <option value={10}>Top 10</option>
              <option value={20}>Top 20</option>
            </select>
          </div>

          {/* Sort Criterion Selector */}
          <div style={filterGroupStyle}>
            <label style={labelStyle}>Sort By</label>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as SortCriterion)}
              style={selectStyle}
            >
              <option value="return">Total Return (%)</option>
              <option value="winRate">Win Rate (%)</option>
              <option value="mdd">Max Drawdown (MDD)</option>
              <option value="tradeCount">Total Trades</option>
              <option value="overallScore">Overall Score (Risk-Adjusted)</option>
            </select>
          </div>
        </div>
      </div>

      {/* Leaderboard Table */}
      <div style={tableWrapperStyle}>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={thCenterStyle}>Rank</th>
              <th style={thLeftStyle}>Exp ID & Version</th>
              <th style={thLeftStyle}>Strategy Composition</th>
              <th style={thLeftStyle}>Market</th>
              <th style={thCenterStyle}>Eligibility</th>
              <th style={thRightStyle}>Return</th>
              <th style={thRightStyle}>Win Rate</th>
              <th style={thRightStyle}>Max Drawdown</th>
              <th style={thRightStyle}>Trades</th>
              {sortBy === 'overallScore' && <th style={thRightStyle}>Overall Score</th>}
              <th style={thCenterStyle}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {processedExperiments.length > 0 ? (
              processedExperiments.map((exp, index) => {
                const isWarningMDD = exp.mdd >= 20;
                const overallScore = calculateExperimentScore(exp);
                const eligibility = eligibilityLabel(exp);
                const versions = Object.entries(exp.strategyVersions || {}).map(([name, version]) => `${name} ${version}`).join(', ');

                return (
                  <tr key={exp.id} style={trStyle}>
                    <td style={tdCenterStyle}>{getRankBadge(index + 1)}</td>
                    <td style={tdLeftStyle}>
                      <span style={expIdStyle}>{exp.id}</span>
                      <span style={versionBadgeStyle}>{versions || 'legacy'}</span>
                    </td>
                    <td style={tdLeftStyle}>
                      <span style={compositionStyle}>
                        {exp.instances.map(i => i.type).join(' + ')}
                      </span>
                      <span style={policyStyle}>{exp.policy}</span>
                    </td>
                    <td style={tdLeftStyle}>{exp.pair && exp.timeframe ? `${exp.pair} · ${exp.timeframe}` : 'legacy / unknown'}</td>
                    <td style={tdCenterStyle}>
                      <span style={isCompetitiveExperiment(exp) ? rankedBadgeStyle : historyBadgeStyle}>{eligibility}</span>
                    </td>
                    <td style={exp.return >= 0 ? tdBuyStyle : tdSellStyle}>
                      {exp.return >= 0 ? `+${exp.return.toFixed(2)}%` : `${exp.return.toFixed(2)}%`}
                    </td>
                    <td style={tdRightStyle}>{exp.winRate.toFixed(2)}%</td>
                    <td style={isWarningMDD ? tdWarnStyle : tdRightStyle}>
                      {exp.mdd.toFixed(2)}%
                      {isWarningMDD && ' '}
                    </td>
                    <td style={tdRightStyle}>{exp.tradeCount}</td>
                    {sortBy === 'overallScore' && (
                      <td style={tdScoreStyle}>{overallScore.toFixed(2)}</td>
                    )}
                    <td style={tdCenterStyle}>
                      <div style={actionGroupStyle}>
                        <button
                          onClick={() => onSelectExperiment(exp)}
                          style={actionBtnStyle}
                          title="View Provenance Metadata"
                        >
                          Details
                        </button>
                        <button
                          onClick={() => onLoadToChart(exp)}
                          style={{ ...actionBtnStyle, backgroundColor: '#2563eb', color: '#ffffff' }}
                          title="View this run's trades and chart below"
                        >
                          Load
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            ) : (
              <tr>
                <td colSpan={sortBy === 'overallScore' ? 11 : 10} style={emptyTdStyle}>
                  {showHistory ? 'No experiments matching search criteria.' : 'No rank-eligible experiments yet. Run a backtest that produces at least one trade.'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const containerStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1.25rem',
  boxSizing: 'border-box',
};

const toolbarStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  marginBottom: '1rem',
  flexWrap: 'wrap',
  gap: '1rem',
};

const leftToolbarStyle: React.CSSProperties = {
  flexGrow: 1,
  maxWidth: '420px',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const historyToggleStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.4rem',
  color: '#64748b',
  fontSize: '0.72rem',
  cursor: 'pointer',
};

const rightToolbarStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1rem',
  alignItems: 'center',
};

const searchStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#ffffff',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.45rem 0.75rem',
  fontSize: '0.8rem',
  width: '100%',
  outline: 'none',
};

const filterGroupStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
};

const labelStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  fontWeight: '600',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#ffffff',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.4rem 0.65rem',
  fontSize: '0.8rem',
  outline: 'none',
  cursor: 'pointer',
};

const tableWrapperStyle: React.CSSProperties = {
  overflowX: 'auto',
};

const tableStyle: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  fontSize: '0.8rem',
};

const trStyle: React.CSSProperties = {
  borderBottom: '1px solid #e2e8f0',
  transition: 'background-color 0.2s',
  cursor: 'pointer',
};

const thLeftStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'left',
  padding: '0.75rem 0.5rem',
  fontWeight: '600',
  borderBottom: '2px solid #e2e8f0',
};

const thRightStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'right',
  padding: '0.75rem 0.5rem',
  fontWeight: '600',
  borderBottom: '2px solid #e2e8f0',
};

const thCenterStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'center',
  padding: '0.75rem 0.5rem',
  fontWeight: '600',
  borderBottom: '2px solid #e2e8f0',
};

const tdLeftStyle: React.CSSProperties = {
  color: '#0f172a',
  padding: '0.75rem 0.5rem',
};

const tdRightStyle: React.CSSProperties = {
  color: '#0f172a',
  textAlign: 'right',
  padding: '0.75rem 0.5rem',
};

const tdCenterStyle: React.CSSProperties = {
  color: '#0f172a',
  textAlign: 'center',
  padding: '0.75rem 0.5rem',
};

const tdBuyStyle: React.CSSProperties = {
  color: '#10b981', // Green
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.75rem 0.5rem',
};

const tdSellStyle: React.CSSProperties = {
  color: '#ef4444', // Red
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.75rem 0.5rem',
};

const tdWarnStyle: React.CSSProperties = {
  color: '#f59e0b', // Orange warn
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.75rem 0.5rem',
};

const tdScoreStyle: React.CSSProperties = {
  color: '#2563eb', // Cyan
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.75rem 0.5rem',
  fontFamily: 'monospace',
};

const emptyTdStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'center',
  padding: '2rem',
};

const medalStyle: React.CSSProperties = {
  padding: '0.2rem 0.5rem',
  borderRadius: '4px',
  fontWeight: '700',
  fontSize: '0.7rem',
  display: 'inline-block',
  minWidth: '22px',
};

const rankTextStyle: React.CSSProperties = {
  fontWeight: '600',
  color: '#64748b',
};

const expIdStyle: React.CSSProperties = {
  fontWeight: '700',
  color: '#cbd5e1',
  fontFamily: 'monospace',
  display: 'block',
};

const versionBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  backgroundColor: '#e2e8f0',
  color: '#2563eb',
  padding: '0.05rem 0.25rem',
  borderRadius: '4px',
  border: '1px solid #cbd5e1',
  fontFamily: 'monospace',
  display: 'inline-block',
};

const rankedBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: 700,
  color: '#047857',
  backgroundColor: '#d1fae5',
  border: '1px solid #6ee7b7',
  borderRadius: '999px',
  padding: '0.15rem 0.4rem',
};

const historyBadgeStyle: React.CSSProperties = {
  ...rankedBadgeStyle,
  color: '#92400e',
  backgroundColor: '#fef3c7',
  borderColor: '#fcd34d',
};

const compositionStyle: React.CSSProperties = {
  fontWeight: '600',
  color: '#0f172a',
  display: 'block',
};

const policyStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
  textTransform: 'uppercase',
  display: 'block',
};

const actionGroupStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.35rem',
  justifyContent: 'center',
};

const actionBtnStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#94a3b8',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.25rem 0.5rem',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'pointer',
  transition: 'all 0.15s',
  outline: 'none',
};
