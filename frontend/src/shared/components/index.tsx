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
            Connecting to realtime market data...
          </>
        ) : (
          <>
            Realtime server unavailable. Dashboard data may be delayed.
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
  backgroundColor: '#ffffff',
  border: '1px solid #fecaca',
  borderRadius: '8px',
  color: '#ffffff',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};

const errorTitleStyle: React.CSSProperties = {
  color: '#dc2626',
  marginBottom: '1rem',
};

const errorMsgStyle: React.CSSProperties = {
  color: '#64748b',
  marginBottom: '1.5rem',
  fontSize: '0.9rem',
};

const errorBtnStyle: React.CSSProperties = {
  backgroundColor: '#dc2626',
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
  backgroundColor: '#fffbeb',
  color: '#92400e',
  padding: '0.55rem',
  textAlign: 'center',
  fontWeight: '600',
  fontSize: '0.9rem',
  borderBottom: '1px solid #fde68a',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};

const bannerDisconnectedStyle: React.CSSProperties = {
  backgroundColor: '#fff7ed',
  color: '#9a3412',
  padding: '0.55rem',
  textAlign: 'center',
  fontWeight: '600',
  fontSize: '0.9rem',
  borderBottom: '1px solid #fed7aa',
  fontFamily: 'system-ui, -apple-system, sans-serif',
};
