import { useState } from 'react';
import {
  AVAILABLE_STRATEGIES_META,
  type SingleStrategyInstance,
} from '../services/mockStrategyData';

interface SingleStrategyListProps {
  instances: SingleStrategyInstance[];
  onCreateInstance: (newInstance: SingleStrategyInstance) => void;
}

export function SingleStrategyList({
  instances,
  onCreateInstance,
}: SingleStrategyListProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedType, setSelectedType] = useState('RSI');
  const [customName, setCustomName] = useState('');
  const [paramValues, setParamValues] = useState<Record<string, any>>({});

  // Resolve metadata for the selected strategy type
  const activeMeta = AVAILABLE_STRATEGIES_META.find((m) => m.name === selectedType);

  const handleTypeChange = (type: string) => {
    setSelectedType(type);
    const meta = AVAILABLE_STRATEGIES_META.find((m) => m.name === type);
    // Initialize default parameter values
    const defaults: Record<string, any> = {};
    if (meta?.parameters) {
      Object.keys(meta.parameters).forEach((key) => {
        defaults[key] = meta.parameters![key].default;
      });
    }
    setParamValues(defaults);
    setCustomName('');
  };

  const handleParamChange = (key: string, value: any) => {
    setParamValues((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleOpenModal = () => {
    handleTypeChange('RSI');
    setIsModalOpen(true);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeMeta) return;

    const instanceId = `${selectedType.toLowerCase()}-${Date.now()}`;
    const name = customName.trim() || `${selectedType} (${Object.values(paramValues).join(', ')})`;
    
    // Format parameters description
    const description = Object.entries(paramValues)
      .map(([k, v]) => `${k.replace(selectedType.toLowerCase(), '')}: ${v}`)
      .join(', ');

    // Mock initial signal randomly
    const signals: ('BUY' | 'SELL' | 'HOLD')[] = ['BUY', 'SELL', 'HOLD'];
    const currentSignal = signals[Math.floor(Math.random() * signals.length)];

    const newInstance: SingleStrategyInstance = {
      id: instanceId,
      name,
      type: selectedType,
      description,
      params: paramValues,
      currentSignal,
    };

    onCreateInstance(newInstance);
    setIsModalOpen(false);
  };

  const getSignalStyle = (sig: 'BUY' | 'SELL' | 'HOLD') => {
    switch (sig) {
      case 'BUY':
        return { color: '#10b981', fontWeight: 'bold' }; // Green
      case 'SELL':
        return { color: '#ef4444', fontWeight: 'bold' }; // Red
      default:
        return { color: '#64748b', fontWeight: 'bold' }; // Grey
    }
  };

  const getSignalSymbol = (sig: 'BUY' | 'SELL' | 'HOLD') => {
    switch (sig) {
      case 'BUY': return '↑ BUY';
      case 'SELL': return '↓ SELL';
      default: return '— HOLD';
    }
  };

  return (
    <div style={panelContainerStyle}>
      <div style={panelHeaderStyle}>
        <h3 style={titleStyle}>Single Indicators</h3>
        <button onClick={handleOpenModal} style={addButtonStyle}>
          + Create Instance
        </button>
      </div>

      {/* Strategies List */}
      <div style={listStyle}>
        {instances.map((instance) => (
          <div key={instance.id} style={cardStyle}>
            <div style={cardTopStyle}>
              <div>
                <span style={typeBadgeStyle}>{instance.type}</span>
                <h4 style={instanceNameStyle}>{instance.name}</h4>
              </div>
              <span style={getSignalStyle(instance.currentSignal)}>
                {getSignalSymbol(instance.currentSignal)}
              </span>
            </div>
            <p style={descStyle}>{instance.description}</p>
          </div>
        ))}
      </div>

      {/* DYNAMIC FORM MODAL */}
      {isModalOpen && (
        <div style={overlayStyle}>
          <div style={modalStyle}>
            <div style={modalHeaderStyle}>
              <h3>Create Indicator Instance</h3>
              <button onClick={() => setIsModalOpen(false)} style={closeBtnStyle}>✕</button>
            </div>
            <form onSubmit={handleSubmit} style={formStyle}>
              {/* Selector for Type */}
              <div style={formGroupStyle}>
                <label style={labelStyle}>Strategy Plugin Type</label>
                <select
                  value={selectedType}
                  onChange={(e) => handleTypeChange(e.target.value)}
                  style={inputStyle}
                >
                  {AVAILABLE_STRATEGIES_META.map((meta) => (
                    <option key={meta.name} value={meta.name}>
                      {meta.name} - {meta.description?.slice(0, 45)}...
                    </option>
                  ))}
                </select>
              </div>

              {/* Custom Display Name */}
              <div style={formGroupStyle}>
                <label style={labelStyle}>Display Name (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. RSI Oversold Aggressive"
                  value={customName}
                  onChange={(e) => setCustomName(e.target.value)}
                  style={inputStyle}
                />
              </div>

              {/* Dynamic parameters fields rendered from parameters schema */}
              {activeMeta?.parameters && (
                <div style={paramsSectionStyle}>
                  <h4 style={sectionTitleStyle}>Configure Strategy Parameters</h4>
                  {Object.entries(activeMeta.parameters).map(([key, config]) => (
                    <div key={key} style={paramRowStyle}>
                      <div style={paramLabelColStyle}>
                        <span style={paramKeyStyle}>{key}</span>
                        <span style={paramDescStyle}>{config.description}</span>
                      </div>
                      <input
                        type={config.type === 'number' ? 'number' : 'text'}
                        value={paramValues[key] ?? config.default}
                        step={config.type === 'number' ? 'any' : undefined}
                        onChange={(e) =>
                          handleParamChange(
                            key,
                            config.type === 'number' ? parseFloat(e.target.value) : e.target.value
                          )
                        }
                        style={paramInputStyle}
                        required
                      />
                    </div>
                  ))}
                </div>
              )}

              <div style={modalActionsStyle}>
                <button type="button" onClick={() => setIsModalOpen(false)} style={cancelBtnStyle}>
                  Cancel
                </button>
                <button type="submit" style={submitBtnStyle}>
                  Add to Registry
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
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
};

const panelHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  marginBottom: '1rem',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.5rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.9rem',
  fontWeight: '700',
  color: '#e2e8f0',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const addButtonStyle: React.CSSProperties = {
  backgroundColor: '#334155',
  color: '#06b6d4',
  border: '1px solid #06b6d4',
  borderRadius: '4px',
  padding: '0.25rem 0.6rem',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'pointer',
  transition: 'all 0.2s',
};

const listStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
  overflowY: 'auto',
  maxHeight: 'calc(100vh - 280px)',
  paddingRight: '4px',
};

const cardStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  border: '1px solid #334155',
  borderRadius: '6px',
  padding: '0.75rem',
};

const cardTopStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'flex-start',
  marginBottom: '0.25rem',
};

const typeBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  backgroundColor: '#0f172a',
  color: '#94a3b8',
  padding: '0.1rem 0.3rem',
  borderRadius: '4px',
  marginRight: '0.5rem',
  fontWeight: '700',
};

const instanceNameStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '600',
  color: '#f8fafc',
  margin: '0.25rem 0 0 0',
  display: 'inline-block',
};

const descStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  margin: 0,
};

// Overlay & Modal Styles
const overlayStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
  backgroundColor: 'rgba(0,0,0,0.65)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 99999,
};

const modalStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  border: '1px solid #334155',
  borderRadius: '8px',
  width: '100%',
  maxWidth: '480px',
  padding: '1.5rem',
  boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.3)',
  color: '#e2e8f0',
  fontFamily: 'system-ui, sans-serif',
};

const modalHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.75rem',
  marginBottom: '1rem',
};

const closeBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#94a3b8',
  fontSize: '1.1rem',
  cursor: 'pointer',
};

const formStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1rem',
};

const formGroupStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.35rem',
};

const labelStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  fontWeight: '600',
  color: '#94a3b8',
};

const inputStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  color: '#ffffff',
  border: '1px solid #475569',
  borderRadius: '4px',
  padding: '0.5rem',
  fontSize: '0.85rem',
  outline: 'none',
};

const paramsSectionStyle: React.CSSProperties = {
  borderTop: '1px solid #1e293b',
  paddingTop: '0.75rem',
  marginTop: '0.5rem',
};

const sectionTitleStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#06b6d4',
  margin: '0 0 0.75rem 0',
  textTransform: 'uppercase',
  letterSpacing: '0.03em',
};

const paramRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  marginBottom: '0.75rem',
};

const paramLabelColStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  maxWidth: '65%',
};

const paramKeyStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '600',
  color: '#cbd5e1',
};

const paramDescStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
};

const paramInputStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  color: '#ffffff',
  border: '1px solid #475569',
  borderRadius: '4px',
  padding: '0.35rem 0.5rem',
  fontSize: '0.8rem',
  width: '80px',
  textAlign: 'center',
  outline: 'none',
};

const modalActionsStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'flex-end',
  gap: '0.75rem',
  borderTop: '1px solid #1e293b',
  paddingTop: '1rem',
  marginTop: '0.5rem',
};

const cancelBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: '1px solid #475569',
  borderRadius: '4px',
  padding: '0.5rem 1rem',
  fontSize: '0.85rem',
  fontWeight: '600',
  cursor: 'pointer',
};

const submitBtnStyle: React.CSSProperties = {
  backgroundColor: '#06b6d4',
  color: '#0f172a',
  border: 'none',
  borderRadius: '4px',
  padding: '0.5rem 1rem',
  fontSize: '0.85rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 4px rgba(6, 182, 212, 0.2)',
};
