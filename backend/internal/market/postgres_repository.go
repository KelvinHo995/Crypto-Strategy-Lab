package market

import (
	"context"
	"database/sql"
	"fmt"
)

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
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO candles(symbol,timeframe,open_time,open,high,low,close,volume) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(symbol,timeframe,open_time) DO UPDATE SET open=EXCLUDED.open,high=EXCLUDED.high,low=EXCLUDED.low,close=EXCLUDED.close,volume=EXCLUDED.volume`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range candles {
		if _, err = stmt.ExecContext(ctx, c.Symbol, c.Timeframe, c.OpenTime, c.Open, c.High, c.Low, c.Close, c.Volume); err != nil {
			return fmt.Errorf("upsert candle %d: %w", c.OpenTime, err)
		}
	}
	return tx.Commit()
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
