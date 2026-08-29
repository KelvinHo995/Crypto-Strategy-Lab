import type { ExperimentResult, Trade } from '../../../types/backtest';

// 1. Mock list of 15 experiments representing different strategy evaluations
export const MOCK_EXPERIMENTS: ExperimentResult[] = [
  {
    id: 'exp-122',
    candidateId: 'cand-001',
    strategies: ['MA', 'RSI', 'SupportResistance'],
    params: { maWindow: 20, rsiPeriod: 14, rsiOverbought: 70, rsiOversold: 30, srSensitivity: 3 },
    policy: 'weighted',
    strategyVersions: { MA: 'v1.2.0', RSI: 'v2.0.1', SupportResistance: 'v1.0.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: 42.18,
    mdd: 12.45,
    tradeCount: 84,
    winRate: 68.21,
    wins: 57,
    losses: 27,
    totalProfit: 4218.00,
    status: 'COMPLETED',
    createdAt: 1723020485000,
  },
  {
    id: 'exp-121',
    candidateId: 'cand-002',
    strategies: ['RSI', 'Bollinger'],
    params: { rsiPeriod: 14, rsiOverbought: 70, rsiOversold: 30, bbPeriod: 20, bbStdDev: 2.0 },
    policy: 'majority',
    strategyVersions: { RSI: 'v2.0.1', Bollinger: 'v1.1.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: 28.54,
    mdd: 9.12,
    tradeCount: 62,
    winRate: 61.45,
    wins: 38,
    losses: 24,
    totalProfit: 2854.00,
    status: 'COMPLETED',
    createdAt: 1723019825000,
  },
  {
    id: 'exp-120',
    candidateId: 'cand-003',
    strategies: ['SMC', 'MA'],
    params: { smcThreshold: 0.05, maWindow: 50 },
    policy: 'weighted',
    strategyVersions: { SMC: 'v1.0.2', MA: 'v1.2.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: 35.12,
    mdd: 15.60,
    tradeCount: 45,
    winRate: 59.10,
    wins: 26,
    losses: 19,
    totalProfit: 3512.00,
    status: 'COMPLETED',
    createdAt: 1723011400000,
  },
  {
    id: 'exp-119',
    candidateId: 'cand-004',
    strategies: ['MA', 'MA'],
    params: { maFastWindow: 20, maSlowWindow: 50 },
    policy: 'majority',
    strategyVersions: { MA: 'v1.2.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: 18.45,
    mdd: 8.32,
    tradeCount: 110,
    winRate: 55.32,
    wins: 60,
    losses: 50,
    totalProfit: 1845.00,
    status: 'COMPLETED',
    createdAt: 1723000850000,
  },
  {
    id: 'exp-118',
    candidateId: 'cand-005',
    strategies: ['SupportResistance'],
    params: { srSensitivity: 3 },
    policy: 'majority',
    strategyVersions: { SupportResistance: 'v1.0.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: 5.40,
    mdd: 14.18,
    tradeCount: 38,
    winRate: 51.18,
    wins: 19,
    losses: 19,
    totalProfit: 540.00,
    status: 'COMPLETED',
    createdAt: 1722989500000,
  },
  {
    id: 'exp-117',
    candidateId: 'cand-006',
    strategies: ['SMC', 'RSI', 'Bollinger'],
    params: { smcThreshold: 0.08, rsiPeriod: 10, bbPeriod: 14, bbStdDev: 1.8 },
    policy: 'weighted',
    strategyVersions: { SMC: 'v1.0.2', RSI: 'v2.0.1', Bollinger: 'v1.1.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: -4.12,
    mdd: 22.45,
    tradeCount: 76,
    winRate: 46.05,
    wins: 35,
    losses: 41,
    totalProfit: -412.00,
    status: 'COMPLETED',
    createdAt: 1722982100000,
  },
  {
    id: 'exp-116',
    candidateId: 'cand-007',
    strategies: ['RSI'],
    params: { rsiPeriod: 14, rsiOverbought: 80, rsiOversold: 20 },
    policy: 'majority',
    strategyVersions: { RSI: 'v2.0.1' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: 12.30,
    mdd: 6.15,
    tradeCount: 22,
    winRate: 58.18,
    wins: 13,
    losses: 9,
    totalProfit: 1230.00,
    status: 'COMPLETED',
    createdAt: 1722971400000,
  },
  {
    id: 'exp-115',
    candidateId: 'cand-008',
    strategies: ['Bollinger'],
    params: { bbPeriod: 20, bbStdDev: 2.5 },
    policy: 'majority',
    strategyVersions: { Bollinger: 'v1.1.0' },
    datasetPeriod: '2024-01-01 to 2024-12-31',
    return: -11.45,
    mdd: 18.90,
    tradeCount: 28,
    winRate: 42.86,
    wins: 12,
    losses: 16,
    totalProfit: -1145.00,
    status: 'FAILED',
    createdAt: 1722960800000,
  }
];

// 2. Generates mock trade details representing execution results of an experiment
export function generateMockTrades(experimentId: string, count: number = 30): Trade[] {
  const trades: Trade[] = [];
  let currentCapital = 10000;
  
  // Deterministic seed based on experimentId characters
  let seed = 0;
  for (let i = 0; i < experimentId.length; i++) {
    seed += experimentId.charCodeAt(i);
  }
  
  const randomWithSeed = () => {
    const x = Math.sin(seed++) * 10000;
    return x - Math.floor(x);
  };

  let time = Date.now() - count * 12 * 60 * 60 * 1000; // 12 hours interval base
  let price = 52000;

  for (let i = 0; i < count; i++) {
    const direction = randomWithSeed() > 0.45 ? 'LONG' : 'SHORT';
    const entryPrice = price + (randomWithSeed() * 200 - 100);
    const stopLoss = direction === 'LONG' ? entryPrice * 0.98 : entryPrice * 1.02;
    const takeProfit = direction === 'LONG' ? entryPrice * 1.04 : entryPrice * 0.96;
    
    // Simulate trade outcome
    const isWin = randomWithSeed() > 0.38; // ~62% winrate mock
    const exitPrice = isWin ? takeProfit : stopLoss;
    
    const profitPct = isWin ? 0.04 : -0.02;
    const positionSizeUSD = currentCapital * 0.1; // 10% position size
    const profit = positionSizeUSD * profitPct * (direction === 'LONG' ? 1 : -1);
    
    currentCapital += profit;

    const transactionCost = positionSizeUSD * 0.001; // 0.1% fee
    const slippage = positionSizeUSD * 0.0005; // 0.05% slippage

    trades.push({
      pair: 'BTCUSDT',
      entryTime: time,
      direction,
      volumeUSD: positionSizeUSD,
      entryPrice: Number(entryPrice.toFixed(2)),
      stopLoss: Number(stopLoss.toFixed(2)),
      takeProfit: Number(takeProfit.toFixed(2)),
      exitPrice: Number(exitPrice.toFixed(2)),
      exitTime: time + Math.floor(randomWithSeed() * 8 * 60 * 60 * 1000), // exit after 1-8 hours
      transactionCost: Number(transactionCost.toFixed(2)),
      slippage: Number(slippage.toFixed(2)),
      profit: Number((profit - transactionCost - slippage).toFixed(2)),
    });

    price = exitPrice;
    time += (12 + randomWithSeed() * 24) * 60 * 60 * 1000; // Next trade after 12-36 hours
  }

  return trades.reverse(); // Newest trades first
}
