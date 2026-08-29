import type { NewsItem } from '../../../types/news';

export interface ExtractionTemplateInfo {
  version: string;
  confidenceScore: number;
  extractedFields: number;
  rawHtmlPreview: string;
  jsonTemplatePreview: string;
}

export interface SelfHealingStats {
  autoHealEnabled: boolean;
  emptyFieldsPct: number;
  formatErrorsPct: number;
  totalErrorsPct: number;
  proposedTemplate: string;
}

export interface SentimentOverview {
  positivePct: number;
  neutralPct: number;
  negativePct: number;
  eventsDistribution: { name: string; pct: number }[];
  mlopsMetrics: {
    avgConfidence: number;
    totalAnalyzed: number;
    sourceCoverage: number;
    activeSources: string;
  };
}

// 1. Mock list of 15 realistic cryptocurrency news articles with sentiment analysis DTO metadata
export const MOCK_NEWS_FEED: NewsItem[] = [
  {
    id: 'news-001',
    title: 'SEC Approves Spot Ethereum ETFs, Trading Set to Begin Next Week',
    content: 'The Securities and Exchange Commission (SEC) has officially approved the final registration statements for several spot Ethereum ETFs, clearing the path for trading on major exchanges starting Tuesday.',
    source: 'CoinDesk',
    url: 'https://coindesk.com/ethereum-etf-approval',
    publishedAt: Date.now() - 5 * 60 * 1000, // 5m ago
    sentiment: {
      newsId: 'news-001',
      sentiment: 'POSITIVE',
      score: 0.94,
      model: { name: 'FinBERT-Crypto', version: 'v3.1' },
      createdAt: Date.now() - 4 * 60 * 1000,
    }
  },
  {
    id: 'news-002',
    title: 'Solana Transaction Volumne Surpasses Ethereum Amid Memecoin Activity',
    content: 'Solana’s weekly DEX volume has topped Ethereum’s once again as high network usage and memecoin trading continue to drive traffic and transaction fees on Solana protocols.',
    source: 'The Block',
    url: 'https://theblock.co/solana-dex-volume',
    publishedAt: Date.now() - 15 * 60 * 1000, // 15m ago
    sentiment: {
      newsId: 'news-002',
      sentiment: 'POSITIVE',
      score: 0.81,
      model: { name: 'FinBERT-Crypto', version: 'v3.1' },
      createdAt: Date.now() - 14 * 60 * 1000,
    }
  },
  {
    id: 'news-003',
    title: 'Macro Alert: US Inflation Falls to 2.9%, Boosting Hopes for Fed Rate Cuts',
    content: 'The consumer price index (CPI) rose 2.9% in July from a year ago, the lowest reading since 2021. Economists believe this guarantees a Federal Reserve interest rate cut in September.',
    source: 'Decrypt',
    url: 'https://decrypt.co/us-inflation-falls-2-9',
    publishedAt: Date.now() - 32 * 60 * 1000, // 32m ago
    sentiment: {
      newsId: 'news-003',
      sentiment: 'POSITIVE',
      score: 0.89,
      model: { name: 'FinBERT-Crypto', version: 'v3.1' },
      createdAt: Date.now() - 31 * 60 * 1000,
    }
  },
  {
    id: 'news-004',
    title: 'Crypto Hack: Cross-chain Bridge Exploited for $45 Million in Liquidity',
    content: 'A major cross-chain bridge protocol suffered a severe exploit today, with attackers draining multiple smart contract pools of USDC and Wrapped ETH. The team has paused operations.',
    source: 'Cointelegraph',
    url: 'https://cointelegraph.com/cross-chain-exploit-45m',
    publishedAt: Date.now() - 45 * 60 * 1000, // 45m ago
    sentiment: {
      newsId: 'news-004',
      sentiment: 'NEGATIVE',
      score: 0.97,
      model: { name: 'FinBERT-Crypto', version: 'v3.1' },
      createdAt: Date.now() - 44 * 60 * 1000,
    }
  },
  {
    id: 'news-005',
    title: 'Ethereum Core Devs Discuss Pectra Upgrade Timeline and Scope Change',
    content: 'Core developers met in an All-Core-Devs call to finalize which EIPs will be included in the Pectra hard fork, confirming EIP-7702 is on track for deployment in late Q4.',
    source: 'Bankless',
    url: 'https://bankless.com/pectra-upgrade-eip-7702',
    publishedAt: Date.now() - 70 * 60 * 1000, // 70m ago
    sentiment: {
      newsId: 'news-005',
      sentiment: 'NEUTRAL',
      score: 0.62,
      model: { name: 'RoBERTa-Crypto-Sentiment', version: 'v1.0' },
      createdAt: Date.now() - 69 * 60 * 1000,
    }
  },
  {
    id: 'news-006',
    title: 'Binance Obtains Regulatory Approval to Operate in India Once Again',
    content: 'The Financial Intelligence Unit (FIU) of India has registered Binance as a reporting entity, officially permitting the exchange to offer trading options to Indian citizens after resolving local tax disputes.',
    source: 'CoinDesk',
    publishedAt: Date.now() - 120 * 60 * 1000,
    sentiment: {
      newsId: 'news-006',
      sentiment: 'POSITIVE',
      score: 0.78,
      model: { name: 'FinBERT-Crypto', version: 'v3.1' },
      createdAt: Date.now() - 119 * 60 * 1000,
    }
  },
  {
    id: 'news-007',
    title: 'Bitcoin Mining Difficulty Hits New Record High as Hashrate Expands',
    content: 'Mining metrics show hashrate expansion leading to a 3.4% difficulty increase on BTC blocks, squeezing mining profit margins as hardware expenses increase globally.',
    source: 'The Block',
    publishedAt: Date.now() - 180 * 60 * 1000,
    sentiment: {
      newsId: 'news-007',
      sentiment: 'NEGATIVE',
      score: 0.58,
      model: { name: 'RoBERTa-Crypto-Sentiment', version: 'v1.0' },
      createdAt: Date.now() - 179 * 60 * 1000,
    }
  }
];

