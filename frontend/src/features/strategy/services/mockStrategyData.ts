import type { StrategyInfo } from '../../../types/strategy';

export interface SingleStrategyInstance {
  id: string;
  name: string;
  type: string; // "RSI" | "MA" | "Bollinger" | "SR" | "SMC"
  description: string;
  params: Record<string, any>;
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
  status: 'IDLE' | 'RUNNING' | 'PAUSED' | 'COMPLETED';
  bestStrategy: {
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
    description: 'Moving Average - Trend-following indicator tracing averages over a sliding candle window.',
    parameters: {
      maWindow: { type: 'number', default: 20, description: 'MA period window' },
    },
  },
  {
    name: 'Bollinger',
    description: 'Bollinger Bands - Volatility bands placed above and below a moving average base line.',
    parameters: {
      bbPeriod: { type: 'number', default: 20, description: 'Bands time period' },
      bbStdDev: { type: 'number', default: 2, description: 'Standard deviation multiplier' },
    },
  },
  {
    name: 'SupportResistance',
    description: 'Support / Resistance zones - Traces structural high/low pivots to detect bounces or breakouts.',
    parameters: {
      srSensitivity: { type: 'number', default: 3, description: 'Pivot search strength' },
    },
  },
  {
    name: 'SMC',
    description: 'Smart Money Concepts - Minimal structure break indicator detecting order blocks and swing levels.',
    parameters: {
      smcThreshold: { type: 'number', default: 0.05, description: 'Structure break volatility factor' },
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
    name: 'MA (20)',
    type: 'MA',
    description: 'Moving Average 20 Period',
    params: { maWindow: 20 },
    currentSignal: 'HOLD',
  },
  {
    id: 'ma-50',
    name: 'MA (50)',
    type: 'MA',
    description: 'Moving Average 50 Period',
    params: { maWindow: 50 },
    currentSignal: 'BUY',
  },
  {
    id: 'bb-20-2',
    name: 'Bollinger Bands (20, 2)',
    type: 'Bollinger',
    description: 'BB period 20, multiplier 2.0',
    params: { bbPeriod: 20, bbStdDev: 2 },
    currentSignal: 'SELL',
  },
  {
    id: 'sr-3',
    name: 'Support & Resistance',
    type: 'SupportResistance',
    description: 'Structural pivots (sensitivity: 3)',
    params: { srSensitivity: 3 },
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

// 4. Mock Discovery Loop status
export const DEFAULT_DISCOVERY_STATS: DiscoveryStats = {
  iteration: 47,
  totalIterations: 500,
  testedCandidates: 2350,
  status: 'RUNNING',
  bestStrategy: {
    name: 'MA(20) + RSI(14) + S/R (Weighted)',
    profit: 2342.18,
    winrate: 68.21,
    mdd: 12.45,
  },
};

// 5. Mock Mini Leaderboard top-5
export const MINI_LEADERBOARD_DATA: MiniLeaderboardItem[] = [
  { rank: 1, name: 'MA(20) + RSI(14) + S/R (Weighted)', profit: 2342.18, winrate: 68.21 },
  { rank: 2, name: 'RSI(14) + BB(20,2) (Majority)', profit: 1890.54, winrate: 61.45 },
  { rank: 3, name: 'SMC Breakout + MA(50)', profit: 1420.12, winrate: 59.10 },
  { rank: 4, name: 'MA(20) + MA(50) Crossover', profit: 980.45, winrate: 55.32 },
  { rank: 5, name: 'Support/Resistance Pivot Only', profit: 540.22, winrate: 51.18 },
];
