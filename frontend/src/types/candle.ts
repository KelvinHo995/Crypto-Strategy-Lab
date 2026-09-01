export interface Candle {
  symbol?: string;
  timeframe?: string;
  openTime: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  isClosed?: boolean;
}

export interface MarketInfo {
  symbol: string;
  baseAsset: string;
  quoteAsset: string;
  timeframes: string[];
}

export interface TradeTick {
  symbol: string;
  tradeId: number;
  tradeTime: number;
  price: number;
  quantity: number;
  side: 'BUY' | 'SELL';
}
