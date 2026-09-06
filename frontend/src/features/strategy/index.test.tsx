// @vitest-environment jsdom

import { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { AppModeContext } from '../../shared/auth';
import { ApiError } from '../../shared/api';
import type { ExperimentResult } from '../../types/backtest';
import { StrategyDiscoveryPage } from '.';

const api = vi.hoisted(() => ({
  fetchMarkets: vi.fn(),
  fetchStrategies: vi.fn(),
  startSearch: vi.fn(),
  startSearchLoop: vi.fn(),
}));

const subscriptions = vi.hoisted(() => new Map<string, (payload: unknown) => void>());

vi.mock('../../shared/api', async importOriginal => ({
  ...(await importOriginal<typeof import('../../shared/api')>()),
  ...api,
}));

vi.mock('../../shared/hooks', () => ({
  useWebSocketSubscription: (type: string, callback: (payload: unknown) => void) => {
    subscriptions.set(type, callback);
  },
}));

function renderLivePage() {
  return render(
    <AppModeContext.Provider value="LIVE">
      <StrategyDiscoveryPage />
    </AppModeContext.Provider>,
  );
}

beforeEach(() => {
  subscriptions.clear();
  api.fetchStrategies.mockResolvedValue(['MA', 'RSI', 'Bollinger', 'SR', 'SMC']);
  api.fetchMarkets.mockResolvedValue([]);
  api.startSearch.mockResolvedValue({ searchId: 'single-1', status: 'STARTED' });
  api.startSearchLoop.mockReset();
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('StrategyDiscoveryPage Search Loop integration', () => {
  it('sends one request and disables duplicate submission while it is pending', async () => {
    let resolveRequest: ((value: { searchId: string; status: 'STARTED'; maxCandidates: number }) => void) | undefined;
    api.startSearchLoop.mockImplementation(() => new Promise(resolve => {
      resolveRequest = resolve;
    }));
    renderLivePage();

    const start = screen.getByRole('button', { name: 'Start Loop Discovery' });
    fireEvent.click(start);
    fireEvent.click(start);

    expect(api.startSearchLoop).toHaveBeenCalledOnce();
    expect(api.startSearchLoop).toHaveBeenCalledWith(expect.objectContaining({
      pair: 'BTCUSDT',
      timeframe: '1h',
      capital: 10_000,
      maxCandidates: 10,
      maxDurationSeconds: 300,
      noImprovementLimit: 5,
    }));
    expect((screen.getByRole('button', { name: 'Discovery đang chạy' }) as HTMLButtonElement).disabled).toBe(true);

    await act(async () => resolveRequest?.({ searchId: 'search-1', status: 'STARTED', maxCandidates: 10 }));
    expect(await screen.findByText('Search ID: search-1')).toBeTruthy();
  });

  it('shows an actionable backfill message for HTTP 422', async () => {
    api.startSearchLoop.mockRejectedValue(new ApiError('insufficient historical candles', 422));
    renderLivePage();

    fireEvent.click(screen.getByRole('button', { name: 'Start Loop Discovery' }));

    expect(await screen.findByText(/Không đủ candle/)).toBeTruthy();
    expect(screen.getByText('FAILED')).toBeTruthy();
  });

  it('renders progress and leaderboard only from WebSocket payloads', async () => {
    api.startSearchLoop.mockResolvedValue({ searchId: 'search-1', status: 'STARTED', maxCandidates: 10 });
    renderLivePage();
    fireEvent.click(screen.getByRole('button', { name: 'Start Loop Discovery' }));
    await screen.findByText('Search ID: search-1');

    await act(async () => subscriptions.get('SEARCH_PROGRESS')?.({
      searchId: 'search-1',
      tested: 3,
      total: 10,
      status: 'RUNNING',
    }));
    expect(screen.getByText('3 / 10')).toBeTruthy();

    const best: ExperimentResult = {
      id: 'result-1',
      searchId: 'search-1',
      searchTotal: 10,
      candidateId: 'candidate-1',
      instances: [{ type: 'MA' }, { type: 'RSI' }],
      policy: 'majority',
      strategyVersions: { MA: 'v1', RSI: 'v1' },
      datasetPeriod: '1-2',
      return: 4.2,
      mdd: 1.1,
      tradeCount: 8,
      winRate: 62.5,
      wins: 5,
      losses: 3,
      totalProfit: 420,
      status: 'COMPLETED',
      createdAt: 1,
    };
    await act(async () => subscriptions.get('LEADERBOARD_UPDATE')?.([best]));

    await waitFor(() => expect(screen.getAllByText('MA + RSI').length).toBeGreaterThanOrEqual(2));
    expect(screen.getAllByText('+420.00 USDT')).toHaveLength(2);
    expect(screen.queryByText(/MA\(20\) \+ RSI\(14\)/)).toBeNull();
  });
});
