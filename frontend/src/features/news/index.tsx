import { useEffect, useState } from 'react';
import { ErrorBoundary } from '../../shared/components';
import { NewsCrawlerHeader } from './components/NewsCrawlerHeader';
import { NewsInputList } from './components/NewsInputList';
import { SentimentAnalyticsPanel } from './components/SentimentAnalyticsPanel';
import { MOCK_NEWS_FEED } from './services/mockNewsData';
import type { NewsItem, SentimentObservation } from '../../types/news';
import { fetchSentimentObservations } from '../../shared/api';
import { useAppMode } from '../../shared/auth';

function toNewsItem(observation: SentimentObservation): NewsItem {
  const isDemoRSS = (observation.source || '').toLowerCase().includes('demo rss');
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
    analysisSource: isDemoRSS ? 'DEMO' : 'LIVE',
  };
}

export function NewsCrawlerDashboard() {
  const mode = useAppMode();
  const [newsFeed, setNewsFeed] = useState<NewsItem[]>(mode === 'DEMO' ? MOCK_NEWS_FEED : []);
  const [observations, setObservations] = useState<SentimentObservation[]>([]);
  const [analysisStatus, setAnalysisStatus] = useState<'LOADING' | 'DEMO' | 'LIVE' | 'EMPTY' | 'ERROR'>(mode === 'DEMO' ? 'DEMO' : 'LOADING');
  const [analysisMessage, setAnalysisMessage] = useState(mode === 'DEMO'
    ? 'Showing explicitly selected offline sample articles.'
    : 'Loading analyzed articles from the live API…');

  useEffect(() => {
    if (mode === 'DEMO') return;
    fetchSentimentObservations()
      .then((real) => {
        setObservations(real);
        if (real.length === 0) {
          setAnalysisStatus('EMPTY');
          setAnalysisMessage('No analyzed RSS articles in the last 24 hours. Run news-ingest; no demo data has been substituted.');
          return;
        }
        setNewsFeed(real.map(toNewsItem));
        setAnalysisStatus('LIVE');
        const demoCount = real.filter(item => (item.source || '').toLowerCase().includes('demo rss')).length;
        setAnalysisMessage(`Showing ${real.length} API-backed article(s) from the last 24h${demoCount ? ` (${demoCount} clearly labelled demo RSS fixture)` : ''}.`);
      })
      .catch((error) => {
        setAnalysisStatus('ERROR');
        setAnalysisMessage(`Could not load the live news feed: ${error instanceof Error ? error.message : String(error)}. No demo data has been substituted.`);
      });
  }, [mode]);

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
