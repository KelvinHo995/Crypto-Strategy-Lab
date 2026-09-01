import type { MarketInfo } from '../../../types/candle';

const symbols = [
  ['BTCUSDT', 'BTC'],
  ['ETHUSDT', 'ETH'],
  ['BNBUSDT', 'BNB'],
  ['SOLUSDT', 'SOL'],
  ['XRPUSDT', 'XRP'],
  ['ADAUSDT', 'ADA'],
  ['DOGEUSDT', 'DOGE'],
  ['AVAXUSDT', 'AVAX'],
] as const;

export const DEFAULT_MARKETS: MarketInfo[] = symbols.map(([symbol, baseAsset]) => ({
  symbol,
  baseAsset,
  quoteAsset: 'USDT',
  timeframes: ['5m', '15m', '1h', '4h'],
}));
