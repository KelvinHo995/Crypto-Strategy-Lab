import { useState, useEffect, useMemo, useCallback } from 'react';
import { ErrorBoundary } from '../../shared/components';
import { NewsCrawlerHeader } from './components/NewsCrawlerHeader';
import { NewsInputList } from './components/NewsInputList';
import { ExtractionPipelinePanel, type ExtractionData } from './components/ExtractionPipelinePanel';
import { SentimentAnalyticsPanel } from './components/SentimentAnalyticsPanel';
import { SourceConfigModal } from './components/SourceConfigModal';
import { MOCK_NEWS_FEED, MOCK_EXTRACTION_TEMPLATE } from './services/mockNewsData';
import type { NewsItem } from '../../types/news';
import { analyzeSentiment } from '../../shared/api';

function generateExtractionData(article?: NewsItem): ExtractionData {
  if (!article) {
    return {
      title: 'No article selected',
      source: 'Unknown',
      rawHtml: MOCK_EXTRACTION_TEMPLATE.rawHtmlPreview,
      jsonTemplate: MOCK_EXTRACTION_TEMPLATE.jsonTemplatePreview,
      confidenceScore: MOCK_EXTRACTION_TEMPLATE.confidenceScore,
      extractedFields: MOCK_EXTRACTION_TEMPLATE.extractedFields,
      version: MOCK_EXTRACTION_TEMPLATE.version,
    };
  }

  const tag = article.title.toUpperCase().includes('ETH')
    ? 'ETH'
    : article.title.toUpperCase().includes('SOL')
    ? 'SOL'
    : article.title.toUpperCase().includes('BNB')
    ? 'BNB'
    : 'BTC';

  const rawHtml = `<article class="crypto-news-card" data-source="${article.source.toLowerCase().replace(/\s+/g, '-')}">
  <header class="article-meta">
    <span class="source-tag">${article.source}</span>
    <time datetime="${new Date(article.publishedAt).toISOString()}">${new Date(article.publishedAt).toLocaleDateString()}</time>
  </header>
  <h1 class="headline">${article.title}</h1>
  <div class="article-body">
    <p class="summary">${article.content}</p>
  </div>
  <footer class="tags">
    <a href="/tag/${tag.toLowerCase()}" class="coin-badge">${tag}</a>
  </footer>
</article>`;

  const jsonTemplate = JSON.stringify(
    {
      version: 'v1.4.2',
      targetSource: article.source,
      schema: 'ArticleSentimentDTO',
      selectors: {
        headline: 'h1.headline::text',
        source: 'span.source-tag::text',
        timestamp: 'time::attr(datetime)',
        bodySummary: 'div.article-body > p.summary::text',
        assetTags: 'footer.tags a.coin-badge::text',
      },
      parsedPayload: {
        id: article.id,
        headline: article.title.slice(0, 45) + (article.title.length > 45 ? '...' : ''),
        detectedAsset: tag,
        sentiment: article.sentiment?.sentiment || 'NEUTRAL',
        confidence: article.sentiment?.score || 0.88,
      },
    },
    null,
    2
  );

  return {
    title: article.title,
    source: article.source,
    rawHtml,
    jsonTemplate,
    confidenceScore: article.sentiment?.score || 0.92,
    extractedFields: 5,
    version: 'v1.4.2',
  };
}

