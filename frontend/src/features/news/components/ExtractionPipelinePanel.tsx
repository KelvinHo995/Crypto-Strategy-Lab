import { useState } from 'react';
import {
  MOCK_EXTRACTION_TEMPLATE,
  MOCK_SELF_HEALING_STATS,
} from '../services/mockNewsData';

export function ExtractionPipelinePanel() {
  const [autoHeal, setAutoHeal] = useState(MOCK_SELF_HEALING_STATS.autoHealEnabled);
  const [templateVersion, setTemplateVersion] = useState(MOCK_EXTRACTION_TEMPLATE.version);
  const [isApplying, setIsApplying] = useState(false);

  const handleApplyTemplate = () => {
    setIsApplying(true);
    setTimeout(() => {
      setTemplateVersion('v1.4.3');
      setIsApplying(false);
      alert('✓ Self-healing success: Applied new Extraction Template version v1.4.3. Expected error rate reduced to 4.1%!');
    }, 1500);
  };

  return (
    <div style={panelContainerStyle}>
      {/* SECTION 1: LLM-ASSISTED EXTRACTION */}
      <div style={sectionStyle}>
        <div style={sectionHeaderStyle}>
          <h4 style={titleStyle}>LLM-Assisted Extraction</h4>
          <span style={versionBadgeStyle}>Template: {templateVersion} ✓</span>
        </div>

        {/* 4-step flowchart */}
        <div style={flowContainerStyle}>
          <div style={flowStepStyle}>1. Raw HTML</div>
          <div style={flowArrowStyle}>➔</div>
          <div style={flowStepStyle}>2. Tag Parsing</div>
          <div style={flowArrowStyle}>➔</div>
          <div style={flowStepStyle}>3. Gen JSON</div>
          <div style={flowArrowStyle}>➔</div>
          <div style={flowStepStyle}>4. Save v{templateVersion.slice(1)}</div>
        </div>

        {/* Code previews carousel/tabs split */}
        <div style={codeSplitGridStyle}>
          <div>
            <div style={codeHeaderStyle}>Raw HTML Snippet</div>
            <pre style={codePreviewStyle}>{MOCK_EXTRACTION_TEMPLATE.rawHtmlPreview}</pre>
          </div>
          <div>
            <div style={codeHeaderStyle}>Generated JSON Template</div>
            <pre style={codePreviewStyle}>{MOCK_EXTRACTION_TEMPLATE.jsonTemplatePreview}</pre>
          </div>
        </div>

        {/* Metrics */}
        <div style={metricFlexStyle}>
          <div style={metricBoxStyle}>
            <span style={metricLabelStyle}>LLM Confidence</span>
            <span style={metricValueStyle}>{(MOCK_EXTRACTION_TEMPLATE.confidenceScore * 100).toFixed(0)}%</span>
          </div>
          <div style={metricBoxStyle}>
            <span style={metricLabelStyle}>Mapped Fields</span>
            <span style={{ ...metricValueStyle, color: '#06b6d4' }}>{MOCK_EXTRACTION_TEMPLATE.extractedFields} / 5</span>
          </div>
        </div>
      </div>

      {/* SECTION 2: SELF-HEALING EXTRACTION */}
      <div style={{ ...sectionStyle, borderTop: '1px solid #1e293b', paddingTop: '1rem' }}>
        <div style={sectionHeaderStyle}>
          <h4 style={titleStyle}>Self-Healing Pipeline</h4>
          <label style={switchLabelStyle}>
            <input
              type="checkbox"
              checked={autoHeal}
              onChange={() => setAutoHeal(!autoHeal)}
              style={checkboxStyle}
            />
            <span style={switchTextStyle}>Auto-healing</span>
          </label>
        </div>

        {/* 4-step self-healing flowchart */}
        <div style={flowContainerStyle}>
          <div style={{ ...flowStepStyle, color: '#f59e0b', borderColor: '#d97706' }}>1. Validate</div>
          <div style={flowArrowStyle}>➔</div>
          <div style={{ ...flowStepStyle, color: '#ef4444', borderColor: '#dc2626' }}>2. Alert (&gt;10%)</div>
          <div style={flowArrowStyle}>➔</div>
          <div style={{ ...flowStepStyle, color: '#3b82f6', borderColor: '#2563eb' }}>3. LLM Repair</div>
          <div style={flowArrowStyle}>➔</div>
          <div style={{ ...flowStepStyle, color: '#10b981', borderColor: '#059669' }}>4. Commit v1.4.3</div>
        </div>

        {/* Error rates table */}
        <div style={errorCardStyle}>
          <div style={errorTitleStyle}>Extraction Error Diagnostics (24h)</div>
          <div style={errorGridStyle}>
            <div style={errorColStyle}>
              <span style={errorLabelStyle}>Empty Fields</span>
              <span style={errorValueStyle}>{MOCK_SELF_HEALING_STATS.emptyFieldsPct}%</span>
            </div>
            <div style={errorColStyle}>
              <span style={errorLabelStyle}>Wrong Format</span>
              <span style={errorValueStyle}>{MOCK_SELF_HEALING_STATS.formatErrorsPct}%</span>
            </div>
            <div style={errorColStyle}>
              <span style={errorLabelStyle}>Total Failures</span>
              <span style={{ ...errorValueStyle, color: '#ef4444' }}>{MOCK_SELF_HEALING_STATS.totalErrorsPct}%</span>
            </div>
          </div>
        </div>

        {/* Healing action block */}
        {templateVersion === 'v1.4.2' && (
          <div style={healingProposalCardStyle}>
            <div style={proposalHeaderStyle}>
              <span>🔧 Proposed Template Version (v1.4.3)</span>
              <span style={{ color: '#10b981', fontSize: '0.7rem' }}>Expected error: 4.1%</span>
            </div>
            <pre style={proposalCodeStyle}>{MOCK_SELF_HEALING_STATS.proposedTemplate}</pre>
            <button
              onClick={handleApplyTemplate}
              disabled={isApplying}
              style={isApplying ? activeApplyBtnStyle : applyBtnStyle}
            >
              {isApplying ? 'Applying Template...' : '✓ Áp dụng ngay (v1.4.3)'}
            </button>
          </div>
        )}
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

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
};

const sectionHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#cbd5e1',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const versionBadgeStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  backgroundColor: 'rgba(16, 185, 129, 0.15)',
  border: '1px solid #10b981',
  color: '#10b981',
  padding: '0.15rem 0.4rem',
  borderRadius: '4px',
  fontWeight: '700',
};

const flowContainerStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  backgroundColor: '#070a13',
  padding: '0.4rem 0.5rem',
  borderRadius: '6px',
  border: '1px solid #1e293b',
};

const flowStepStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  color: '#94a3b8',
  border: '1px solid #334155',
  padding: '0.15rem 0.35rem',
  borderRadius: '4px',
};

const flowArrowStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#475569',
};

const codeSplitGridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
  gap: '0.5rem',
};

const codeHeaderStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
  fontWeight: '600',
  marginBottom: '0.2rem',
};

const codePreviewStyle: React.CSSProperties = {
  backgroundColor: '#070a13',
  border: '1px solid #1e293b',
  borderRadius: '4px',
  padding: '0.5rem',
  color: '#06b6d4',
  fontSize: '0.65rem',
  margin: 0,
  fontFamily: 'monospace',
  maxHeight: '110px',
  overflow: 'auto',
};

const metricFlexStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.5rem',
};

const metricBoxStyle: React.CSSProperties = {
  flex: 1,
  backgroundColor: '#1e293b',
  borderRadius: '6px',
  padding: '0.4rem',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

const metricLabelStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
};

const metricValueStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#cbd5e1',
};

const switchLabelStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.35rem',
  cursor: 'pointer',
};

const checkboxStyle: React.CSSProperties = {
  cursor: 'pointer',
};

const switchTextStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#94a3b8',
  fontWeight: '600',
};

const errorCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(239, 68, 68, 0.03)',
  border: '1px solid rgba(239, 68, 68, 0.15)',
  borderRadius: '6px',
  padding: '0.5rem',
};

const errorTitleStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  fontWeight: '700',
  color: '#ef4444',
  textTransform: 'uppercase',
  marginBottom: '0.4rem',
  letterSpacing: '0.03em',
};

const errorGridStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
};

const errorColStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  flex: 1,
};

const errorLabelStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#64748b',
};

const errorValueStyle: React.CSSProperties = {
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#cbd5e1',
};

const healingProposalCardStyle: React.CSSProperties = {
  backgroundColor: '#070a13',
  border: '1px solid #1e293b',
  borderRadius: '6px',
  padding: '0.5rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const proposalHeaderStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.7rem',
  fontWeight: '600',
  color: '#cbd5e1',
};

const proposalCodeStyle: React.CSSProperties = {
  backgroundColor: '#070a13',
  color: '#f59e0b',
  fontSize: '0.65rem',
  margin: 0,
  fontFamily: 'monospace',
  maxHeight: '100px',
  overflow: 'auto',
  border: '1px dashed #334155',
  padding: '0.4rem',
  borderRadius: '4px',
};

const applyBtnStyle: React.CSSProperties = {
  backgroundColor: '#10b981',
  color: '#064e3b',
  border: 'none',
  borderRadius: '4px',
  padding: '0.4rem',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'pointer',
  outline: 'none',
  boxShadow: '0 2px 4px rgba(16, 185, 129, 0.2)',
};

const activeApplyBtnStyle: React.CSSProperties = {
  backgroundColor: '#1e293b',
  color: '#475569',
  border: '1px solid #334155',
  borderRadius: '4px',
  padding: '0.4rem',
  fontSize: '0.75rem',
  fontWeight: '700',
  cursor: 'not-allowed',
  outline: 'none',
};
