import { useState, useMemo } from 'react';
import { type SingleStrategyInstance, COMPOSITE_PRESETS } from '../services/mockStrategyData';
import type { StrategyInstance } from '../../../types/backtest';

interface CompositeStrategyBuilderProps {
  singleInstances: SingleStrategyInstance[];
  onStartBacktest: (config: {
    instances: StrategyInstance[];
    policy: 'majority' | 'weighted';
  }) => void;
}

export function CompositeStrategyBuilder({
  singleInstances,
  onStartBacktest,
}: CompositeStrategyBuilderProps) {
  // Track selected indicators for composition
  const [selectedIds, setSelectedIds] = useState<string[]>(['rsi-14', 'ma-20', 'sr-3']);
  const [weights, setWeights] = useState<Record<string, number>>({
    'rsi-14': 0.4,
    'ma-20': 0.3,
    'sr-3': 0.3,
  });
  const [policy, setPolicy] = useState<'majority' | 'weighted'>('weighted');
  const [threshold, setThreshold] = useState<number>(0.3);

  // Load preset configurations
  const applyPreset = (presetName: string) => {
    const preset = COMPOSITE_PRESETS.find((p) => p.name === presetName);
    if (!preset) return;

    // Filter preset strategies to only those available in list
    const validIds = preset.strategies.filter(id => 
      singleInstances.some(inst => inst.id === id)
    );

    if (validIds.length === 0) {
      alert("Preset indicators aren't loaded in the Single list. Create them first!");
      return;
    }

    setSelectedIds(validIds);
    setWeights(preset.weights);
    setPolicy(preset.policy as 'majority' | 'weighted');
  };

  // Composite signal calculation (derived from state during render)
  const { compositeScore, compositeSignal } = useMemo(() => {
    if (selectedIds.length === 0) {
      return { compositeScore: 0, compositeSignal: 'HOLD' as const };
    }

    let score = 0;
    let sumWeights = 0;

    // Filter weights for active selections only
    selectedIds.forEach((id) => {
      const weight = weights[id] ?? 0;
      sumWeights += weight;
    });

    if (policy === 'weighted') {
      selectedIds.forEach((id) => {
        const inst = singleInstances.find((i) => i.id === id);
        if (!inst) return;

        // Map signal to value: BUY = +1, SELL = -1, HOLD = 0
        let sigValue = 0;
        if (inst.currentSignal === 'BUY') sigValue = 1;
        else if (inst.currentSignal === 'SELL') sigValue = -1;

        // Normalized weight
        const normalizedWeight = sumWeights > 0 ? (weights[id] ?? 0) / sumWeights : 0;
        score += sigValue * normalizedWeight;
      });

      const finalScore = Number(score.toFixed(2));
      let finalSignal: 'LONG' | 'SHORT' | 'HOLD' = 'HOLD';
      if (score >= threshold) finalSignal = 'LONG';
      else if (score <= -threshold) finalSignal = 'SHORT';

      return { compositeScore: finalScore, compositeSignal: finalSignal };
    } else {
      // Majority voting logic
      let buyCount = 0;
      let sellCount = 0;
      let holdCount = 0;

      selectedIds.forEach((id) => {
        const inst = singleInstances.find((i) => i.id === id);
        if (!inst) return;
        if (inst.currentSignal === 'BUY') buyCount++;
        else if (inst.currentSignal === 'SELL') sellCount++;
        else holdCount++;
      });

      const maxVotes = Math.max(buyCount, sellCount, holdCount);
      let finalSignal: 'LONG' | 'SHORT' | 'HOLD' = 'HOLD';
      if (maxVotes === buyCount && buyCount > 0) finalSignal = 'LONG';
      else if (maxVotes === sellCount && sellCount > 0) finalSignal = 'SHORT';

      return { compositeScore: 0, compositeSignal: finalSignal };
    }
  }, [selectedIds, weights, policy, threshold, singleInstances]);

  const handleToggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      let next = [...prev];
      if (prev.includes(id)) {
        next = next.filter((item) => item !== id);
      } else {
        next.push(id);
      }

      // Re-initialize default weight for newly selected indicators
      if (next.length > 0) {
        const equalWeight = Number((1 / next.length).toFixed(2));
        const updatedWeights: Record<string, number> = {};
        next.forEach((activeId) => {
          updatedWeights[activeId] = weights[activeId] ?? equalWeight;
        });
        setWeights(updatedWeights);
      }

      return next;
    });
  };

  const handleWeightChange = (id: string, val: number) => {
    setWeights((prev) => ({
      ...prev,
      [id]: val,
    }));
  };

  const handleBacktestSubmit = () => {
    if (selectedIds.length === 0) {
      alert('Please select at least one indicator to build a composite strategy.');
      return;
    }
    
    // Normalize weights, then pair each selected instance with its own
    // type/params/weight — no shared bag, so MA(20) and MA(50) (or any two
    // instances of the same type) stay independently configured.
    const sum = selectedIds.reduce((acc, id) => acc + (weights[id] ?? 0), 0);
    const instances: StrategyInstance[] = selectedIds
      .map((id): StrategyInstance | null => {
        const inst = singleInstances.find((i) => i.id === id);
        if (!inst) return null;
        const weight = sum > 0 ? Number(((weights[id] ?? 0) / sum).toFixed(2)) : 0;
        return { type: inst.type, params: inst.params, weight };
      })
      .filter((inst): inst is StrategyInstance => inst !== null);

    onStartBacktest({ instances, policy });
  };

  const getCompositeSignalStyle = (sig: 'LONG' | 'SHORT' | 'HOLD') => {
    switch (sig) {
      case 'LONG':
        return { backgroundColor: '#064e3b', color: '#10b981', border: '1px solid #10b981' };
      case 'SHORT':
        return { backgroundColor: '#7f1d1d', color: '#ef4444', border: '1px solid #ef4444' };
      default:
        return { backgroundColor: '#e2e8f0', color: '#94a3b8', border: '1px solid #475569' };
    }
  };

  const sumOfWeights = selectedIds.reduce((acc, id) => acc + (weights[id] ?? 0), 0);

  return (
    <div style={panelContainerStyle}>
      <h3 style={titleStyle}>Composite Strategy Builder</h3>

      {/* Preset configurations quick selector */}
      <div style={presetSectionStyle}>
        <span style={sectionLabelStyle}>Quick Combination Presets:</span>
        <div style={presetButtonGroupStyle}>
          {COMPOSITE_PRESETS.map((preset) => (
            <button
              key={preset.name}
              onClick={() => applyPreset(preset.name)}
              style={presetButtonStyle}
            >
              {preset.name}
            </button>
          ))}
        </div>
      </div>

      {/* Checklist and sliders */}
      <div style={builderSectionStyle}>
        <h4 style={sectionHeaderStyle}>Configure Weighted Voting Weights</h4>
        <div style={votingListStyle}>
          {singleInstances.map((inst) => {
            const isChecked = selectedIds.includes(inst.id);
            const weightVal = weights[inst.id] ?? 0;

            return (
              <div key={inst.id} style={isChecked ? activeRowStyle : rowStyle}>
                <div style={checkColStyle}>
                  <input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => handleToggleSelect(inst.id)}
                    style={checkboxStyle}
                  />
                  <div>
                    <div style={instNameStyle}>{inst.name}</div>
                    <div style={instSignalStyle}>
                      Signal: <span style={inst.currentSignal === 'BUY' ? buyTextStyle : inst.currentSignal === 'SELL' ? sellTextStyle : holdTextStyle}>
                        {inst.currentSignal}
                      </span>
                    </div>
                  </div>
                </div>

                {isChecked && (
                  <div style={sliderColStyle}>
                    <input
                      type="range"
                      min="0.0"
                      max="1.0"
                      step="0.05"
                      value={weightVal}
                      onChange={(e) => handleWeightChange(inst.id, parseFloat(e.target.value))}
                      style={sliderStyle}
                    />
                    <span style={weightBadgeStyle}>w: {weightVal.toFixed(2)}</span>
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* Weights validation status helper */}
        {selectedIds.length > 0 && (
          <div style={weightValidationStyle}>
            <span>Sum of raw weights: <strong>{sumOfWeights.toFixed(2)}</strong></span>
            {Math.abs(sumOfWeights - 1.0) > 0.01 && (
              <span style={{ color: '#f59e0b', fontSize: '0.7rem' }}>
                Weights will be automatically normalized to 1.00 for backtests.
              </span>
            )}
          </div>
        )}
      </div>

      {/* Decision policy & Entry thresholds configuration */}
      <div style={policySectionStyle}>
        <div style={policyRowStyle}>
          <div style={formGroupStyle}>
            <label style={labelStyle}>Decision Policy</label>
            <select
              value={policy}
              onChange={(e) => setPolicy(e.target.value as 'majority' | 'weighted')}
              style={selectStyle}
            >
              <option value="weighted">Weighted Vote Score</option>
              <option value="majority">Majority Rule (Equal weights)</option>
            </select>
          </div>

          {policy === 'weighted' && (
            <div style={formGroupStyle}>
              <label style={labelStyle}>Entry Threshold (|score| ≥ X)</label>
              <input
                type="number"
                min="0.05"
                max="0.95"
                step="0.05"
                value={threshold}
                onChange={(e) => setThreshold(parseFloat(e.target.value))}
                style={numInputStyle}
              />
            </div>
          )}
        </div>

        {/* Realtime Composite Signal display */}
        <div style={signalAreaStyle}>
          <span style={signalTitleStyle}>Realtime Composite Signal:</span>
          <div style={signalFlexStyle}>
            <div style={{ ...signalBadgeStyle, ...getCompositeSignalStyle(compositeSignal) }}>
              <span style={signalTextLgStyle}>{compositeSignal}</span>
              {policy === 'weighted' && (
                <span style={scoreTextStyle}>Score: {compositeScore > 0 ? `+${compositeScore}` : compositeScore}</span>
              )}
            </div>
            <span style={statusLabelStyle}>● Auto updating from market ticks</span>
          </div>
        </div>
      </div>

      {/* Action buttons */}
      <div style={actionsContainerStyle}>
        <button style={saveButtonStyle}>
          Save Composite Strategy
        </button>
        <button onClick={handleBacktestSubmit} style={backtestButtonStyle}>
          Run Backtest Now
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

const presetSectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const sectionLabelStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  fontWeight: '500',
};

const presetButtonGroupStyle: React.CSSProperties = {
  display: 'flex',
  flexWrap: 'wrap',
  gap: '0.4rem',
};

const presetButtonStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#94a3b8',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.25rem 0.5rem',
  fontSize: '0.7rem',
  fontWeight: '600',
  cursor: 'pointer',
  transition: 'all 0.15s',
};

const builderSectionStyle: React.CSSProperties = {
  flexGrow: 1,
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
};

const sectionHeaderStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#2563eb',
  margin: 0,
  textTransform: 'uppercase',
};

const votingListStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
  overflowY: 'auto',
  maxHeight: '280px',
};

const rowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '0.5rem',
  backgroundColor: 'transparent',
  border: '1px solid #e2e8f0',
  borderRadius: '6px',
};

const activeRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '0.5rem',
  backgroundColor: '#e2e8f0',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
};

const checkColStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
};

const checkboxStyle: React.CSSProperties = {
  width: '14px',
  height: '14px',
  cursor: 'pointer',
};

const instNameStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '600',
  color: '#0f172a',
};

const instSignalStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
};

const buyTextStyle: React.CSSProperties = { color: '#10b981', fontWeight: 'bold' };
const sellTextStyle: React.CSSProperties = { color: '#ef4444', fontWeight: 'bold' };
const holdTextStyle: React.CSSProperties = { color: '#64748b', fontWeight: 'bold' };

const sliderColStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
};

const sliderStyle: React.CSSProperties = {
  width: '70px',
  cursor: 'pointer',
};

const weightBadgeStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  backgroundColor: '#ffffff',
  color: '#2563eb',
  padding: '0.1rem 0.3rem',
  borderRadius: '4px',
  fontFamily: 'monospace',
  border: '1px solid #cbd5e1',
};

const weightValidationStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.75rem',
  color: '#94a3b8',
};

const policySectionStyle: React.CSSProperties = {
  borderTop: '1px solid #e2e8f0',
  paddingTop: '0.75rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.85rem',
};

const policyRowStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1rem',
};

const formGroupStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.25rem',
  flexGrow: 1,
};

const labelStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  fontWeight: '600',
};

const selectStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#ffffff',
  border: '1px solid #475569',
  borderRadius: '4px',
  padding: '0.35rem 0.5rem',
  fontSize: '0.8rem',
  outline: 'none',
  cursor: 'pointer',
};

const numInputStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#ffffff',
  border: '1px solid #475569',
  borderRadius: '4px',
  padding: '0.35rem 0.5rem',
  fontSize: '0.8rem',
  width: '50px',
  textAlign: 'center',
  outline: 'none',
};

const signalAreaStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.35rem',
};

const signalTitleStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#94a3b8',
  fontWeight: '500',
};

const signalFlexStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '1rem',
};

const signalBadgeStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  borderRadius: '6px',
  width: '120px',
  padding: '0.4rem 0.75rem',
  textAlign: 'center',
};

const signalTextLgStyle: React.CSSProperties = {
  fontSize: '1.1rem',
  fontWeight: '800',
  letterSpacing: '0.05em',
};

const scoreTextStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '600',
  opacity: 0.95,
  fontFamily: 'monospace',
};

const statusLabelStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#475569',
  fontWeight: '500',
};

const actionsContainerStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.75rem',
  borderTop: '1px solid #e2e8f0',
  paddingTop: '1rem',
};

const saveButtonStyle: React.CSSProperties = {
  flexGrow: 1,
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: '1px solid #475569',
  borderRadius: '6px',
  padding: '0.6rem',
  fontSize: '0.85rem',
  fontWeight: '600',
  cursor: 'pointer',
  transition: 'background-color 0.2s',
};

const backtestButtonStyle: React.CSSProperties = {
  flexGrow: 2,
  backgroundColor: '#10b981', // Green-500
  color: '#064e3b',
  border: 'none',
  borderRadius: '6px',
  padding: '0.6rem',
  fontSize: '0.85rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 8px rgba(16, 185, 129, 0.2)',
  transition: 'background-color 0.2s',
};
