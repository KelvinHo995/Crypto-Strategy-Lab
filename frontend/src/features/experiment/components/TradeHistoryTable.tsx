import { useState } from 'react';
import type { Trade } from '../../../types/backtest';

interface TradeHistoryTableProps {
  trades: Trade[];
  onHoverTrade?: (trade: Trade | null) => void;
  onClickTrade?: (trade: Trade) => void;
}

export function TradeHistoryTable({
  trades,
  onHoverTrade,
  onClickTrade,
}: TradeHistoryTableProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 8;

  const totalPages = Math.ceil(trades.length / itemsPerPage);

  const getPaginatedData = () => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return trades.slice(startIndex, startIndex + itemsPerPage);
  };

  const paginatedTrades = getPaginatedData();

  const getResultBadgeStyle = (profit: number) => {
    const isWin = profit > 0;
    return {
      fontSize: '0.65rem',
      fontWeight: '700',
      color: isWin ? '#10b981' : '#ef4444',
      backgroundColor: isWin ? 'rgba(16, 185, 129, 0.12)' : 'rgba(239, 68, 68, 0.12)',
      border: `1px solid ${isWin ? '#10b981' : '#ef4444'}`,
      borderRadius: '4px',
      padding: '0.15rem 0.4rem',
      textTransform: 'uppercase' as const,
      display: 'inline-block',
    };
  };

  return (
    <div style={containerStyle}>
      <h3 style={titleStyle}>Simulated Execution Log</h3>
      
      {/* Table grid */}
      <div style={tableWrapperStyle}>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={thLeftStyle}>Entry Time</th>
              <th style={thCenterStyle}>Direction</th>
              <th style={thRightStyle}>Volume (USD)</th>
              <th style={thRightStyle}>Entry Price</th>
              <th style={thRightStyle}>Exit Price</th>
              <th style={thRightStyle}>Net Profit</th>
              <th style={thCenterStyle}>Outcome</th>
            </tr>
          </thead>
          <tbody>
            {paginatedTrades.length > 0 ? (
              paginatedTrades.map((trade, idx) => {
                const formattedDate = new Date(trade.entryTime).toLocaleString(undefined, {
                  month: 'short',
                  day: 'numeric',
                  hour: '2-digit',
                  minute: '2-digit',
                });

                const isWin = trade.profit > 0;

                return (
                  <tr
                    key={idx}
                    style={trStyle}
                    onMouseEnter={() => onHoverTrade && onHoverTrade(trade)}
                    onMouseLeave={() => onHoverTrade && onHoverTrade(null)}
                    onClick={() => onClickTrade && onClickTrade(trade)}
                  >
                    <td style={tdLeftStyle}>{formattedDate}</td>
                    <td style={tdCenterStyle}>
                      <span style={trade.direction === 'LONG' ? longBadgeStyle : shortBadgeStyle}>
                        {trade.direction}
                      </span>
                    </td>
                    <td style={tdRightStyle}>${trade.volumeUsd.toLocaleString(undefined, { maximumFractionDigits: 0 })}</td>
                    <td style={tdRightStyle}>${trade.entryPrice.toLocaleString()}</td>
                    <td style={tdRightStyle}>${trade.exitPrice.toLocaleString()}</td>
                    <td style={isWin ? tdProfitStyle : tdLossStyle}>
                      {isWin ? `+$${trade.profit.toFixed(2)}` : `-$${Math.abs(trade.profit).toFixed(2)}`}
                    </td>
                    <td style={tdCenterStyle}>
                      <span style={getResultBadgeStyle(trade.profit)}>
                        {isWin ? 'WIN' : 'LOSS'}
                      </span>
                    </td>
                  </tr>
                );
              })
            ) : (
              <tr>
                <td colSpan={7} style={emptyTdStyle}>
                  No trade execution log available. Run a backtest first.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination controls */}
      {totalPages > 1 && (
        <div style={paginationStyle}>
          <button
            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
            disabled={currentPage === 1}
            style={currentPage === 1 ? disabledPageBtnStyle : pageBtnStyle}
          >
            ◀ Prev
          </button>
          <span style={pageIndicatorStyle}>
            Page {currentPage} of {totalPages}
          </span>
          <button
            onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
            disabled={currentPage === totalPages}
            style={currentPage === totalPages ? disabledPageBtnStyle : pageBtnStyle}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

// ==========================================
// STYLING PRESET
// ==========================================
const containerStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  border: '1px solid #e2e8f0',
  borderRadius: '8px',
  padding: '1.25rem',
  boxSizing: 'border-box',
};

const titleStyle: React.CSSProperties = {
  fontSize: '0.85rem',
  fontWeight: '700',
  color: '#94a3b8',
  margin: '0 0 1rem 0',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
  borderBottom: '1px solid #e2e8f0',
  paddingBottom: '0.5rem',
};

const tableWrapperStyle: React.CSSProperties = {
  overflowX: 'auto',
};

const tableStyle: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  fontSize: '0.8rem',
};

const trStyle: React.CSSProperties = {
  borderBottom: '1px solid #e2e8f0',
  transition: 'background-color 0.15s ease',
  cursor: 'crosshair',
};

const thLeftStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'left',
  padding: '0.65rem 0.5rem',
  fontWeight: '600',
  borderBottom: '1px solid #cbd5e1',
};

const thRightStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'right',
  padding: '0.65rem 0.5rem',
  fontWeight: '600',
  borderBottom: '1px solid #cbd5e1',
};

const thCenterStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'center',
  padding: '0.65rem 0.5rem',
  fontWeight: '600',
  borderBottom: '1px solid #cbd5e1',
};

const tdLeftStyle: React.CSSProperties = {
  color: '#cbd5e1',
  padding: '0.65rem 0.5rem',
};

const tdRightStyle: React.CSSProperties = {
  color: '#cbd5e1',
  textAlign: 'right',
  padding: '0.65rem 0.5rem',
};

const tdCenterStyle: React.CSSProperties = {
  color: '#cbd5e1',
  textAlign: 'center',
  padding: '0.65rem 0.5rem',
};

const tdProfitStyle: React.CSSProperties = {
  color: '#10b981',
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.65rem 0.5rem',
};

const tdLossStyle: React.CSSProperties = {
  color: '#ef4444',
  fontWeight: '700',
  textAlign: 'right',
  padding: '0.65rem 0.5rem',
};

const emptyTdStyle: React.CSSProperties = {
  color: '#64748b',
  textAlign: 'center',
  padding: '2.5rem',
};

const longBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  backgroundColor: 'rgba(16, 185, 129, 0.1)',
  color: '#10b981',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
  fontWeight: '700',
};

const shortBadgeStyle: React.CSSProperties = {
  fontSize: '0.65rem',
  backgroundColor: 'rgba(239, 68, 68, 0.1)',
  color: '#ef4444',
  padding: '0.1rem 0.35rem',
  borderRadius: '4px',
  fontWeight: '700',
};

const paginationStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  gap: '1rem',
  marginTop: '1rem',
};

const pageBtnStyle: React.CSSProperties = {
  backgroundColor: '#e2e8f0',
  color: '#cbd5e1',
  border: '1px solid #cbd5e1',
  borderRadius: '4px',
  padding: '0.3rem 0.65rem',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'pointer',
  outline: 'none',
};

const disabledPageBtnStyle: React.CSSProperties = {
  backgroundColor: 'transparent',
  color: '#475569',
  border: '1px solid #e2e8f0',
  borderRadius: '4px',
  padding: '0.3rem 0.65rem',
  fontSize: '0.75rem',
  fontWeight: '600',
  cursor: 'not-allowed',
  outline: 'none',
};

const pageIndicatorStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  color: '#64748b',
};
