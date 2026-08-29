import { useState, useEffect } from 'react';

interface TradeTick {
  id: string;
  time: string;
  price: number;
  size: number;
  side: 'BUY' | 'SELL';
}

function getFormattedTime(date: Date): string {
  return date.toTimeString().split(' ')[0];
}

export function RecentTicksPanel() {
  const [ticks, setTicks] = useState<TradeTick[]>(() => {
    const initialTicks: TradeTick[] = [];
    let price = 50250;
    
    for (let i = 0; i < 10; i++) {
      price += (Math.random() * 20 - 10);
      initialTicks.unshift({
        id: `t-${Date.now()}-${i}`,
        time: getFormattedTime(new Date(Date.now() - i * 1500)),
        price: Number(price.toFixed(2)),
        size: Number((Math.random() * 2 + 0.01).toFixed(4)),
        side: Math.random() > 0.5 ? 'BUY' : 'SELL',
      });
    }
    return initialTicks;
  });

  // Simulates a live stream of trades from Binance WS
  useEffect(() => {
    const interval = setInterval(() => {
      setTicks((prev) => {
        const lastTick = prev[0];
        if (!lastTick) return prev;

        const priceChange = (Math.random() * 12 - 6);
        const newPrice = Number((lastTick.price + priceChange).toFixed(2));
        const newTick: TradeTick = {
          id: `t-${Date.now()}`,
          time: getFormattedTime(new Date()),
          price: newPrice,
          size: Number((Math.random() * 1.5 + 0.005).toFixed(4)),
          side: Math.random() > 0.48 ? 'BUY' : 'SELL',
        };

        // Keep last 12 ticks
        return [newTick, ...prev.slice(0, 11)];
      });
    }, 1200);

    return () => clearInterval(interval);
  }, []);

  return (
    <div style={panelContainerStyle}>
      <h3 style={sectionTitleStyle}>Live Feed Ticks</h3>
      
      {/* Ticks Table */}
      <div style={tableWrapperStyle}>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={thLeftStyle}>Time</th>
              <th style={thRightStyle}>Price</th>
              <th style={thRightStyle}>Size</th>
            </tr>
          </thead>
          <tbody>
            {ticks.map((tick) => (
              <tr key={tick.id} style={trStyle}>
                <td style={tdLeftStyle}>{tick.time}</td>
                <td style={tick.side === 'BUY' ? tdBuyStyle : tdSellStyle}>
                  {tick.price.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                  {tick.side === 'BUY' ? ' ↗' : ' ↘'}
                </td>
                <td style={tdRightStyle}>{tick.size}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Chart Legend */}
      <div style={legendSectionStyle}>
        <h3 style={sectionTitleStyle}>Indicators Legend</h3>
        <div style={legendGridStyle}>
          <div style={legendItemStyle}>
            <span style={greenCandleIconStyle} />
            <span style={legendLabelStyle}>Bullish Candlestick</span>
          </div>
          <div style={legendItemStyle}>
            <span style={redCandleIconStyle} />
            <span style={legendLabelStyle}>Bearish Candlestick</span>
          </div>
          <div style={legendItemStyle}>
            <span style={maLineIconStyle} />
            <span style={legendLabelStyle}>MA (20) Line</span>
          </div>
          <div style={legendItemStyle}>
            <span style={bbBandIconStyle} />
            <span style={legendLabelStyle}>Bollinger Bands (20, 2)</span>
          </div>
          <div style={legendItemStyle}>
            <span style={buyMarkerIconStyle}>▲</span>
            <span style={legendLabelStyle}>Strategy BUY Signal</span>
          </div>
          <div style={legendItemStyle}>
            <span style={sellMarkerIconStyle}>▼</span>
            <span style={legendLabelStyle}>Strategy SELL Signal</span>
          </div>
        </div>
      </div>
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const panelContainerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '1rem',
  backgroundColor: '#0f172a',
  border: '1px solid #1e293b',
  borderRadius: '8px',
  padding: '1rem',
  height: '100%',
  boxSizing: 'border-box',
};

const sectionTitleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#94a3b8',
  margin: '0 0 0.5rem 0',
  borderBottom: '1px solid #1e293b',
  paddingBottom: '0.25rem',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
};

const tableWrapperStyle: React.CSSProperties = {
  flexGrow: 1,
  overflowY: 'auto',
  maxHeight: '360px',
};

const tableStyle: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  fontSize: '0.75rem',
};

const trStyle: React.CSSProperties = {
  borderBottom: '1px solid #0f172a',
};

const thLeftStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'left',
  padding: '0.25rem 0',
  fontWeight: '600',
};

const thRightStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'right',
  padding: '0.25rem 0',
  fontWeight: '600',
};

const tdLeftStyle: React.CSSProperties = {
  color: '#94a3b8',
  padding: '0.35rem 0',
};

const tdRightStyle: React.CSSProperties = {
  color: '#cbd5e1',
  textAlign: 'right',
  padding: '0.35rem 0',
};

const tdBuyStyle: React.CSSProperties = {
  color: '#10b981',
  textAlign: 'right',
  fontWeight: '600',
  padding: '0.35rem 0',
};

const tdSellStyle: React.CSSProperties = {
  color: '#ef4444',
  textAlign: 'right',
  fontWeight: '600',
  padding: '0.35rem 0',
};

const legendSectionStyle: React.CSSProperties = {
  marginTop: '0.5rem',
};

const legendGridStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: '0.5rem',
};

const legendItemStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.5rem',
  fontSize: '0.75rem',
};

const legendLabelStyle: React.CSSProperties = {
  color: '#94a3b8',
};

const greenCandleIconStyle: React.CSSProperties = {
  width: '8px',
  height: '14px',
  backgroundColor: '#10b981',
  borderRadius: '1px',
};

const redCandleIconStyle: React.CSSProperties = {
  width: '8px',
  height: '14px',
  backgroundColor: '#ef4444',
  borderRadius: '1px',
};

const maLineIconStyle: React.CSSProperties = {
  width: '16px',
  height: '2px',
  backgroundColor: '#3b82f6',
};

const bbBandIconStyle: React.CSSProperties = {
  width: '16px',
  height: '2px',
  backgroundColor: '#a855f7',
  borderStyle: 'dashed',
};

const buyMarkerIconStyle: React.CSSProperties = {
  color: '#10b981',
  fontWeight: 'bold',
  fontSize: '0.75rem',
};

const sellMarkerIconStyle: React.CSSProperties = {
  color: '#ef4444',
  fontWeight: 'bold',
  fontSize: '0.75rem',
};
