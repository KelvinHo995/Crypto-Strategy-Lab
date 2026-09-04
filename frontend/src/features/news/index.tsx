import { useCallback, useEffect, useMemo, useState } from 'react';
import { ErrorBoundary } from '../../shared/components';
import { NewsCrawlerHeader, type AssetFilter, type NewsSourceTab } from './components/NewsCrawlerHeader';
import { NewsInputList } from './components/NewsInputList';
import { ExtractionPipelinePanel } from './components/ExtractionPipelinePanel';
import { SentimentAnalyticsPanel } from './components/SentimentAnalyticsPanel';
import { SourceConfigModal } from './components/SourceConfigModal';
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
  const [activeSource, setActiveSource] = useState<NewsSourceTab>('RSS Feeds');
  const [activeAsset, setActiveAsset] = useState<AssetFilter>('ALL');
  const [autoRefreshInterval, setAutoRefreshInterval] = useState<number>(60000); // 1m default
  const [selectedNews, setSelectedNews] = useState<NewsItem | null>(() => MOCK_NEWS_FEED[0] ?? null);
  const [isSourceModalOpen, setIsSourceModalOpen] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [analysisStatus, setAnalysisStatus] = useState<'DEMO' | 'LIVE' | 'ERROR'>('DEMO');
  const [analysisMessage, setAnalysisMessage] = useState('Showing local sample articles — waiting on live ingestion pipeline.');

  const applyObservations = useCallback((real: SentimentObservation[]) => {
    setObservations(real);
    if (real.length > 0) {
      const liveItems = real.map(toNewsItem);
      setNewsFeed(liveItems);
      setSelectedNews((prev) => (prev && liveItems.some((i) => i.id === prev.id) ? prev : liveItems[0]));
      setAnalysisStatus('LIVE');
      setAnalysisMessage(`Showing ${real.length} real analyzed article(s) from live pipeline.`);
    }
  }, []);

  const handleManualRefresh = useCallback(async () => {
    setIsRefreshing(true);
    try {
      const real = await fetchSentimentObservations();
      applyObservations(real);
    } catch {
      // Fallback gracefully
    } finally {
      setIsRefreshing(false);
    }
  }, [applyObservations]);

  // Initial load
  useEffect(() => {
    fetchSentimentObservations().then(applyObservations).catch(() => undefined);
  }, [applyObservations]);

  // Auto-refresh timer
  useEffect(() => {
    if (autoRefreshInterval <= 0) return;
    const timer = setInterval(() => {
      fetchSentimentObservations().then(applyObservations).catch(() => undefined);
    }, autoRefreshInterval);
    return () => clearInterval(timer);
  }, [autoRefreshInterval, applyObservations]);

  // Filtered news items
  const filteredNews = useMemo(() => {
    return newsFeed.filter((item) => {
      // 1. Asset Filter
      if (activeAsset !== 'ALL') {
        const text = `${item.title} ${item.content}`.toUpperCase();
        if (activeAsset === 'BTC' && !text.includes('BTC') && !text.includes('BITCOIN')) return false;
        if (activeAsset === 'ETH' && !text.includes('ETH') && !text.includes('ETHEREUM')) return false;
        if (activeAsset === 'SOL' && !text.includes('SOL') && !text.includes('SOLANA')) return false;
        if (activeAsset === 'BNB' && !text.includes('BNB') && !text.includes('BINANCE')) return false;
        if (activeAsset === 'XRP' && !text.includes('XRP') && !text.includes('RIPPLE')) return false;
      }

      // 2. Source Tab Filter
      if (activeSource === 'Website Scraper') {
        const isScraper = item.source.toLowerCase().includes('binance') || item.source.toLowerCase().includes('scraper');
        // If not labeled scraper, allow standard sample to demonstrate functionality
        if (!isScraper && !item.source.toLowerCase().includes('coindesk')) return false;
      } else if (activeSource === 'Raw HTML Import') {
        const isRaw = item.source.toLowerCase().includes('raw') || item.source.toLowerCase().includes('import');
        if (!isRaw && !item.source.toLowerCase().includes('the block')) return false;
      }

      return true;
    });
  }, [newsFeed, activeAsset, activeSource]);

  const handleSelectNews = (item: NewsItem) => {
    setSelectedNews(item);
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
        {/* Status Pill */}
        <div style={analysisStatus === 'LIVE' ? liveStatusStyle : analysisStatus === 'ERROR' ? errorStatusStyle : demoStatusStyle}>
          <strong>{analysisStatus}</strong> {analysisMessage}
        </div>

        {/* Top Controls Bar */}
        <NewsCrawlerHeader
          activeSource={activeSource}
          onSelectSource={setActiveSource}
          activeAsset={activeAsset}
          onSelectAsset={setActiveAsset}
          autoRefreshInterval={autoRefreshInterval}
          onSelectAutoRefresh={setAutoRefreshInterval}
          onOpenSourceModal={() => setIsSourceModalOpen(true)}
          onManualRefresh={handleManualRefresh}
          isRefreshing={isRefreshing}
        />

        {/* 3-Column Layout: Feed + Pipeline Detail + Sentiment Breakdown & Action */}
        <div style={gridStyle}>
          <div style={leftColStyle}>
            <NewsInputList
              news={filteredNews}
              selectedNewsId={selectedNews?.id}
              onSelectNews={handleSelectNews}
            />
          </div>

          <div style={midColStyle}>
            <ExtractionPipelinePanel selectedNews={selectedNews} />
          </div>

          <div style={rightColStyle}>
            <SentimentAnalyticsPanel observations={observations} />
          </div>
        </div>

        {/* Source Configuration Modal */}
        <SourceConfigModal
          isOpen={isSourceModalOpen}
          onClose={() => setIsSourceModalOpen(false)}
        />
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
  gap: '1rem',
  width: '100%',
  boxSizing: 'border-box',
};

const statusBaseStyle: React.CSSProperties = { fontSize: '0.75rem', padding: '0.6rem 0.8rem', borderRadius: '6px', border: '1px solid' };
const demoStatusStyle: React.CSSProperties = { ...statusBaseStyle, color: '#92400e', background: '#fffbeb', borderColor: '#fde68a' };
const liveStatusStyle: React.CSSProperties = { ...statusBaseStyle, color: '#047857', background: '#ecfdf5', borderColor: '#a7f3d0' };
const errorStatusStyle: React.CSSProperties = { ...statusBaseStyle, color: '#b91c1c', background: '#fef2f2', borderColor: '#fecaca' };

const gridStyle: React.CSSProperties = {
  display: 'flex',
  gap: '1rem',
  alignItems: 'stretch',
  flexWrap: 'wrap',
};

const leftColStyle: React.CSSProperties = {
  flex: '1 1 320px',
  display: 'flex',
  flexDirection: 'column',
};

const midColStyle: React.CSSProperties = {
  flex: '1.4 1 380px',
  display: 'flex',
  flexDirection: 'column',
};

const rightColStyle: React.CSSProperties = {
  flex: '1 1 280px',
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
