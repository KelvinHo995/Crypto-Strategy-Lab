// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, waitFor } from '@testing-library/react';
import { AppModeContext } from '../../../shared/auth';
import type { ExperimentResult } from '../../../types/backtest';
import type { Candle } from '../../../types/candle';
import { BacktestChart } from './BacktestChart';

const api = vi.hoisted(() => ({ fetchCandles: vi.fn() }));

vi.mock('../../../shared/api', async importOriginal => ({
  ...(await importOriginal<typeof import('../../../shared/api')>()),
  fetchCandles: api.fetchCandles,
}));

vi.mock('../../market/components/TradingChart', () => ({
  TradingChart: () => <div data-testid="trading-chart" />,
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('BacktestChart market provenance', () => {
  it('loads candles using the experiment pair and actual timeframe', async () => {
    const candles: Candle[] = Array.from({ length: 20 }, (_, index) => ({
      symbol: 'BTCUSDT', timeframe: '1h', openTime: index + 1,
      open: 100, high: 101, low: 99, close: 100, volume: 1,
    }));
    api.fetchCandles.mockResolvedValue(candles);
    const experiment: ExperimentResult = {
      id: 'experiment-1', candidateId: 'candidate-1', pair: 'BTCUSDT', timeframe: '1h',
      instances: [{ type: 'MA' }], policy: 'majority', strategyVersions: { MA: 'v1' },
      datasetPeriod: '1-20', return: 1, mdd: 0, tradeCount: 0, winRate: 0,
      wins: 0, losses: 0, totalProfit: 0, status: 'COMPLETED', createdAt: 1,
    };

    render(
      <AppModeContext.Provider value="LIVE">
        <BacktestChart experiment={experiment} trades={[]} highlightedTrade={null} />
      </AppModeContext.Provider>,
    );

    await waitFor(() => expect(api.fetchCandles).toHaveBeenCalledWith('BTCUSDT', '1h', 1, 20, 5000));
    expect(api.fetchCandles).not.toHaveBeenCalledWith('BTCUSDT', '4h', 1, 20, 5000);
  });
});
