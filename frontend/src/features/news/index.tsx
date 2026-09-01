import { useState } from 'react';
import { ErrorBoundary } from '../../shared/components';
import { NewsCrawlerHeader } from './components/NewsCrawlerHeader';
import { NewsInputList } from './components/NewsInputList';
import { ExtractionPipelinePanel } from './components/ExtractionPipelinePanel';
import { SentimentAnalyticsPanel } from './components/SentimentAnalyticsPanel';
import { MOCK_NEWS_FEED } from './services/mockNewsData';
import type { NewsItem } from '../../types/news';
import { analyzeSentiment } from '../../shared/api';

export function NewsCrawlerDashboard() {
  const [newsFeed, setNewsFeed] = useState<NewsItem[]>(MOCK_NEWS_FEED);
  const [isCrawling, setIsCrawling] = useState(false);
  const [analysisStatus, setAnalysisStatus] = useState<'DEMO' | 'LIVE' | 'ERROR'>('DEMO');
  const [analysisMessage, setAnalysisMessage] = useState('Sample articles are local fixtures; analyze a sample to verify the live backend pipeline.');

  const handleCrawlStart = async () => {
    setIsCrawling(true);
    setAnalysisMessage('Sending sample article through Go API, FastAPI model, and PostgreSQL...');
    const publishedAt = Date.now();
    const newId = `live-news-${publishedAt}`;
    try {
      const observation = await analyzeSentiment(
        newId,
        'Bitcoin records bullish gains after major regulatory approval and strong institutional inflow.',
        publishedAt,
      );
      const newArticle: NewsItem = {
        id: newId,
        title: 'Bitcoin institutional inflow strengthens after regulatory approval',
        content: 'Bitcoin records bullish gains after major regulatory approval and strong institutional inflow.',
        source: 'Live pipeline sample',
        publishedAt,
        sentiment: {
          newsId: newId,
          sentiment: observation.sentiment,
          score: observation.score,
          model: { name: observation.modelName, version: observation.modelVersion },
          createdAt: observation.analyzedAt,
        },
        analysisSource: 'LIVE',
      };
      setNewsFeed((prev) => [newArticle, ...prev]);
      setAnalysisStatus('LIVE');
      setAnalysisMessage(`Stored ${observation.newsId} with ${observation.modelName}/${observation.modelVersion}.`);
    } catch (error) {
      setAnalysisStatus('ERROR');
      setAnalysisMessage(`Live sentiment failed: ${String(error)}`);
    } finally {
      setIsCrawling(false);
    }
  };

  return (
    <ErrorBoundary
      fallback={
        <div style={degradedContainerStyle}>
          <h4>News Crawler Feed Unavailable</h4>
          <p>The Sentiment Model or Crawler Service is currently undergoing self-healing. Rest of the Strategy Lab remains operational.</p>
          <button onClick={() => window.location.reload()} style={retryBtnStyle}>Retry Connection</button>
        </div>
      }
    >
      <div style={dashboardContainerStyle}>
        <div style={analysisStatus === 'LIVE' ? liveStatusStyle : analysisStatus === 'ERROR' ? errorStatusStyle : demoStatusStyle}>
          <strong>{analysisStatus}</strong> {analysisMessage}
        </div>
        {/* Top Controls Bar */}
        <NewsCrawlerHeader onCrawlStart={handleCrawlStart} isCrawling={isCrawling} />

        {/* 3-Column Layout */}
        <div style={gridStyle}>
          {/* Column 1: Input feed list */}
          <div style={columnStyle}>
            <NewsInputList news={newsFeed} />
          </div>

          {/* Column 2: HTML extraction pipeline */}
          <div style={columnStyle}>
            <ExtractionPipelinePanel />
          </div>

          {/* Column 3: Output sentiment analytics */}
          <div style={columnStyle}>
            <SentimentAnalyticsPanel />
          </div>
        </div>
      </div>
    </ErrorBoundary>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const dashboardContainerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1.25rem',
  width: '100%',
  boxSizing: 'border-box',
};

const statusBaseStyle: React.CSSProperties = { fontSize: '0.75rem', padding: '0.65rem 0.8rem', borderRadius: '8px', border: '1px solid' };
const demoStatusStyle: React.CSSProperties = { ...statusBaseStyle, color: '#92400e', background: '#fffbeb', borderColor: '#fde68a' };
const liveStatusStyle: React.CSSProperties = { ...statusBaseStyle, color: '#047857', background: '#ecfdf5', borderColor: '#a7f3d0' };
const errorStatusStyle: React.CSSProperties = { ...statusBaseStyle, color: '#b91c1c', background: '#fef2f2', borderColor: '#fecaca' };

const gridStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1rem',
  alignItems: 'stretch',
  flexWrap: 'wrap', // Responsive wrapping on smaller screens
};

const columnStyle: React.CSSProperties = {
  flex: '1 1 320px', // Min-width 320px for columns, expands equally
  display: 'flex',
  flexDirection: 'column',
};

const degradedContainerStyle: React.CSSProperties = {
  padding: '2rem',
  margin: '2rem auto',
  maxWidth: '480px',
  textAlign: 'center',
  backgroundColor: '#111827',
  border: '1px dashed #f59e0b',
  borderRadius: '8px',
  color: '#ffffff',
  fontFamily: 'system-ui, sans-serif',
};

const retryBtnStyle: React.CSSProperties = {
  backgroundColor: '#f59e0b',
  color: '#000000',
  border: 'none',
  padding: '0.5rem 1rem',
  borderRadius: '4px',
  cursor: 'pointer',
  fontWeight: 'bold',
  marginTop: '1rem',
};
