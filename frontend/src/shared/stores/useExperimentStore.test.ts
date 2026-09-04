import { describe, it, expect, beforeEach } from 'vitest';
import { useExperimentStore, tradesToMarkers } from './useExperimentStore';
import type { ExperimentResult, Trade } from '../../types/backtest';

describe('useExperimentStore', () => {
  beforeEach(() => {
    useExperimentStore.getState().clearExperiment();
    useExperimentStore.setState({
      activeTab: 'charts',
      builderInstances: [],
      builderPolicy: 'weighted',
    });
  });

  it('transforms trades into TradingView chart markers accurately', () => {
    const mockTrades: Trade[] = [
      {
        pair: 'BTCUSDT',
        entryTime: 1700000000000,
        direction: 'LONG',
        volumeUSD: 1000,
        entryPrice: 40000,
        stopLoss: 39000,
        takeProfit: 42000,
        exitPrice: 41000,
        exitTime: 1700003600000,
        transactionCost: 1,
        slippage: 0,
        profit: 25,
      },
      {
        pair: 'BTCUSDT',
        entryTime: 1700007200000,
        direction: 'SHORT',
        volumeUSD: 1000,
        entryPrice: 41000,
        stopLoss: 42000,
        takeProfit: 39000,
        exitPrice: 41500,
        exitTime: 1700010800000,
        transactionCost: 1,
        slippage: 0,
        profit: -12,
      },
    ];

    const markers = tradesToMarkers(mockTrades);
    expect(markers.length).toBe(4); // 2 entries + 2 exits

    // First trade entry marker
    expect(markers[0]).toEqual({
      time: 1700000000000,
      position: 'belowBar',
      color: '#10b981',
      shape: 'arrowUp',
      text: 'BUY',
    });

    // First trade exit marker
    expect(markers[1]).toEqual({
      time: 1700003600000,
      position: 'aboveBar',
      color: '#10b981',
      shape: 'circle',
      text: 'WIN +$25',
    });

    // Second trade entry marker
    expect(markers[2]).toEqual({
      time: 1700007200000,
      position: 'aboveBar',
      color: '#ef4444',
      shape: 'arrowDown',
      text: 'SELL',
    });
  });

  it('loads experiment to chart and switches activeTab to charts', () => {
    const mockExp: ExperimentResult = {
      id: 'exp-001',
      candidateId: 'cand-001',
      instances: [{ type: 'MA', params: { maShortWindow: 20, maLongWindow: 50 }, weight: 1 }],
      policy: 'majority',
      strategyVersions: { MA: 'v1.0' },
      datasetPeriod: '180d',
      return: 15.5,
      mdd: 6.2,
      tradeCount: 10,
      winRate: 60,
      wins: 6,
      losses: 4,
      totalProfit: 1550,
      status: 'COMPLETED',
      createdAt: Date.now(),
    };

    useExperimentStore.getState().loadToChart(mockExp);

    const state = useExperimentStore.getState();
    expect(state.activeExperiment?.id).toBe('exp-001');
    expect(state.activeTab).toBe('charts');
    expect(state.activeMarkers.length).toBeGreaterThan(0);
  });

  it('replicates experiment instances to builder and switches activeTab to builder', () => {
    const instances = [
      { type: 'RSI', params: { rsiPeriod: 14 }, weight: 0.6 },
      { type: 'Bollinger', params: { bollingerPeriod: 20 }, weight: 0.4 },
    ];

    useExperimentStore.getState().replicateToBuilder(instances, 'weighted');

    const state = useExperimentStore.getState();
    expect(state.builderInstances).toEqual(instances);
    expect(state.builderPolicy).toBe('weighted');
    expect(state.activeTab).toBe('builder');
  });

  it('injects Sentiment strategy to builder smoothly', () => {
    useExperimentStore.getState().injectStrategyToBuilder({
      type: 'Sentiment',
      params: { sentimentThreshold: 0.75 },
      weight: 0.5,
    });

    const state = useExperimentStore.getState();
    expect(state.builderInstances.some((i) => i.type === 'Sentiment')).toBe(true);
    expect(state.activeTab).toBe('builder');
  });
});
