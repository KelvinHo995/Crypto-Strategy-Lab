export interface SentimentModelInfo {
  name: string;
  version: string;
}

export interface SentimentAnalysis {
  newsId: string;
  sentiment: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
  score: number;
  model: SentimentModelInfo;
  createdAt: number;
}

export interface NewsItem {
  id: string;
  title: string;
  content: string;
  source: string;
  url?: string;
  publishedAt: number; // Unix timestamp in milliseconds
  sentiment?: SentimentAnalysis;
}
