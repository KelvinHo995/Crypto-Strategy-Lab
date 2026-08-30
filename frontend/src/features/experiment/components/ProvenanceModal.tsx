import type { ExperimentResult } from '../../../types/backtest';

interface ProvenanceModalProps {
  experiment: ExperimentResult | null;
  onClose: () => void;
  onReplicate: (exp: ExperimentResult) => void;
}

export function ProvenanceModal({
  experiment,
  onClose,
  onReplicate,
}: ProvenanceModalProps) {
  if (!experiment) return null;

  const formattedDate = new Date(experiment.createdAt).toLocaleString();

  return (
    <div style={overlayStyle}>
      <div style={modalStyle}>
        {/* Header */}
        <div style={modalHeaderStyle}>
          <div>
            <h3 style={titleStyle}>Provenance Metadata Viewer</h3>
            <span style={hashStyle}>Run ID: {experiment.id}</span>
          </div>
          <button onClick={onClose} style={closeBtnStyle}>✕</button>
        </div>

        {/* Content Body */}
        <div style={bodyStyle}>
          {/* Metadata Grid */}
          <div style={sectionStyle}>
            <h4 style={sectionTitleStyle}>Execution Information</h4>
            <div style={metaGridStyle}>
              <div style={metaRowStyle}>
                <span style={metaLabelStyle}>Run Timestamp:</span>
                <span style={metaValueStyle}>{formattedDate}</span>
              </div>
              <div style={metaRowStyle}>
                <span style={metaLabelStyle}>Backtest Period:</span>
                <span style={metaValueStyle}>{experiment.datasetPeriod}</span>
              </div>
              <div style={metaRowStyle}>
                <span style={metaLabelStyle}>Candidate Ref:</span>
                <span style={{ ...metaValueStyle, fontFamily: 'monospace' }}>{experiment.candidateId}</span>
              </div>
              <div style={metaRowStyle}>
                <span style={metaLabelStyle}>Combination Policy:</span>
                <span style={{ ...metaValueStyle, textTransform: 'uppercase', color: '#06b6d4' }}>
                  {experiment.policy}
                </span>
              </div>
            </div>
          </div>

          {/* Code Versioning Provenance */}
          <div style={sectionStyle}>
            <h4 style={sectionTitleStyle}>Code & Model Provenance (Reproducibility)</h4>
            <div style={versionGridStyle}>
              {Object.entries(experiment.strategyVersions).map(([strategyName, version]) => (
                <div key={strategyName} style={versionBadgeStyle}>
                  <span style={versionKeyStyle}>{strategyName}</span>
                  <span style={versionValStyle}>{version}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Strategy parameters JSON config */}
          <div style={sectionStyle}>
            <h4 style={sectionTitleStyle}>Active Parameters Schema (JSON)</h4>
            <pre style={jsonCodeStyle}>
              {JSON.stringify(experiment.params, null, 2)}
            </pre>
          </div>
        </div>

        {/* Footer Actions */}
        <div style={footerStyle}>
          <button onClick={onClose} style={cancelBtnStyle}>
            Close
          </button>
          <button onClick={() => onReplicate(experiment)} style={submitBtnStyle}>
            ⚙️ Tái lập vào Strategy Builder
          </button>
        </div>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const overlayStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
  backgroundColor: 'rgba(0,0,0,0.7)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 99999,
  fontFamily: 'system-ui, sans-serif',
};

const modalStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  border: '1px solid #334155',
  borderRadius: '8px',
  width: '95%',
  maxWidth: '520px',
  padding: '1.5rem',
  boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.4)',
  color: '#e2e8f0',
};

const modalHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'flex-start',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.75rem',
  marginBottom: '1rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '1rem',
  fontWeight: '700',
  color: '#f8fafc',
  margin: 0,
};

const hashStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  fontFamily: 'monospace',
};

const closeBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#94a3b8',
  fontSize: '1.1rem',
  cursor: 'pointer',
  outline: 'none',
};

const bodyStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1.25rem',
  maxHeight: '380px',
  overflowY: 'auto',
  paddingRight: '4px',
};

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const sectionTitleStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  fontWeight: '700',
  color: '#94a3b8',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  margin: 0,
};

const metaGridStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.4rem',
  backgroundColor: '#1e293b',
  border: '1px solid #334155',
  borderRadius: '6px',
  padding: '0.6rem 0.75rem',
};

const metaRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.75rem',
};

const metaLabelStyle: React.CSSProperties = {
  color: '#64748b',
};

const metaValueStyle: React.CSSProperties = {
  color: '#cbd5e1',
  fontWeight: '600',
};

const versionGridStyle: React.CSSProperties = {
  display: 'flex',
  flexWrap: 'wrap',
  gap: '0.5rem',
};

const versionBadgeStyle: React.CSSProperties = {
  display: 'flex',
  backgroundColor: '#1e293b',
  border: '1px solid #334155',
  borderRadius: '4px',
  overflow: 'hidden',
  fontSize: '0.7rem',
};

const versionKeyStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  padding: '0.2rem 0.4rem',
  color: '#94a3b8',
  fontWeight: '600',
};

const versionValStyle: React.CSSProperties = {
  padding: '0.2rem 0.4rem',
  color: '#06b6d4',
  fontWeight: '700',
  fontFamily: 'monospace',
};

const jsonCodeStyle: React.CSSProperties = {
  backgroundColor: '#070a13',
  border: '1px solid #1e293b',
  borderRadius: '6px',
  padding: '0.75rem',
  fontSize: '0.75rem',
  color: '#34d399', // Greenish code
  margin: 0,
  fontFamily: 'monospace',
  overflowX: 'auto',
};

const footerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'flex-end',
  gap: '0.75rem',
  borderTop: '1px solid #1e293b',
  paddingTop: '1rem',
  marginTop: '1rem',
};

const cancelBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#94a3b8',
  border: '1px solid #475569',
  borderRadius: '6px',
  padding: '0.5rem 1rem',
  fontSize: '0.8rem',
  fontWeight: '600',
  cursor: 'pointer',
  outline: 'none',
};

const submitBtnStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6',
  color: '#ffffff',
  border: 'none',
  borderRadius: '6px',
  padding: '0.5rem 1rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 4px rgba(59, 130, 246, 0.2)',
  outline: 'none',
};
