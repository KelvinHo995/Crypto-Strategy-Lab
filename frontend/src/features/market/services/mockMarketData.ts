import type { Candle } from '../../../types/candle';
import type { ChartMarker, SRZone } from '../components/TradingChart';

export interface MarketDataDTO {
  candles: Candle[];
  ma20Line: number[];
  bbands: {
    upper: number[];
    basis: number[];
    lower: number[];
  };
  srZones: SRZone[];
  markers: ChartMarker[];
}

/**
 * Returns timeframe duration in milliseconds.
 */
export function getTimeframeMs(timeframe: string): number {
  const num = parseInt(timeframe.slice(0, -1), 10);
  const unit = timeframe.slice(-1);

  switch (unit) {
    case 'm': return num * 60 * 1000;
    case 'h': return num * 60 * 60 * 1000;
    case 'd': return num * 24 * 60 * 60 * 1000;
    default: return 5 * 60 * 1000; // default 5m
  }
}

/**
 * Helper to calculate Moving Average locally for mockup/DTO construction
 */
function calculateMA(closePrices: number[], period: number): number[] {
  const ma: number[] = [];
  for (let i = 0; i < closePrices.length; i++) {
    if (i < period - 1) {
      ma.push(NaN);
      continue;
    }
    let sum = 0;
    for (let j = 0; j < period; j++) {
      sum += closePrices[i - j];
    }
    ma.push(sum / period);
  }
  return ma;
}

/**
 * Helper to calculate Bollinger Bands locally for mockup/DTO construction
 */
function calculateBBands(
  closePrices: number[],
  maBasis: number[],
  period: number,
  stdDevMultiplier: number = 2
): { upper: number[]; basis: number[]; lower: number[] } {
  const upper: number[] = [];
  const lower: number[] = [];

  for (let i = 0; i < closePrices.length; i++) {
    if (i < period - 1 || isNaN(maBasis[i])) {
      upper.push(NaN);
      lower.push(NaN);
      continue;
    }

    const basis = maBasis[i];
    let varianceSum = 0;
    for (let j = 0; j < period; j++) {
      varianceSum += Math.pow(closePrices[i - j] - basis, 2);
    }
    const stdDev = Math.sqrt(varianceSum / period);

    upper.push(basis + stdDevMultiplier * stdDev);
    lower.push(basis - stdDevMultiplier * stdDev);
  }

  return { upper, basis: maBasis, lower };
}

/**
 * Fetches market data DTO. Mimics backend REST endpoint `/candles?symbol=...&timeframe=...`
 */
export function fetchMarketDataDTO(symbol: string, timeframe: string, count: number = 200): MarketDataDTO {
  const candles = generateHistoricalCandles(symbol, timeframe, count);
  const closePrices = candles.map(c => c.close);

  // Calculate indicators for DTO representation
  const ma20Line = calculateMA(closePrices, 20);
  const bbands = calculateBBands(closePrices, ma20Line, 20, 2);

  // Simulate S/R Zones (using historical low/high as zones)
  const sortedClose = [...closePrices].sort((a, b) => a - b);
  const lowPrice = sortedClose[Math.floor(sortedClose.length * 0.1)];
  const highPrice = sortedClose[Math.floor(sortedClose.length * 0.9)];
  
  const srZones: SRZone[] = [
    { price: Number(lowPrice.toFixed(2)), type: 'SUPPORT' },
    { price: Number(highPrice.toFixed(2)), type: 'RESISTANCE' }
  ];

  // Generate Buy/Sell Signals based on simple crossing rules (strictly for visualization demo)
  const markers: ChartMarker[] = [];
  for (let i = 21; i < candles.length; i++) {
    const prevCandle = candles[i - 1];
    const currCandle = candles[i];
    const prevMA = ma20Line[i - 1];
    const currMA = ma20Line[i];

    if (isNaN(prevMA) || isNaN(currMA)) continue;

    // Cross up MA20 -> Buy signal
    if (prevCandle.close <= prevMA && currCandle.close > currMA && Math.random() > 0.6) {
      markers.push({
        time: currCandle.openTime,
        position: 'belowBar',
        color: '#10b981',
        shape: 'arrowUp',
        text: 'BUY',
      });
    } 
    // Cross down MA20 -> Sell signal
    else if (prevCandle.close >= prevMA && currCandle.close < currMA && Math.random() > 0.6) {
      markers.push({
        time: currCandle.openTime,
        position: 'aboveBar',
        color: '#ef4444',
        shape: 'arrowDown',
        text: 'SELL',
      });
    }
  }

  return {
    candles,
    ma20Line,
    bbands,
    srZones,
    markers
  };
}

/**
 * Generates historical candlestick data for UI testing.
 */
function generateHistoricalCandles(
  symbol: string,
  timeframe: string,
  count: number = 300
): Candle[] {
  const candles: Candle[] = [];
  let basePrice = 50000;
  if (symbol.includes('ETH')) basePrice = 3000;
  else if (symbol.includes('SOL')) basePrice = 140;
  else if (symbol.includes('BNB')) basePrice = 580;

  const timeframeMs = getTimeframeMs(timeframe);
  let currentTime = Date.now() - count * timeframeMs;
  let currentClose = basePrice;

  for (let i = 0; i < count; i++) {
    const volatility = 0.002; 
    const change = currentClose * volatility * (Math.random() - 0.495); 
    const open = currentClose;
    const close = open + change;
    
    const high = Math.max(open, close) + Math.random() * open * 0.001;
    const low = Math.min(open, close) - Math.random() * open * 0.001;
    const volume = Math.random() * 50 + (symbol.includes('BTC') ? 10 : 5);

    candles.push({
      symbol,
      timeframe,
      openTime: currentTime,
      open,
      high,
      low,
      close,
      volume,
      isClosed: true,
    });

    currentClose = close;
    currentTime += timeframeMs;
  }

  return candles;
}

/**
 * Generates a mock realtime tick based on the last candle.
 */
export function generateNextTick(lastCandle: Candle): Candle {
  const timeframeMs = getTimeframeMs(lastCandle.timeframe || '1m');
  const now = Date.now();
  const shouldCreateNew = now >= lastCandle.openTime + timeframeMs;
  
  let openTime = lastCandle.openTime;
  let open = lastCandle.open;
  let basePrice = lastCandle.close;
  
  if (shouldCreateNew) {
    openTime = lastCandle.openTime + timeframeMs;
    open = lastCandle.close;
    basePrice = lastCandle.close;
  }

  const volatility = 0.0005; 
  const priceChange = basePrice * volatility * (Math.random() - 0.5);
  const close = basePrice + priceChange;
  
  let high: number;
  let low: number;

  if (shouldCreateNew) {
    high = Math.max(open, close);
    low = Math.min(open, close);
  } else {
    high = Math.max(lastCandle.high, close);
    low = Math.min(lastCandle.low, close);
  }

  const volume = (shouldCreateNew ? 0 : lastCandle.volume) + Math.random() * 1.5;

  return {
    symbol: lastCandle.symbol,
    timeframe: lastCandle.timeframe,
    openTime,
    open,
    high,
    low,
    close,
    volume,
    isClosed: shouldCreateNew,
  };
}