export function NewsCrawlerDashboard() {
  const [newsFeed, setNewsFeed] = useState<NewsItem[]>(MOCK_NEWS_FEED);
  const [selectedArticleId, setSelectedArticleId] = useState<string | null>(
    () => MOCK_NEWS_FEED[0]?.id || null
  );

  const [isCrawling, setIsCrawling] = useState(false);
  const [analysisStatus, setAnalysisStatus] = useState<'DEMO' | 'LIVE' | 'ERROR'>('DEMO');
  const [analysisMessage, setAnalysisMessage] = useState(
    'Sample articles are local fixtures; analyze a sample to verify the live backend pipeline.'
  );

  // Filter & Header states
  const [activeSource, setActiveSource] = useState<'website' | 'rss' | 'html'>('website');
  const [activeAsset, setActiveAsset] = useState<string>('ALL');
  const [refreshInterval, setRefreshInterval] = useState<string>('2m');
  const [isSourceConfigOpen, setIsSourceConfigOpen] = useState(false);

  const handleCrawlStart = useCallback(async () => {
    setIsCrawling(true);
    setAnalysisMessage('Sending sample article through Go API, FastAPI model, and PostgreSQL...');
    const publishedAt = Date.now();
    const newId = `live-news-${publishedAt}`;
    try {
      const observation = await analyzeSentiment(
        newId,
        'Bitcoin records bullish gains after major regulatory approval and strong institutional inflow.',
        publishedAt
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
      setSelectedArticleId(newId);
      setAnalysisStatus('LIVE');
      setAnalysisMessage(`Stored ${observation.newsId} with ${observation.modelName}/${observation.modelVersion}.`);
    } catch (error) {
      setAnalysisStatus('ERROR');
      setAnalysisMessage(`Live sentiment failed: ${String(error)}`);
    } finally {
      setIsCrawling(false);
    }
  }, []);

  // Auto-Refresh Timer
  useEffect(() => {
    if (refreshInterval === 'off') return;
    const intervalMs =
      refreshInterval === '1m' ? 60_000 : refreshInterval === '2m' ? 120_000 : 300_000;

    const timer = window.setInterval(() => {
      console.log(`[Auto-Refresh ${refreshInterval}] Polling news updates...`);
      void handleCrawlStart();
    }, intervalMs);

    return () => window.clearInterval(timer);
  }, [refreshInterval, handleCrawlStart]);

  // Asset & Source Filtering
  const filteredNews = useMemo(() => {
    return newsFeed.filter((item) => {
      // 1. Asset keyword matching
      if (activeAsset !== 'ALL') {
        const searchKey = activeAsset.toUpperCase();
        const assetKeywords: Record<string, string[]> = {
          BTC: ['BTC', 'BITCOIN'],
          ETH: ['ETH', 'ETHEREUM', 'ETHER'],
          SOL: ['SOL', 'SOLANA'],
          BNB: ['BNB', 'BINANCE'],
        };
        const keywords = assetKeywords[searchKey] || [searchKey];
        const searchTarget = `${item.title} ${item.content} ${item.source}`.toUpperCase();
        const matchesAsset = keywords.some((kw) => searchTarget.includes(kw));
        if (!matchesAsset) return false;
      }

      // 2. Source category matching
      if (activeSource === 'rss') {
        const isRss =
          item.source.toLowerCase().includes('rss') ||
          item.url?.includes('/rss') ||
          item.source === 'CoinDesk' ||
          item.source === 'The Block';
        return isRss;
      } else if (activeSource === 'html') {
        return (
          item.source.toLowerCase().includes('raw') ||
          item.source.toLowerCase().includes('html') ||
          item.source === 'Live pipeline sample'
        );
      }

      return true;
    });
  }, [newsFeed, activeAsset, activeSource]);

  // Active article & extraction data for Column 2
  const activeArticle = useMemo(() => {
    if (selectedArticleId) {
      const found = newsFeed.find((item) => item.id === selectedArticleId);
      if (found) return found;
    }
    return filteredNews[0] || newsFeed[0];
  }, [newsFeed, filteredNews, selectedArticleId]);

  const activeExtractionData = useMemo(() => {
    return generateExtractionData(activeArticle);
  }, [activeArticle]);

  return (
    <ErrorBoundary
      fallback={
        <div style={degradedContainerStyle}>
          <h4>News Crawler Feed Unavailable</h4>
          <p>
            The Sentiment Model or Crawler Service is currently undergoing self-healing. Rest of the Strategy Lab
            remains operational.
          </p>
          <button type="button" onClick={() => window.location.reload()} style={retryBtnStyle}>
            Retry Connection
          </button>
        </div>
      }
    >
      <div style={dashboardContainerStyle}>
        <div
          style={
            analysisStatus === 'LIVE'
              ? liveStatusStyle
              : analysisStatus === 'ERROR'
              ? errorStatusStyle
              : demoStatusStyle
          }
        >
          <strong>{analysisStatus}</strong> {analysisMessage}
        </div>

        {/* Top Controls Bar with Filters and Auto Refresh */}
        <NewsCrawlerHeader
          onCrawlStart={handleCrawlStart}
          isCrawling={isCrawling}
          activeSource={activeSource}
          onSourceChange={setActiveSource}
          activeAsset={activeAsset}
          onAssetChange={setActiveAsset}
          refreshInterval={refreshInterval}
          onRefreshIntervalChange={setRefreshInterval}
          onOpenSourceConfig={() => setIsSourceConfigOpen(true)}
        />

        {/* 3-Column Layout */}
        <div style={gridStyle}>
          {/* Column 1: Input feed list with active filters and selection */}
          <div style={columnStyle}>
            <NewsInputList
              news={filteredNews}
              selectedArticleId={activeArticle?.id || selectedArticleId}
              onSelectArticle={setSelectedArticleId}
            />
          </div>

          {/* Column 2: HTML extraction pipeline connected to active article */}
          <div style={columnStyle}>
            <ExtractionPipelinePanel activeExtractionData={activeExtractionData} />
          </div>

          {/* Column 3: Output sentiment analytics connected to strategy engine */}
          <div style={columnStyle}>
            <SentimentAnalyticsPanel />
          </div>
        </div>

        {/* Source Configuration Modal */}
        <SourceConfigModal
          isOpen={isSourceConfigOpen}
          onClose={() => setIsSourceConfigOpen(false)}
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
  gap: '1.25rem',
  width: '100%',
  boxSizing: 'border-box',
};

const statusBaseStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  padding: '0.65rem 0.8rem',
  borderRadius: '8px',
  border: '1px solid',
};
const demoStatusStyle: React.CSSProperties = {
  ...statusBaseStyle,
  color: '#92400e',
  background: '#fffbeb',
  borderColor: '#fde68a',
};
const liveStatusStyle: React.CSSProperties = {
  ...statusBaseStyle,
  color: '#047857',
  background: '#ecfdf5',
  borderColor: '#a7f3d0',
};
const errorStatusStyle: React.CSSProperties = {
  ...statusBaseStyle,
  color: '#b91c1c',
  background: '#fef2f2',
  borderColor: '#fecaca',
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
