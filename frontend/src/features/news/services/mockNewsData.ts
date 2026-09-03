import type { NewsItem } from '../../../types/news';

// Mock list of realistic cryptocurrency news articles with sentiment analysis DTO metadata
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
