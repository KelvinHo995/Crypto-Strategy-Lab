import { useEffect, useState } from 'react';
import { ErrorBoundary } from '../../shared/components';
import { NewsCrawlerHeader } from './components/NewsCrawlerHeader';
import { NewsInputList } from './components/NewsInputList';
import { SentimentAnalyticsPanel } from './components/SentimentAnalyticsPanel';
import { MOCK_NEWS_FEED } from './services/mockNewsData';
import type { NewsItem, SentimentObservation } from '../../types/news';
import { fetchSentimentObservations } from '../../shared/api';

function toNewsItem(observation: SentimentObservation): NewsItem {
  return {
    id: observation.newsId,
    title: observation.title || observation.newsId,
    content: '',
    source: observation.source || 'RSS',
    url: observation.url,
    publishedAt: observation.publishedAt,
    sentiment: {
      newsId: observation.newsId,
      sentiment: observation.sentiment,
      score: observation.score,
      model: { name: observation.modelName, version: observation.modelVersion },
      createdAt: observation.analyzedAt,
    },
    analysisSource: 'LIVE',
  };
}

export function NewsCrawlerDashboard() {
  const [newsFeed, setNewsFeed] = useState<NewsItem[]>(MOCK_NEWS_FEED);
  const [observations, setObservations] = useState<SentimentObservation[]>([]);
  const [analysisStatus, setAnalysisStatus] = useState<'DEMO' | 'LIVE' | 'ERROR'>('DEMO');
  const [analysisMessage, setAnalysisMessage] = useState('Showing local sample articles — waiting on the live ingestion pipeline.');

  useEffect(() => {
    // Real, already-ingested articles (via the RSS ingestion pipeline) — if
    // there are none yet, the demo fixtures stay so the page isn't empty.
    // A failed fetch is not shown as an error here either: this is a
    // background load, not something the user triggered.
    fetchSentimentObservations()
      .then((real) => {
        setObservations(real);
        if (real.length === 0) return;
        setNewsFeed(real.map(toNewsItem));
        setAnalysisStatus('LIVE');
        setAnalysisMessage(`Showing ${real.length} real analyzed article(s) from the last 24h.`);
      })
      .catch(() => undefined);
  }, []);

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
        <NewsCrawlerHeader />

        {/* 2-Column Layout: real feed + real sentiment breakdown */}
        <div style={gridStyle}>
          <div style={columnStyle}>
            <NewsInputList news={newsFeed} />
          </div>
          <div style={columnStyle}>
            <SentimentAnalyticsPanel observations={observations} />
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
