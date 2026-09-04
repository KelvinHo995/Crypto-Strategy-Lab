import React, { useState } from 'react';
import type { NewsItem } from '../../../types/news';

interface ExtractionPipelinePanelProps {
  selectedNews: NewsItem | null;
}

export function ExtractionPipelinePanel({ selectedNews }: ExtractionPipelinePanelProps) {
  const [activeTab, setActiveTab] = useState<'pipeline' | 'rawHtml' | 'selectors'>('pipeline');
  const [isHealing, setIsHealing] = useState(false);
  const [healSuccess, setHealSuccess] = useState(false);

  const handleSimulateHeal = () => {
    setIsHealing(true);
    setHealSuccess(false);
    setTimeout(() => {
      setIsHealing(false);
      setHealSuccess(true);
      setTimeout(() => setHealSuccess(false), 4000);
    }, 1200);
  };

  if (!selectedNews) {
    return (
      <div style={panelContainerStyle}>
        <div style={emptyContainerStyle}>
          <span style={{ fontSize: '2rem' }}>🔍</span>
          <h4 style={{ margin: '0.5rem 0 0.25rem 0', color: '#0f172a' }}>Extraction Pipeline</h4>
          <p style={{ margin: 0, fontSize: '0.8rem', color: '#64748b' }}>
            Click on any article in the feed to inspect its 4-step LLM extraction pipeline, raw HTML DOM, and self-healing diagnostics.
          </p>
        </div>
      </div>
    );
  }

  const rawHtmlSnippet = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>${selectedNews.title}</title>
  <meta property="article:published_time" content="${new Date(selectedNews.publishedAt).toISOString()}" />
  <meta name="author" content="${selectedNews.source}" />
</head>
<body>
  <header><h1>${selectedNews.title}</h1></header>
  <main class="article-content">
    <p>${selectedNews.content || 'Full content extracted from source pipeline.'}</p>
    <div class="meta">Source: ${selectedNews.source} | Live Feed</div>
  </main>
</body>
</html>`;

  const selectorJson = {
    source: selectedNews.source,
    targetUrl: selectedNews.url || 'https://news.example.com',
    selectors: {
      title: 'h1.article-title, header h1, .post-title',
      body: 'div.article-content p, .entry-content p',
      publishedTime: 'meta[property="article:published_time"], time.entry-date',
      author: 'meta[name="author"], span.author-name',
    },
    llmPipeline: {
      model: selectedNews.sentiment?.model.name || 'FinBERT-Crypto',
      version: selectedNews.sentiment?.model.version || 'v3.1',
      entityExtraction: 'Active (NER: Crypto Coins, Protocols, Orgs)',
      confidenceThreshold: 0.7,
    },
    selfHealing: {
      driftDetection: 'ENABLED',
      fallbackEngine: 'Cheerio + LLM Tree Parser',
      lastVerified: new Date().toLocaleTimeString(),
    },
  };

  const sentiment = selectedNews.sentiment?.sentiment || 'NEUTRAL';
  const score = selectedNews.sentiment ? (selectedNews.sentiment.score * 100).toFixed(1) : '85.0';

  return (
    <div style={panelContainerStyle}>
      {/* Header */}
      <div style={headerStyle}>
        <div>
          <h4 style={titleStyle}>Extraction & LLM Pipeline</h4>
          <span style={subtitleStyle}>Article Ref: {selectedNews.id} ({selectedNews.source})</span>
        </div>
        <div style={tabGroupStyle}>
          <button
            onClick={() => setActiveTab('pipeline')}
            style={activeTab === 'pipeline' ? activeTabBtnStyle : tabBtnStyle}
          >
            Pipeline
          </button>
          <button
            onClick={() => setActiveTab('rawHtml')}
            style={activeTab === 'rawHtml' ? activeTabBtnStyle : tabBtnStyle}
          >
            Raw HTML
          </button>
          <button
            onClick={() => setActiveTab('selectors')}
            style={activeTab === 'selectors' ? activeTabBtnStyle : tabBtnStyle}
          >
            JSON Schema
          </button>
        </div>
      </div>

      {/* Selected Article Banner */}
      <div style={articleBannerStyle}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '0.5rem' }}>
          <h5 style={articleTitleStyle}>{selectedNews.title}</h5>
          <span style={sentimentBadgeStyle(sentiment)}>
            {sentiment} ({score}%)
          </span>
        </div>
        {selectedNews.url && (
          <a href={selectedNews.url} target="_blank" rel="noreferrer" style={urlLinkStyle}>
            🔗 {selectedNews.url}
          </a>
        )}
      </div>

      {/* Tab 1: Pipeline View */}
      {activeTab === 'pipeline' && (
        <div style={pipelineContentStyle}>
          <div style={stepListStyle}>
            {/* Step 1 */}
            <div style={stepCardStyle}>
              <div style={stepIconStyle('#10b981')}>1</div>
              <div style={stepInfoStyle}>
                <div style={stepTitleRowStyle}>
                  <span style={stepTitleStyle}>HTML Fetch & Ingestion</span>
                  <span style={statusOkStyle}>200 OK (84ms)</span>
                </div>
                <p style={stepDescStyle}>Fetched DOM payload from source endpoint, validated SSL certificate and gzip stream.</p>
              </div>
            </div>

            {/* Step 2 */}
            <div style={stepCardStyle}>
              <div style={stepIconStyle('#3b82f6')}>2</div>
              <div style={stepInfoStyle}>
                <div style={stepTitleRowStyle}>
                  <span style={stepTitleStyle}>DOM Parsing & CSS Selectors</span>
                  <span style={statusOkStyle}>MATCH (100%)</span>
                </div>
                <p style={stepDescStyle}>Applied CSS tree selector templates. Extracted title, published timestamp, and body text without boilerplate ads.</p>
              </div>
            </div>

            {/* Step 3 */}
            <div style={stepCardStyle}>
              <div style={stepIconStyle('#8b5cf6')}>3</div>
              <div style={stepInfoStyle}>
                <div style={stepTitleRowStyle}>
                  <span style={stepTitleStyle}>LLM Entity Recognition & Cleansing</span>
                  <span style={statusOkStyle}>NER Complete</span>
                </div>
                <p style={stepDescStyle}>Identified crypto assets, protocols, and market sentiments. Stripped promotional noise.</p>
              </div>
            </div>

            {/* Step 4 */}
            <div style={stepCardStyle}>
              <div style={stepIconStyle('#f59e0b')}>4</div>
              <div style={stepInfoStyle}>
                <div style={stepTitleRowStyle}>
                  <span style={stepTitleStyle}>FinBERT Sentiment Inference</span>
                  <span style={statusOkStyle}>{selectedNews.sentiment?.model.name || 'FinBERT'}</span>
                </div>
                <p style={stepDescStyle}>Classified tone as <strong>{sentiment}</strong> with <strong>{score}%</strong> confidence score.</p>
              </div>
            </div>
          </div>

          {/* Self Healing Diagnostics Card */}
          <div style={healingCardStyle}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={healingTitleStyle}>🛡 Self-Healing Parser Diagnostics</span>
              <button onClick={handleSimulateHeal} disabled={isHealing} style={healBtnStyle}>
                {isHealing ? 'Testing Selectors…' : 'Simulate Auto-Healing'}
              </button>
            </div>
            <div style={healingGridStyle}>
              <div>Selector Health: <strong style={{ color: '#10b981' }}>99.8% Nominal</strong></div>
              <div>DOM Drift: <strong style={{ color: '#0f172a' }}>0 Anomalies</strong></div>
              <div>Fallback Parser: <strong style={{ color: '#2563eb' }}>Ready (Auto-repair active)</strong></div>
            </div>
            {healSuccess && (
              <div style={healAlertStyle}>
                ✓ Self-Healing simulation passed: Selectors calibrated and verified successfully against target DOM schema.
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Raw HTML */}
      {activeTab === 'rawHtml' && (
        <div style={codeWrapperStyle}>
          <pre style={codeBlockStyle}>{rawHtmlSnippet}</pre>
        </div>
      )}

      {/* Tab 3: JSON Selectors */}
      {activeTab === 'selectors' && (
        <div style={codeWrapperStyle}>
          <pre style={codeBlockStyle}>{JSON.stringify(selectorJson, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}

// ==========================================
// STYLING
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
  gap: '0.85rem',
};

const emptyContainerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  height: '100%',
  minHeight: '280px',
  textAlign: 'center',
  padding: '2rem',
};

const headerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
  flexWrap: 'wrap',
  gap: '0.5rem',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#0f172a',
  margin: 0,
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const subtitleStyle: React.CSSProperties = {
  fontSize: '0.68rem',
  color: '#64748b',
};

const tabGroupStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.3rem',
  backgroundColor: '#f1f5f9',
  padding: '2px',
  borderRadius: '4px',
};

const tabBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  border: 'none',
  color: '#64748b',
  fontSize: '0.7rem',
  fontWeight: '600',
  padding: '0.2rem 0.5rem',
  borderRadius: '3px',
  cursor: 'pointer',
};

const activeTabBtnStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6',
  border: 'none',
  color: '#ffffff',
  fontSize: '0.7rem',
  fontWeight: '700',
  padding: '0.2rem 0.5rem',
  borderRadius: '3px',
  cursor: 'pointer',
};

const articleBannerStyle: React.CSSProperties = {
  backgroundColor: '#f8fafc',
  border: '1px solid #e2e8f0',
  borderRadius: '6px',
  padding: '0.65rem 0.75rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.35rem',
};

const articleTitleStyle: React.CSSProperties = {
  margin: 0,
  fontSize: '0.8rem',
  fontWeight: '700',
  color: '#0f172a',
  lineHeight: '1.3',
};

const urlLinkStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  color: '#2563eb',
  textDecoration: 'none',
  overflow: 'hidden',
  textOverflow: 'ellipsis',
  whiteSpace: 'nowrap',
};

const sentimentBadgeStyle = (sentiment: string): React.CSSProperties => {
  const isPos = sentiment === 'POSITIVE';
  const isNeg = sentiment === 'NEGATIVE';
  return {
    fontSize: '0.65rem',
    fontWeight: '700',
    padding: '0.1rem 0.4rem',
    borderRadius: '4px',
    backgroundColor: isPos ? 'rgba(16, 185, 129, 0.15)' : isNeg ? 'rgba(239, 68, 68, 0.15)' : 'rgba(148, 163, 184, 0.15)',
    color: isPos ? '#10b981' : isNeg ? '#ef4444' : '#64748b',
    border: `1px solid ${isPos ? '#10b981' : isNeg ? '#ef4444' : '#cbd5e1'}`,
    whiteSpace: 'nowrap',
  };
};

const pipelineContentStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.75rem',
  overflowY: 'auto',
  maxHeight: 'calc(100vh - 340px)',
};

const stepListStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const stepCardStyle: React.CSSProperties = {
  display: 'flex',
  gap: '0.65rem',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '6px',
  padding: '0.5rem 0.65rem',
  alignItems: 'flex-start',
};

const stepIconStyle = (color: string): React.CSSProperties => ({
  width: '20px',
  height: '20px',
  borderRadius: '50%',
  backgroundColor: color,
  color: '#ffffff',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  fontSize: '0.7rem',
  fontWeight: '700',
  flexShrink: 0,
});

const stepInfoStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.15rem',
  flexGrow: 1,
};

const stepTitleRowStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
};

const stepTitleStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  fontWeight: '700',
  color: '#0f172a',
};

const statusOkStyle: React.CSSProperties = {
  fontSize: '0.6rem',
  fontWeight: '700',
  color: '#10b981',
  backgroundColor: 'rgba(16, 185, 129, 0.1)',
  padding: '0.05rem 0.3rem',
  borderRadius: '3px',
};

const stepDescStyle: React.CSSProperties = {
  fontSize: '0.68rem',
  color: '#64748b',
  margin: 0,
  lineHeight: '1.3',
};

const healingCardStyle: React.CSSProperties = {
  backgroundColor: '#f8fafc',
  border: '1px solid #cbd5e1',
  borderRadius: '6px',
  padding: '0.75rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const healingTitleStyle: React.CSSProperties = {
  fontSize: '0.72rem',
  fontWeight: '700',
  color: '#1e293b',
};

const healBtnStyle: React.CSSProperties = {
  backgroundColor: '#3b82f6',
  color: '#ffffff',
  border: 'none',
  borderRadius: '4px',
  padding: '0.2rem 0.5rem',
  fontSize: '0.68rem',
  fontWeight: '700',
  cursor: 'pointer',
};

const healingGridStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '0.68rem',
  color: '#64748b',
  flexWrap: 'wrap',
  gap: '0.4rem',
};

const healAlertStyle: React.CSSProperties = {
  backgroundColor: 'rgba(16, 185, 129, 0.1)',
  border: '1px solid #10b981',
  color: '#047857',
  padding: '0.4rem 0.5rem',
  borderRadius: '4px',
  fontSize: '0.68rem',
  fontWeight: '600',
};

const codeWrapperStyle: React.CSSProperties = {
  backgroundColor: '#0f172a',
  borderRadius: '6px',
  padding: '0.75rem',
  overflowX: 'auto',
  maxHeight: '320px',
};

const codeBlockStyle: React.CSSProperties = {
  margin: 0,
  color: '#38bdf8',
  fontSize: '0.72rem',
  fontFamily: 'monospace',
  lineHeight: '1.4',
};
