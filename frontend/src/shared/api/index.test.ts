import { afterEach, describe, expect, it } from 'vitest';
import { AxiosError, AxiosHeaders, type AxiosAdapter } from 'axios';
import { apiClient, startSearch, startSearchLoop } from '.';
import type { StartSearchLoopRequest } from '../../types/backtest';

const request: StartSearchLoopRequest = {
  pair: 'ETHUSDT',
  timeframe: '1h',
  from: 1,
  to: 2,
  capital: 10_000,
  maxCandidates: 5,
  maxDurationSeconds: 300,
  noImprovementLimit: 3,
};

const originalAdapter = apiClient.defaults.adapter;

afterEach(() => {
  apiClient.defaults.adapter = originalAdapter;
});

describe('startSearchLoop', () => {
  it('posts the complete Discovery contract to /search/loop', async () => {
    const adapter: AxiosAdapter = async config => {
      expect(config.method).toBe('post');
      expect(config.url).toBe('/search/loop');
      expect(JSON.parse(String(config.data))).toEqual(request);
      return {
        data: { searchId: 'search-1', status: 'STARTED', maxCandidates: 5 },
        status: 202,
        statusText: 'Accepted',
        headers: new AxiosHeaders(),
        config,
      };
    };
    apiClient.defaults.adapter = adapter;

    await expect(startSearchLoop(request)).resolves.toEqual({
      searchId: 'search-1',
      status: 'STARTED',
      maxCandidates: 5,
    });
  });

  it('preserves HTTP 422 so the UI can explain that backfill is required', async () => {
    const adapter: AxiosAdapter = async config => {
      throw new AxiosError('Request failed', 'ERR_BAD_REQUEST', config, undefined, {
        data: 'insufficient historical candles; run backfill first',
        status: 422,
        statusText: 'Unprocessable Entity',
        headers: new AxiosHeaders(),
        config,
      });
    };
    apiClient.defaults.adapter = adapter;

    await expect(startSearchLoop(request)).rejects.toMatchObject({
      name: 'ApiError',
      status: 422,
      message: 'insufficient historical candles; run backfill first',
    });
  });
});

describe('startSearch', () => {
  it('normalizes instances and only posts valid StartSearchRequest fields to /search/start', async () => {
    let capturedBody: unknown = null;
    const adapter: AxiosAdapter = async config => {
      expect(config.method).toBe('post');
      expect(config.url).toBe('/search/start');
      capturedBody = JSON.parse(String(config.data));
      return {
        data: { searchId: 'search-123', status: 'STARTED' },
        status: 202,
        statusText: 'Accepted',
        headers: new AxiosHeaders(),
        config,
      };
    };
    apiClient.defaults.adapter = adapter;

    const res = await startSearch({
      pair: 'btcusdt',
      timeframe: '1h',
      from: 1000,
      to: 2000,
      capital: 10000,
      fee: 0.1, // Extra FE fields should NOT be sent in JSON body
      slippage: 5,
      instances: [
        { type: 'MA', params: { period: 20 }, weight: 0.5 },
        { type: 'BBands', params: {}, weight: 0.5 }, // Should be normalized to Bollinger
      ],
      policy: 'weighted',
    });

    expect(res).toEqual({ searchId: 'search-123', status: 'STARTED' });
    expect(capturedBody).toEqual({
      pair: 'BTCUSDT',
      timeframe: '1h',
      from: 1000,
      to: 2000,
      capital: 10000,
      instances: [
        { type: 'MA', params: { period: 20 }, weight: 0.5 },
        { type: 'Bollinger', weight: 0.5 },
      ],
      policy: 'weighted',
    });
    // Ensure no unknown fields exist
    expect(capturedBody).not.toHaveProperty('fee');
    expect(capturedBody).not.toHaveProperty('slippage');
  });
});
