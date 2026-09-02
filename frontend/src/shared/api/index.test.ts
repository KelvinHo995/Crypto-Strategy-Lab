import { afterEach, describe, expect, it } from 'vitest';
import { AxiosError, AxiosHeaders, type AxiosAdapter } from 'axios';
import { apiClient, startSearchLoop } from '.';
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
