import { MOCK_SENTIMENT_OVERVIEW } from '../services/mockNewsData';

export function SentimentAnalyticsPanel() {
  const data = MOCK_SENTIMENT_OVERVIEW;

  const handleApplyStrategy = () => {
    alert('NewsSentimentStrategy has been integrated!\nThis adds a sentiment threshold filter (score > 0.8 to buy, score < 0.2 to sell) as an entry constraint to the strategy builder.');
  };

  return (
    <div style={panelContainerStyle}>
      {/* KHỐI 1: SENTIMENT ANALYSIS 24H */}
      <div style={sectionStyle}>
        <h4 style={titleStyle}>24H Sentiment Analytics</h4>

        {/* Stacked gauge bar */}
        <div style={gaugeAreaStyle}>
          <div style={gaugeLabelsStyle}>
            <span style={{ color: '#10b981' }}>Positive ({data.positivePct}%)</span>
            <span style={{ color: '#94a3b8' }}>Neutral ({data.neutralPct}%)</span>
            <span style={{ color: '#ef4444' }}>Negative ({data.negativePct}%)</span>
          </div>
          <div style={gaugeBarContainerStyle}>
            <div style={{ ...gaugeFillStyle, width: `${data.positivePct}%`, backgroundColor: '#10b981' }} title="Positive" />
            <div style={{ ...gaugeFillStyle, width: `${data.neutralPct}%`, backgroundColor: '#475569' }} title="Neutral" />
            <div style={{ ...gaugeFillStyle, width: `${data.negativePct}%`, backgroundColor: '#ef4444' }} title="Negative" />
          </div>
        </div>

        {/* Event Type distribution list */}
        <div style={distributionAreaStyle}>
          <span style={subTitleStyle}>Event Type Distribution (Top Categories)</span>
          <div style={distListStyle}>
            {data.eventsDistribution.map((event) => (
              <div key={event.name} style={distRowStyle}>
                <div style={distLabelRowStyle}>
                  <span style={distLabelStyle}>{event.name}</span>
                  <span style={distValStyle}>{event.pct}%</span>
                </div>
                <div style={distBarContainerStyle}>
                  <div style={{ ...distBarFillStyle, width: `${event.pct}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* MLOps Quality Metrics table */}
        <div style={mlopsAreaStyle}>
          <span style={subTitleStyle}>MLOps Model Quality Metrics</span>
          <div style={mlopsGridStyle}>
            <div style={mlopsRowStyle}>
              <span style={mlopsLabelStyle}>Avg Confidence:</span>
              <span style={mlopsValueStyle}>{(data.mlopsMetrics.avgConfidence * 100).toFixed(0)}%</span>
            </div>
            <div style={mlopsRowStyle}>
              <span style={mlopsLabelStyle}>Total Analyzed:</span>
              <span style={mlopsValueStyle}>{data.mlopsMetrics.totalAnalyzed.toLocaleString()} articles</span>
            </div>
            <div style={mlopsRowStyle}>
              <span style={mlopsLabelStyle}>Source Coverage:</span>
              <span style={mlopsValueStyle}>{data.mlopsMetrics.sourceCoverage}%</span>
            </div>
            <div style={mlopsRowStyle}>
              <span style={mlopsLabelStyle}>Active Feed Sources:</span>
              <span style={mlopsValueStyle}>{data.mlopsMetrics.activeSources}</span>
            </div>
          </div>
        </div>
      </div>

      {/* KHỐI 2: STRATEGY ENGINE INTEGRATION */}
      <div style={integrationCardStyle}>
        <h4 style={{ ...titleStyle, color: '#2563eb' }}>Strategy Engine Integration</h4>
        
        {/* Visual pipeline representation */}
        <div style={pipeContainerStyle}>
          <div style={pipeBoxStyle}>News Sentiment</div>
          <div style={pipeLineStyle}>───►</div>
          <div style={pipeBoxActiveStyle}>Strategy Engine</div>
        </div>

        <p style={pipeDescStyle}>
          Allows filtering strategy entry conditions based on live sentiment score. E.g. Block BUY signals if news sentiment is heavily negative.
        </p>

        {/* Strategy plugin trigger */}
        <div style={actionRowStyle}>
          <span style={strategyBadgeStyle}>
            NewsSentimentStrategy
          </span>
          <button onClick={handleApplyStrategy} style={applyBtnStyle}>
            Apply to Builder
          </button>
        </div>
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
  flexGrow: 1,
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#cbd5e1',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
};

const subTitleStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  fontWeight: '600',
  textTransform: 'uppercase',
  letterSpacing: '0.03em',
  display: 'block',
};

const gaugeAreaStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.4rem',
  backgroundColor: '#e2e8f0',
  padding: '0.6rem',
  borderRadius: '6px',
  border: '1px solid #cbd5e1',
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

const distributionAreaStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const distListStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.4rem',
};

const distRowStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.2rem',
};

const distLabelRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.7rem',
};

const distLabelStyle: React.CSSProperties = {
  color: '#cbd5e1',
  fontWeight: '500',
};

const distValStyle: React.CSSProperties = {
  color: '#94a3b8',
  fontWeight: '600',
};

const distBarContainerStyle: React.CSSProperties = {
  height: '4px',
  backgroundColor: '#e2e8f0',
  borderRadius: '9999px',
  overflow: 'hidden',
};

const distBarFillStyle: React.CSSProperties = {
  height: '100%',
  backgroundColor: '#3b82f6',
  borderRadius: '9999px',
};

const mlopsAreaStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
  borderTop: '1px solid #e2e8f0',
  paddingTop: '0.75rem',
};

const mlopsGridStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.4rem',
  backgroundColor: '#ffffff',
  padding: '0.5rem 0.75rem',
  borderRadius: '6px',
  border: '1px solid #e2e8f0',
};

const mlopsRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.7rem',
};

const mlopsLabelStyle: React.CSSProperties = {
  color: '#64748b',
};

const mlopsValueStyle: React.CSSProperties = {
  color: '#cbd5e1',
  fontWeight: '600',
};

const integrationCardStyle: React.CSSProperties = {
  backgroundColor: 'rgba(6, 182, 212, 0.03)',
  border: '1px solid rgba(6, 182, 212, 0.2)',
  borderRadius: '8px',
  padding: '0.75rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.65rem',
};

const pipeContainerStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  gap: '0.5rem',
  backgroundColor: '#ffffff',
  padding: '0.4rem',
  borderRadius: '4px',
};

const pipeBoxStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#94a3b8',
  border: '1px solid #cbd5e1',
  padding: '0.2rem 0.4rem',
  borderRadius: '4px',
};

const pipeBoxActiveStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#ffffff',
  backgroundColor: '#2563eb',
  border: '1px solid #2563eb',
  padding: '0.2rem 0.4rem',
  borderRadius: '4px',
  fontWeight: '700',
};

const pipeLineStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#475569',
};

const pipeDescStyle: React.CSSProperties = {
  fontSize: '0.7rem',
  color: '#64748b',
  lineHeight: '1.4',
  margin: 0,
};

const actionRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  marginTop: '0.25rem',
};

const strategyBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  backgroundColor: 'rgba(6, 182, 212, 0.1)',
  border: '1px solid #2563eb',
  color: '#2563eb',
  padding: '0.2rem 0.4rem',
  borderRadius: '4px',
  fontWeight: '700',
};

const applyBtnStyle: React.CSSProperties = {
  backgroundColor: '#2563eb',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.35rem 0.65rem',
  fontSize: '0.7rem',
  fontWeight: '700',
  cursor: 'pointer',
  outline: 'none',
  boxShadow: '0 2px 4px rgba(6, 182, 212, 0.1)',
};
