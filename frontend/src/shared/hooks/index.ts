import { useState, useEffect } from 'react';
import { wsManager } from '../ws';
import type { WSMessageType } from '../../types/websocket';

/**
 * Custom hook to get and monitor the current WebSocket connection state.
 */
export function useWebSocketState() {
  const [state, setState] = useState(wsManager.getConnectionState());

  useEffect(() => {
    return wsManager.subscribeState((newState) => {
      setState(newState);
    });
  }, []);

  return state;
}

/**
 * Custom hook to subscribe to a specific WebSocket message type.
 * Automatically unsubscribes when the component unmounts or parameters change.
 */
export function useWebSocketSubscription<T = any>(
  type: WSMessageType,
  onMessage: (payload: T) => void
) {
  useEffect(() => {
    return wsManager.subscribe<T>(type, onMessage);
  }, [type, onMessage]);
}
