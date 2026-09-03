import type { StrategyInfo } from '../../../types/strategy';

export interface SingleStrategyInstance {
  id: string;
  name: string;
  type: string; // "RSI" | "MA" | "Bollinger" | "SR" | "SMC"
  description: string;
  params: Record<string, unknown>;
  currentSignal: 'BUY' | 'SELL' | 'HOLD';
}

export interface CompositePreset {
  name: string;
  strategies: string[];
  weights: Record<string, number>;
  policy: string;
}

export interface DiscoveryStats {
  iteration: number;
  totalIterations: number;
  testedCandidates: number;
  status: 'IDLE' | 'RUNNING' | 'COMPLETED' | 'STOPPED' | 'FAILED';
  searchId?: string;
  statusMessage?: string;
  bestStrategy?: {
    name: string;
    profit: number;
    winrate: number;
    mdd: number;
  };
}

export interface MiniLeaderboardItem {
  rank: number;
  name: string;
  profit: number;
  winrate: number;
}

// 1. Definition of available strategies in Backend Registry (Plugin metadata)
export const AVAILABLE_STRATEGIES_META: StrategyInfo[] = [
  {
    name: 'RSI',
    description: 'Relative Strength Index - Momentum oscillator tracking overbought (>70) and oversold (<30) conditions.',
    parameters: {
      rsiPeriod: { type: 'number', default: 14, description: 'Number of lookback candles' },
      rsiOverbought: { type: 'number', default: 70, description: 'Overbought boundary' },
      rsiOversold: { type: 'number', default: 30, description: 'Oversold boundary' },
    },
  },
  {
    name: 'MA',
    description: 'Moving Average crossover - a fast window crossing above/below a slower window signals a trend shift.',
    parameters: {
      maShortWindow: { type: 'number', default: 20, description: 'Fast MA period' },
      maLongWindow: { type: 'number', default: 50, description: 'Slow MA period' },
    },
  },
  {
    name: 'Bollinger',
    description: 'Bollinger Bands - Volatility bands placed above and below a moving average base line.',
    parameters: {
      bollingerPeriod: { type: 'number', default: 20, description: 'Bands time period' },
      bollingerStdDev: { type: 'number', default: 2, description: 'Standard deviation multiplier' },
    },
  },
  {
    name: 'SR',
    description: 'Support / Resistance zones - Traces structural high/low pivots to detect bounces or breakouts.',
    parameters: {
      srWindow: { type: 'number', default: 20, description: 'Pivot lookback window' },
      srTolerance: { type: 'number', default: 0.005, description: 'Proximity tolerance to a pivot (e.g. 0.005 = 0.5%)' },
    },
  },
  {
    name: 'SMC',
    description: 'Smart Money Concepts - Minimal structure break indicator detecting order blocks and swing levels.',
    parameters: {
      smcLookback: { type: 'number', default: 10, description: 'Swing high/low lookback window' },
    },
  },
];

// 2. Default instantiated single strategies (Active indicators in the workspace)
export const DEFAULT_SINGLE_STRATEGIES: SingleStrategyInstance[] = [
  {
    id: 'rsi-14',
    name: 'RSI (14)',
    type: 'RSI',
    description: 'Overbought 70, Oversold 30',
    params: { rsiPeriod: 14, rsiOverbought: 70, rsiOversold: 30 },
    currentSignal: 'BUY',
  },
  {
    id: 'ma-20',
    name: 'MA (10/20)',
    type: 'MA',
    description: 'Fast crossover, 10/20 period',
    params: { maShortWindow: 10, maLongWindow: 20 },
    currentSignal: 'HOLD',
  },
  {
    id: 'ma-50',
    name: 'MA (20/50)',
    type: 'MA',
    description: 'Slower crossover, 20/50 period',
    params: { maShortWindow: 20, maLongWindow: 50 },
    currentSignal: 'BUY',
  },
  {
    id: 'bb-20-2',
    name: 'Bollinger Bands (20, 2)',
    type: 'Bollinger',
    description: 'BB period 20, multiplier 2.0',
    params: { bollingerPeriod: 20, bollingerStdDev: 2 },
    currentSignal: 'SELL',
  },
  {
    id: 'sr-3',
    name: 'Support & Resistance',
    type: 'SR',
    description: 'Structural pivots, window 20',
    params: { srWindow: 20, srTolerance: 0.005 },
    currentSignal: 'HOLD',
  },
];

// 3. Combination Presets
export const COMPOSITE_PRESETS: CompositePreset[] = [
  {
    name: 'MA + RSI Trend',
    strategies: ['ma-20', 'rsi-14'],
    weights: { 'ma-20': 0.5, 'rsi-14': 0.5 },
    policy: 'majority',
  },
  {
    name: 'Volatility Breakout (RSI + BB)',
    strategies: ['rsi-14', 'bb-20-2'],
    weights: { 'rsi-14': 0.4, 'bb-20-2': 0.6 },
    policy: 'weighted',
  },
  {
    name: 'Triple Confirm (MA + RSI + S/R)',
    strategies: ['ma-20', 'rsi-14', 'sr-3'],
    weights: { 'ma-20': 0.3, 'rsi-14': 0.4, 'sr-3': 0.3 },
    policy: 'weighted',
  },
];

// 4. Discovery initial state; LIVE progress is populated only by backend events.
export const DEFAULT_DISCOVERY_STATS: DiscoveryStats = {
  iteration: 0,
  totalIterations: 0,
  testedCandidates: 0,
  status: 'IDLE',
};

// 5. Mock Mini Leaderboard top-5
export const MINI_LEADERBOARD_DATA: MiniLeaderboardItem[] = [
  { rank: 1, name: 'MA(20) + RSI(14) + S/R (Weighted)', profit: 2342.18, winrate: 68.21 },
  { rank: 2, name: 'RSI(14) + BB(20,2) (Majority)', profit: 1890.54, winrate: 61.45 },
  { rank: 3, name: 'SMC Breakout + MA(50)', profit: 1420.12, winrate: 59.10 },
  { rank: 4, name: 'MA(20) + MA(50) Crossover', profit: 980.45, winrate: 55.32 },
  { rank: 5, name: 'Support/Resistance Pivot Only', profit: 540.22, winrate: 51.18 },
];
