package experiment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) Save(ctx context.Context, r Result) error {
	strategies, err := json.Marshal(r.Strategies)
	if err != nil {
		return fmt.Errorf("marshal strategies: %w", err)
	}
	params, err := json.Marshal(r.Params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	versions, err := json.Marshal(r.StrategyVersions)
	if err != nil {
		return fmt.Errorf("marshal strategy versions: %w", err)
	}

	_, err = p.db.ExecContext(ctx, `
		INSERT INTO experiments (
			id, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET
			candidate_id = EXCLUDED.candidate_id,
			strategies = EXCLUDED.strategies,
			params = EXCLUDED.params,
			policy = EXCLUDED.policy,
			strategy_versions = EXCLUDED.strategy_versions,
			dataset_period = EXCLUDED.dataset_period,
			return_pct = EXCLUDED.return_pct,
			mdd = EXCLUDED.mdd,
			trade_count = EXCLUDED.trade_count,
			win_rate = EXCLUDED.win_rate,
			wins = EXCLUDED.wins,
			losses = EXCLUDED.losses,
			total_profit = EXCLUDED.total_profit,
			status = EXCLUDED.status,
			created_at = EXCLUDED.created_at
	`, r.ID, r.CandidateID, string(strategies), string(params), r.Policy, string(versions),
		r.DatasetPeriod, r.Return, r.MDD, r.TradeCount, r.WinRate, r.Wins, r.Losses,
		r.TotalProfit, r.Status, r.CreatedAt)
	if err != nil {
		return fmt.Errorf("save experiment: %w", err)
	}
	return nil
}

func (p *PostgresRepository) Get(ctx context.Context, id string) (Result, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT id, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at
		FROM experiments WHERE id = $1
	`, id)

	r, err := scanResult(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, ErrNotFound
	}
	if err != nil {
		return Result{}, fmt.Errorf("get experiment: %w", err)
	}
	return r, nil
}

func (p *PostgresRepository) List(ctx context.Context) ([]Result, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at
		FROM experiments ORDER BY return_pct DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list experiments: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		r, err := scanResult(rows)
		if err != nil {
			return nil, fmt.Errorf("scan experiment: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanResult(s scanner) (Result, error) {
	var r Result
	var strategies, params, versions []byte

	err := s.Scan(
		&r.ID, &r.CandidateID, &strategies, &params, &r.Policy, &versions,
		&r.DatasetPeriod, &r.Return, &r.MDD, &r.TradeCount, &r.WinRate, &r.Wins, &r.Losses,
		&r.TotalProfit, &r.Status, &r.CreatedAt,
	)
	if err != nil {
		return Result{}, err
	}

	if err := json.Unmarshal(strategies, &r.Strategies); err != nil {
		return Result{}, fmt.Errorf("unmarshal strategies: %w", err)
	}
	if err := json.Unmarshal(params, &r.Params); err != nil {
		return Result{}, fmt.Errorf("unmarshal params: %w", err)
	}
	if err := json.Unmarshal(versions, &r.StrategyVersions); err != nil {
		return Result{}, fmt.Errorf("unmarshal strategy versions: %w", err)
	}
	return r, nil
}
