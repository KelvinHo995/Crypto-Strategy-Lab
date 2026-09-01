import { useWebSocketState } from '../../../shared/hooks';
import { useAppMode } from '../../../shared/auth';

export function MarketStatusBar() {
  const wsState = useWebSocketState();
  const mode = useAppMode();

  const getStatusColor = () => {
    switch (wsState) {
      case 'CONNECTED': return '#10b981'; // Green
      case 'CONNECTING': return '#f59e0b'; // Yellow
      case 'DISCONNECTED': return '#ef4444'; // Red
    }
  };

  const getStatusText = () => {
    switch (wsState) {
      case 'CONNECTED': return 'WS Connected';
      case 'CONNECTING': return 'Reconnecting...';
      case 'DISCONNECTED': return 'Offline';
    }
  };

  return (
    <div style={statusBarContainerStyle}>
      {/* Network Status Dot */}
      <div style={itemStyle}>
        <span style={labelStyle}>Connection:</span>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
          <span style={{ ...statusDotStyle, backgroundColor: getStatusColor(), boxShadow: `0 0 8px ${getStatusColor()}` }} />
          <span style={{ fontSize: '0.75rem', fontWeight: '700', color: getStatusColor() }}>
            {getStatusText()}
          </span>
        </div>
      </div>

      {/* Latency display */}
      <div style={itemStyle}>
        <span style={labelStyle}>Latency:</span>
        <span style={valueStyle}>{wsState === 'CONNECTED' ? 'Streaming' : '--'}</span>
      </div>

      {/* Provider Details */}
      <div style={itemStyle}>
        <span style={labelStyle}>Feed Provider:</span>
        <span style={{ ...valueStyle, color: '#3b82f6' }}>{mode === 'LIVE' ? 'Binance API' : 'Offline fixture'}</span>
      </div>

      {/* Adapter Interface */}
      <div style={itemStyle}>
        <span style={labelStyle}>Channel:</span>
        <span style={valueStyle}>Go Monolith WebSocket</span>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const statusBarContainerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.85rem',
  padding: '1rem',
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
};

const itemStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  fontSize: '0.75rem',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
};

const labelStyle: React.CSSProperties = {
  color: '#64748b',
  fontWeight: '500',
};

const valueStyle: React.CSSProperties = {
  color: '#0f172a',
  fontWeight: '600',
};

const statusDotStyle: React.CSSProperties = {
  width: '7px',
  height: '7px',
  borderRadius: '50%',
  display: 'inline-block',
};
