import { useState } from 'react';
import { ErrorBoundary } from '../../shared/components';
import { NewsCrawlerHeader } from './components/NewsCrawlerHeader';
import { NewsInputList } from './components/NewsInputList';
import { ExtractionPipelinePanel } from './components/ExtractionPipelinePanel';
import { SentimentAnalyticsPanel } from './components/SentimentAnalyticsPanel';
import { MOCK_NEWS_FEED } from './services/mockNewsData';
import type { NewsItem } from '../../types/news';

export function NewsCrawlerDashboard() {
  const [newsFeed, setNewsFeed] = useState<NewsItem[]>(MOCK_NEWS_FEED);
  const [isCrawling, setIsCrawling] = useState(false);

  const handleCrawlStart = () => {
    setIsLoadingCrawl();
  };

  const setIsLoadingCrawl = () => {
    setIsCrawling(true);

    // Simulate live crawling and sentiment extraction
    setTimeout(() => {
      const newId = `news-${Date.now().toString().slice(-3)}`;
      const randomCoin = Math.random() > 0.5 ? 'SOL' : 'BTC';

      const newArticle: NewsItem = {
        id: newId,
        title: `${randomCoin} Whale Accumlates over $25 Million in Private Wallet Transactions`,
        content: `On-chain data indicates a prominent trader address has accumulated massive quantities of ${randomCoin} over the last 48 hours, suggesting positive sentiment.`,
        source: 'The Block',
        publishedAt: Date.now(),
        sentiment: {
          newsId: newId,
          sentiment: 'POSITIVE',
          score: 0.87,
          model: { name: 'FinBERT-Crypto', version: 'v3.1' },
          createdAt: Date.now(),
        }
      };

      setNewsFeed((prev) => [newArticle, ...prev]);
      setIsCrawling(false);
      alert(`Crawl success: Extracted 1 new article [ID: ${newId}]. Sentiment analysis score: Positive (87%)`);
    }, 2000);
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
        <div style={{color:'#f59e0b',fontSize:'0.75rem'}}>DEMO: News collector/sentiment feed chưa có endpoint production trong MVP.</div>
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
