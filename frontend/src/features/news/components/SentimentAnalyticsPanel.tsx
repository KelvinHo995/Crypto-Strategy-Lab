import type { SentimentObservation } from '../../../types/news';

interface SentimentAnalyticsPanelProps {
  observations: SentimentObservation[];
}

export function SentimentAnalyticsPanel({ observations }: SentimentAnalyticsPanelProps) {
  const total = observations.length;
  const counts = { POSITIVE: 0, NEUTRAL: 0, NEGATIVE: 0 };
  for (const observation of observations) counts[observation.sentiment]++;
  const pct = (n: number) => (total > 0 ? Math.round((n / total) * 100) : 0);

  return (
    <div style={panelContainerStyle}>
      <div style={sectionStyle}>
        <h4 style={titleStyle}>Sentiment Breakdown ({total} article{total === 1 ? '' : 's'})</h4>

        {total === 0 ? (
          <p style={emptyStyle}>
            No analyzed articles in this window yet — try analyzing a sample article.
          </p>
        ) : (
          <div style={gaugeAreaStyle}>
            <div style={gaugeLabelsStyle}>
              <span style={{ color: '#10b981' }}>Positive ({pct(counts.POSITIVE)}%)</span>
              <span style={{ color: '#94a3b8' }}>Neutral ({pct(counts.NEUTRAL)}%)</span>
              <span style={{ color: '#ef4444' }}>Negative ({pct(counts.NEGATIVE)}%)</span>
            </div>
            <div style={gaugeBarContainerStyle}>
              <div style={{ ...gaugeFillStyle, width: `${pct(counts.POSITIVE)}%`, backgroundColor: '#10b981' }} title="Positive" />
              <div style={{ ...gaugeFillStyle, width: `${pct(counts.NEUTRAL)}%`, backgroundColor: '#94a3b8' }} title="Neutral" />
              <div style={{ ...gaugeFillStyle, width: `${pct(counts.NEGATIVE)}%`, backgroundColor: '#ef4444' }} title="Negative" />
            </div>
          </div>
        )}
      </div>

      {/* Informational only — Sentiment is already a real, selectable strategy
          in the Strategy Engine (Registry "Sentiment"); there's no separate
          one-click wiring to do, this just points at where it lives. */}
      <div style={integrationCardStyle}>
        <h4 style={{ ...titleStyle, color: '#2563eb' }}>Used by the Strategy Engine</h4>
        <p style={pipeDescStyle}>
          Sentiment analysis already feeds a real strategy — select <strong>Sentiment</strong> as one of your
          composite instances on the Strategies page to use live news sentiment as an entry condition.
        </p>
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

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.85rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#475569',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
};

const emptyStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  lineHeight: '1.4',
  margin: 0,
};

const gaugeAreaStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.4rem',
  backgroundColor: '#f8fafc',
  padding: '0.6rem',
  borderRadius: '6px',
  border: '1px solid #e2e8f0',
};

const gaugeLabelsStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.7rem',
  fontWeight: '700',
};

const gaugeBarContainerStyle: React.CSSProperties = {
  height: '8px',
  borderRadius: '9999px',
  overflow: 'hidden',
  display: 'flex',
  width: '100%',
};

const gaugeFillStyle: React.CSSProperties = {
  height: '100%',
};

const integrationCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(6, 182, 212, 0.03)',
  border: '1px solid rgba(6, 182, 212, 0.2)',
  borderRadius: '8px',
  padding: '0.75rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const pipeDescStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  lineHeight: '1.4',
  margin: 0,
};
