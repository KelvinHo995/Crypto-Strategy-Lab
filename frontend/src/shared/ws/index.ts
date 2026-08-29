import type { WSMessage, WSMessageType } from '../../types/websocket';

type MessageListener<T = any> = (payload: T) => void;
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
  private reconnectTimeoutId: any = null;
  private intentionallyClosed = false;

  constructor() {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // During dev, point to backend on port 8080; in production, use standard host
    const wsHost = import.meta.env.VITE_WS_HOST || 'localhost:8080';
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
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connection established successfully.');
        this.updateState('CONNECTED');
        this.reconnectAttempts = 0;
        if (this.reconnectTimeoutId) {
          clearTimeout(this.reconnectTimeoutId);
          this.reconnectTimeoutId = null;
        }
        // Resend active subscriptions or initialization state if needed
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          this.triggerListeners(message.type, message.payload);
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err, event.data);
        }
      };

      this.ws.onclose = (event) => {
        this.updateState('DISCONNECTED');
        this.ws = null;
        if (!this.intentionallyClosed) {
          console.warn(`WebSocket closed: ${event.reason || 'No reason'}. Attempting reconnect...`);
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
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
      this.ws.close();
      this.ws = null;
    }
    this.updateState('DISCONNECTED');
  }

  /**
   * Send a command message to the backend.
   */
  public send(type: string, payload: any): boolean {
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
    this.send('SUBSCRIBE_CANDLES', { symbol, timeframe });
  }

  /**
   * Unsubscribes from a symbol and timeframe candle updates.
   */
  public unsubscribeCandles(symbol: string, timeframe: string): void {
    this.send('UNSUBSCRIBE_CANDLES', { symbol, timeframe });
  }

  /**
   * Register a listener for a specific WS message type.
   */
  public subscribe<T = any>(type: WSMessageType, callback: MessageListener<T>): () => void {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(callback);

    // Return an unsubscribe function
    return () => {
      const typeListeners = this.listeners.get(type);
      if (typeListeners) {
        typeListeners.delete(callback);
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

  private triggerListeners(type: WSMessageType, payload: any): void {
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
