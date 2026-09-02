// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

class FakeWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 3;
  static instances: FakeWebSocket[] = [];

  readonly url: string;
  readyState = FakeWebSocket.CONNECTING;
  sent: string[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: ((event: { reason: string }) => void) | null = null;
  onerror: ((error: unknown) => void) | null = null;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  send(data: string) {
    this.sent.push(data);
  }

  close() {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.({ reason: 'client closed' });
  }

  open() {
    this.readyState = FakeWebSocket.OPEN;
    this.onopen?.();
  }

  receive(message: unknown) {
    this.onmessage?.({ data: JSON.stringify(message) });
  }

  disconnectUnexpectedly() {
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.({ reason: 'network lost' });
  }
}

let wsManager: typeof import('.')['wsManager'];

beforeEach(async () => {
  vi.resetModules();
  vi.useRealTimers();
  FakeWebSocket.instances = [];
  vi.stubGlobal('WebSocket', FakeWebSocket);
  vi.spyOn(console, 'log').mockImplementation(() => undefined);
  vi.spyOn(console, 'warn').mockImplementation(() => undefined);
  vi.spyOn(console, 'error').mockImplementation(() => undefined);
  ({ wsManager } = await import('.'));
});

afterEach(() => {
  wsManager.disconnect();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe('WebSocketManager', () => {
  it('connects once and publishes connection state transitions', () => {
    const states: string[] = [];
    wsManager.subscribeState(state => states.push(state));

    wsManager.connect();
    wsManager.connect();
    expect(FakeWebSocket.instances).toHaveLength(1);
    expect(FakeWebSocket.instances[0].url).toBe('ws://localhost:3000/ws');

    FakeWebSocket.instances[0].open();
    expect(states).toEqual(['DISCONNECTED', 'CONNECTING', 'CONNECTED']);

    wsManager.disconnect();
    expect(states.at(-1)).toBe('DISCONNECTED');
  });

  it('deduplicates subscriptions and restores them after reconnect', async () => {
    vi.useFakeTimers();
    wsManager.subscribeCandles('ETHUSDT', '1h');
    wsManager.subscribeCandles('ETHUSDT', '1h');
    wsManager.connect();
    FakeWebSocket.instances[0].open();

    expect(FakeWebSocket.instances[0].sent).toEqual([
      JSON.stringify({ type: 'SUBSCRIBE_CANDLES', payload: { symbol: 'ETHUSDT', timeframe: '1h' } }),
    ]);

    FakeWebSocket.instances[0].disconnectUnexpectedly();
    await vi.advanceTimersByTimeAsync(1000);
    expect(FakeWebSocket.instances).toHaveLength(2);
    FakeWebSocket.instances[1].open();
    expect(FakeWebSocket.instances[1].sent).toEqual([
      JSON.stringify({ type: 'SUBSCRIBE_CANDLES', payload: { symbol: 'ETHUSDT', timeframe: '1h' } }),
    ]);

    wsManager.unsubscribeCandles('ETHUSDT', '1h');
    expect(FakeWebSocket.instances[1].sent).toHaveLength(1);
    wsManager.unsubscribeCandles('ETHUSDT', '1h');
    expect(FakeWebSocket.instances[1].sent.at(-1)).toBe(
      JSON.stringify({ type: 'UNSUBSCRIBE_CANDLES', payload: { symbol: 'ETHUSDT', timeframe: '1h' } }),
    );
  });

  it('dispatches only matching message types and supports unsubscribe', () => {
    const progress = vi.fn();
    const candles = vi.fn();
    const unsubscribe = wsManager.subscribe('SEARCH_PROGRESS', progress);
    wsManager.subscribe('CANDLE_UPDATE', candles);
    wsManager.connect();
    FakeWebSocket.instances[0].open();

    FakeWebSocket.instances[0].receive({ type: 'SEARCH_PROGRESS', payload: { tested: 2, total: 5 } });
    expect(progress).toHaveBeenCalledWith({ tested: 2, total: 5 });
    expect(candles).not.toHaveBeenCalled();

    unsubscribe();
    FakeWebSocket.instances[0].receive({ type: 'SEARCH_PROGRESS', payload: { tested: 3, total: 5 } });
    expect(progress).toHaveBeenCalledOnce();
  });
});