// 2. Mock LLM Extraction Template preview
export const MOCK_EXTRACTION_TEMPLATE: ExtractionTemplateInfo = {
  version: 'v1.4.2',
  confidenceScore: 0.92,
  extractedFields: 5,
  rawHtmlPreview: `<!-- Coindesk raw scrap preview -->
<div class="news-article-card">
  <h2 class="title-class-xyz">SEC Approves Spot Ethereum ETFs...</h2>
  <span class="pub-date" data-unix="1723020000">5m ago</span>
  <p class="summary-body">The SEC has approved spot Ethereum ETFs...</p>
  <div class="author-info">By CoinDesk Team</div>
  <a href="/eth-etf-approved" class="link-tag">Read Article</a>
</div>`,
  jsonTemplatePreview: `{
  "template_version": "v1.4.2",
  "selectors": {
    "title": "div.news-article-card > h2.title-class-xyz",
    "summary": "div.news-article-card > p.summary-body",
    "source": "div.author-info",
    "publishedAt": "span.pub-date[data-unix]",
    "url": "a.link-tag[href]"
  },
  "asset_heuristics": ["ETH", "Solana", "BTC"]
}`,
};

// 3. Mock Self-healing loop diagnostics
export const MOCK_SELF_HEALING_STATS: SelfHealingStats = {
  autoHealEnabled: true,
  emptyFieldsPct: 8.7,
  formatErrorsPct: 3.2,
  totalErrorsPct: 11.9,
  proposedTemplate: `{
  "template_version": "v1.4.3",
  "selectors": {
    "title": "div.news-article-card h2",
    "summary": "p.summary-body, div.article-body",
    "source": "div.author-info, span.publisher",
    "publishedAt": "span.pub-date",
    "url": "a.link-tag"
  }
}`,
};

// 4. Mock 24h Sentiment statistics & MLOps quality metrics
export const MOCK_SENTIMENT_OVERVIEW: SentimentOverview = {
  positivePct: 58,
  neutralPct: 27,
  negativePct: 15,
  eventsDistribution: [
    { name: 'ETF & Institutional Flow', pct: 28 },
    { name: 'Protocol Upgrade / Code', pct: 22 },
    { name: 'Regulation & Policy', pct: 15 },
    { name: 'Partnerships & Ecosystem', pct: 12 },
    { name: 'Market Speculation & Trends', pct: 23 },
  ],
  mlopsMetrics: {
    avgConfidence: 0.78,
    totalAnalyzed: 1248,
    sourceCoverage: 92,
    activeSources: '23 / 25',
  },
};
