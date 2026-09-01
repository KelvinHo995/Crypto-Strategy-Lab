import type { WSMessage, WSMessageType } from '../../types/websocket';

type MessageListener = (payload: unknown) => void;
type ConnectionState = 'CONNECTED' | 'CONNECTING' | 'DISCONNECTED';
type ConnectionStateListener = (state: ConnectionState) => void;

class WebSocketManager {
  private ws: WebSocket | null = null;
  private url: string;
  private listeners: Map<WSMessageType, Set<MessageListener>> = new Map();
  private stateListeners: Set<ConnectionStateListener> = new Set();
  private connectionState: ConnectionState = 'DISCONNECTED';
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private baseReconnectDelay = 1000; // Start with 1 second delay
  private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
  private intentionallyClosed = false;
  private candleSubscriptions = new Map<string, number>();
  private tradeSubscriptions = new Map<string, number>();

  constructor() {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // During dev, point to backend on port 8080; in production, use standard host
    const wsHost = import.meta.env.VITE_WS_HOST || window.location.host;
    this.url = `${wsProtocol}//${wsHost}/ws`;
  }

  /**
   * Connects to the Go WebSocket server.
   */
  public connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.intentionallyClosed = false;
    this.updateState('CONNECTING');
    console.log(`Connecting to WebSocket at ${this.url}...`);

    try {
      const socket = new WebSocket(this.url);
      this.ws = socket;

      socket.onopen = () => {
        if (this.ws !== socket) {
          socket.close();
          return;
        }
        console.log('WebSocket connection established successfully.');
        this.updateState('CONNECTED');
        this.reconnectAttempts = 0;
        if (this.reconnectTimeoutId) {
          clearTimeout(this.reconnectTimeoutId);
          this.reconnectTimeoutId = null;
        }
        this.candleSubscriptions.forEach((_count, key) => {
          const [symbol, timeframe] = key.split(':');
          this.send('SUBSCRIBE_CANDLES', { symbol, timeframe });
        });
        this.tradeSubscriptions.forEach((_count, symbol) => {
          this.send('SUBSCRIBE_TRADES', { symbol });
        });
      };

      socket.onmessage = (event) => {
        if (this.ws !== socket) return;
        try {
          const message: WSMessage = JSON.parse(event.data);
          this.triggerListeners(message.type, message.payload);
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err, event.data);
        }
      };

      socket.onclose = (event) => {
        if (this.ws !== socket) return;
        this.updateState('DISCONNECTED');
        this.ws = null;
        if (!this.intentionallyClosed) {
          console.warn(`WebSocket closed: ${event.reason || 'No reason'}. Attempting reconnect...`);
          this.attemptReconnect();
        }
      };

      socket.onerror = (error) => {
        if (this.ws !== socket) return;
        console.error('WebSocket encountered an error:', error);
        // connection close will trigger reconnect
      };
    } catch (err) {
      console.error('Error instantiating WebSocket connection:', err);
      this.updateState('DISCONNECTED');
      this.attemptReconnect();
    }
  }

  /**
   * Closes the connection and stops automatic reconnection.
   */
  public disconnect(): void {
    this.intentionallyClosed = true;
    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }
    if (this.ws) {
      const socket = this.ws;
      this.ws = null;
      socket.close();
    }
    this.updateState('DISCONNECTED');
  }

  /**
   * Send a command message to the backend.
   */
  public send(type: string, payload: unknown): boolean {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('Cannot send WebSocket message: Socket is not open.');
      return false;
    }
    try {
      this.ws.send(JSON.stringify({ type, payload }));
      return true;
    } catch (err) {
      console.error('Failed to send WebSocket message:', err);
      return false;
    }
  }

  /**
   * Subscribes to a symbol and timeframe candle updates.
   */
  public subscribeCandles(symbol: string, timeframe: string): void {
    const key = `${symbol.toUpperCase()}:${timeframe}`;
    const count = this.candleSubscriptions.get(key) ?? 0;
    this.candleSubscriptions.set(key, count + 1);
    if (count === 0) this.send('SUBSCRIBE_CANDLES', { symbol, timeframe });
  }

  /**
   * Unsubscribes from a symbol and timeframe candle updates.
   */
  public unsubscribeCandles(symbol: string, timeframe: string): void {
    const key = `${symbol.toUpperCase()}:${timeframe}`;
    const count = this.candleSubscriptions.get(key) ?? 0;
    if (count <= 1) {
      this.candleSubscriptions.delete(key);
      if (count === 1) this.send('UNSUBSCRIBE_CANDLES', { symbol, timeframe });
      return;
    }
    this.candleSubscriptions.set(key, count - 1);
  }

  public subscribeTrades(symbol: string): void {
    const key = symbol.toUpperCase();
    const count = this.tradeSubscriptions.get(key) ?? 0;
    this.tradeSubscriptions.set(key, count + 1);
    if (count === 0) this.send('SUBSCRIBE_TRADES', { symbol });
  }

  public unsubscribeTrades(symbol: string): void {
    const key = symbol.toUpperCase();
    const count = this.tradeSubscriptions.get(key) ?? 0;
    if (count <= 1) {
      this.tradeSubscriptions.delete(key);
      if (count === 1) this.send('UNSUBSCRIBE_TRADES', { symbol });
      return;
    }
    this.tradeSubscriptions.set(key, count - 1);
  }

  /**
   * Register a listener for a specific WS message type.
   */
  public subscribe<T = unknown>(type: WSMessageType, callback: (payload: T) => void): () => void {
    const listener: MessageListener = payload => callback(payload as T);
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(listener);

    // Return an unsubscribe function
    return () => {
      const typeListeners = this.listeners.get(type);
      if (typeListeners) {
        typeListeners.delete(listener);
        if (typeListeners.size === 0) {
          this.listeners.delete(type);
        }
      }
    };
  }

  /**
   * Register a listener for connection state changes (e.g. for connection state indicator).
   */
  public subscribeState(callback: ConnectionStateListener): () => void {
    this.stateListeners.add(callback);
    callback(this.connectionState); // Notify immediately with current state
    return () => {
      this.stateListeners.delete(callback);
    };
  }

  public getConnectionState(): ConnectionState {
    return this.connectionState;
  }

  private updateState(state: ConnectionState): void {
    if (this.connectionState !== state) {
      this.connectionState = state;
      this.stateListeners.forEach((callback) => callback(state));
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max WebSocket reconnect attempts reached. Giving up.');
      return;
    }

    const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms...`);

    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
    }

    this.reconnectTimeoutId = setTimeout(() => {
      this.connect();
    }, delay);
  }

  private triggerListeners(type: WSMessageType, payload: unknown): void {
    const typeListeners = this.listeners.get(type);
    if (typeListeners) {
      typeListeners.forEach((callback) => {
        try {
          callback(payload);
        } catch (err) {
          console.error(`Error in WebSocket listener for type ${type}:`, err);
        }
      });
    }
  }
}

// Singleton instance to be shared across the entire application
export const wsManager = new WebSocketManager();
