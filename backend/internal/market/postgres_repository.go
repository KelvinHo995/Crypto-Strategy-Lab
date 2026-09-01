package market

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const candleUpsertBatchSize = 500

type PostgresCandleRepository struct{ db *sql.DB }

func NewPostgresCandleRepository(db *sql.DB) *PostgresCandleRepository {
	return &PostgresCandleRepository{db: db}
}
func (r *PostgresCandleRepository) Upsert(ctx context.Context, candles []Candle) error {
	if len(candles) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for start := 0; start < len(candles); start += candleUpsertBatchSize {
		end := start + candleUpsertBatchSize
		if end > len(candles) {
			end = len(candles)
		}
		query, args := candleUpsertStatement(candles[start:end])
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("upsert candle batch at %d: %w", candles[start].OpenTime, err)
		}
	}
	return tx.Commit()
}

func candleUpsertStatement(candles []Candle) (string, []any) {
	var query strings.Builder
	query.WriteString(`INSERT INTO candles(symbol,timeframe,open_time,open,high,low,close,volume) VALUES `)
	args := make([]any, 0, len(candles)*8)
	for index, candle := range candles {
		if index > 0 {
			query.WriteByte(',')
		}
		base := index*8 + 1
		fmt.Fprintf(&query, "($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", base, base+1, base+2, base+3, base+4, base+5, base+6, base+7)
		args = append(args, candle.Symbol, candle.Timeframe, candle.OpenTime, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
	}
	query.WriteString(` ON CONFLICT(symbol,timeframe,open_time) DO UPDATE SET open=EXCLUDED.open,high=EXCLUDED.high,low=EXCLUDED.low,close=EXCLUDED.close,volume=EXCLUDED.volume`)
	return query.String(), args
}
func (r *PostgresCandleRepository) Range(ctx context.Context, symbol, timeframe string, from, to int64) ([]Candle, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT symbol,timeframe,open_time,open,high,low,close,volume FROM candles WHERE symbol=$1 AND timeframe=$2 AND open_time BETWEEN $3 AND $4 ORDER BY open_time`, symbol, timeframe, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Candle, 0)
	for rows.Next() {
		var c Candle
		if err := rows.Scan(&c.Symbol, &c.Timeframe, &c.OpenTime, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume); err != nil {
			return nil, err
		}
		c.IsClosed = true
		out = append(out, c)
	}
	return out, rows.Err()
}
