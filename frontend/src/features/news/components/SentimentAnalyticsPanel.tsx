import type { SentimentObservation } from '../../../types/news';
import { useExperimentStore } from '../../../shared/stores/useExperimentStore';

interface SentimentAnalyticsPanelProps {
  observations: SentimentObservation[];
}

export function SentimentAnalyticsPanel({ observations }: SentimentAnalyticsPanelProps) {
  const injectStrategyToBuilder = useExperimentStore((state) => state.injectStrategyToBuilder);

  const total = observations.length;
  const counts = { POSITIVE: 0, NEUTRAL: 0, NEGATIVE: 0 };
  for (const observation of observations) {
    if (observation.sentiment === 'POSITIVE' || observation.sentiment === 'NEGATIVE' || observation.sentiment === 'NEUTRAL') {
      counts[observation.sentiment]++;
    }
  }
  const pct = (n: number) => (total > 0 ? Math.round((n / total) * 100) : 0);

  const handleApplyToBuilder = () => {
    injectStrategyToBuilder({
      type: 'Sentiment',
      params: { sentimentThreshold: 0.7 },
      weight: 0.5,
    });
  };

  return (
    <div style={panelContainerStyle}>
      {/* 24h Sentiment Breakdown */}
      <div style={sectionStyle}>
        <h4 style={titleStyle}>Sentiment Breakdown ({total} article{total === 1 ? '' : 's'})</h4>

        {total === 0 ? (
          <p style={emptyStyle}>
            No analyzed articles in this window yet — live observations will stream in automatically.
          </p>
        ) : (
          <div style={gaugeAreaStyle}>
            <div style={gaugeLabelsStyle}>
              <span style={{ color: '#10b981' }}>Positive ({pct(counts.POSITIVE)}%)</span>
              <span style={{ color: '#64748b' }}>Neutral ({pct(counts.NEUTRAL)}%)</span>
              <span style={{ color: '#ef4444' }}>Negative ({pct(counts.NEGATIVE)}%)</span>
            </div>
            <div style={gaugeBarContainerStyle}>
              <div
                style={{ ...gaugeFillStyle, width: `${pct(counts.POSITIVE)}%`, backgroundColor: '#10b981' }}
                title="Positive"
              />
              <div
                style={{ ...gaugeFillStyle, width: `${pct(counts.NEUTRAL)}%`, backgroundColor: '#94a3b8' }}
                title="Neutral"
              />
              <div
                style={{ ...gaugeFillStyle, width: `${pct(counts.NEGATIVE)}%`, backgroundColor: '#ef4444' }}
                title="Negative"
              />
            </div>
          </div>
        )}
      </div>

      {/* Integration Card with Action Button */}
      <div style={integrationCardStyle}>
        <h4 style={{ ...titleStyle, color: '#2563eb', borderBottom: 'none', paddingBottom: 0 }}>
          Used by Strategy Engine
        </h4>
        <p style={pipeDescStyle}>
          The <strong>Sentiment</strong> plugin acts as a live market regime filter in composite strategies. Configure threshold confidence and combine it with MA, RSI, or Bollinger Bands.
        </p>
        <button
          type="button"
          onClick={handleApplyToBuilder}
          style={applyBtnStyle}
          title="Add Sentiment strategy instance to Strategy Builder"
        >
          🚀 Apply Sentiment to Builder
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

const sectionStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#0f172a',
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
  backgroundColor: '#e2e8f0',
};

const gaugeFillStyle: React.CSSProperties = {
  height: '100%',
  transition: 'width 0.3s ease',
};

const integrationCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(37, 99, 235, 0.04)',
  border: '1px solid rgba(37, 99, 235, 0.2)',
  borderRadius: '8px',
  padding: '0.85rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.65rem',
  marginTop: 'auto',
};

const pipeDescStyle: React.CSSProperties = {
  fontSize: '0.72rem',
  color: '#64748b',
  lineHeight: '1.4',
  margin: 0,
};

const applyBtnStyle: React.CSSProperties = {
  backgroundColor: '#2563eb',
  color: '#ffffff',
  border: 'none',
  borderRadius: '6px',
  padding: '0.55rem 0.85rem',
  fontSize: '0.8rem',
  fontWeight: '700',
  cursor: 'pointer',
  boxShadow: '0 2px 4px rgba(37, 99, 235, 0.2)',
  transition: 'all 0.15s ease',
  textAlign: 'center',
};
