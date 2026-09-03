export function NewsCrawlerHeader() {
  return (
    <div style={containerStyle}>
      <span style={hintStyle}>
        Articles below are collected, deduplicated, and sentiment-scored by the live news pipeline.
      </span>
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
  padding: '0.75rem 1.25rem',
};

const hintStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
  lineHeight: '1.4',
};
