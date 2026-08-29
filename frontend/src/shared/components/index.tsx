import React, { Component } from 'react';
import type { ErrorInfo, ReactNode } from 'react';
import { useWebSocketState } from '../hooks';

// ==========================================
// 1. ERROR BOUNDARY COMPONENT
// ==========================================
interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  public state: ErrorBoundaryState = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an unhandled error:', error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      if (this.fallbackCustom()) {
        return this.fallbackCustom();
      }
      return (
        <div style={errorContainerStyle}>
          <h2 style={errorTitleStyle}>Something went wrong</h2>
          <p style={errorMsgStyle}>{this.state.error?.message || 'An unexpected client error occurred.'}</p>
          <button style={errorBtnStyle} onClick={() => window.location.reload()}>
            Reload Application
          </button>
        </div>
      );
    }

    return this.props.children;
  }

  private fallbackCustom() {
    return this.props.fallback || null;
  }
}

// ==========================================
// 2. WEBSOCKET STATE BANNER COMPONENT
// ==========================================
export function WebSocketStateBanner() {
  const wsState = useWebSocketState();

  if (wsState === 'CONNECTED') {
    return null;
  }

  const isConnecting = wsState === 'CONNECTING';

  return (
    <div style={isConnecting ? bannerConnectingStyle : bannerDisconnectedStyle}>
      <span style={bannerTextStyle}>
        {isConnecting ? (
          <>
            ⚠️ Connection lost. Attempting to reconnect...
          </>
        ) : (
          <>
            🚨 Disconnected from Strategy Lab server. Checking server status...
          </>
        )}
      </span>
    </div>
  );
}

// ==========================================
// INLINE STYLES FOR SIMPLICITY
// ==========================================
const errorContainerStyle: React.CSSProperties = {
  padding: '2rem',
  margin: '2rem auto',
  maxWidth: '500px',
  textAlign: 'center',
  backgroundColor: '#1a1a1a',
  border: '1px solid #ff4a4a',
  borderRadius: '8px',
  color: '#ffffff',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};

const errorTitleStyle: React.CSSProperties = {
  color: '#ff4a4a',
  marginBottom: '1rem',
};

const errorMsgStyle: React.CSSProperties = {
  color: '#cccccc',
  marginBottom: '1.5rem',
  fontSize: '0.9rem',
};

const errorBtnStyle: React.CSSProperties = {
  backgroundColor: '#ff4a4a',
  color: 'white',
  border: 'none',
  padding: '0.5rem 1rem',
  borderRadius: '4px',
  cursor: 'pointer',
  fontWeight: 'bold',
};

const bannerTextStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  gap: '0.5rem',
};

const bannerConnectingStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0,
  left: 0,
  right: 0,
  backgroundColor: '#f59e0b', // Yellow-500
  color: '#000000',
  padding: '0.5rem',
  textAlign: 'center',
  fontWeight: '600',
  fontSize: '0.9rem',
  zIndex: 9999,
  boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};

const bannerDisconnectedStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0,
  left: 0,
  right: 0,
  backgroundColor: '#ef4444', // Red-500
  color: '#ffffff',
  padding: '0.5rem',
  textAlign: 'center',
  fontWeight: '600',
  fontSize: '0.9rem',
  zIndex: 9999,
  boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};
