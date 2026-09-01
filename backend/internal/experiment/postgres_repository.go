package experiment

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
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
			id, search_id, search_total, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (id) DO UPDATE SET
			search_id = EXCLUDED.search_id,
			search_total = EXCLUDED.search_total,
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
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at
	`, r.ID, r.SearchID, r.SearchTotal, r.CandidateID, string(strategies), string(params), r.Policy, string(versions),
		r.DatasetPeriod, r.Return, r.MDD, r.TradeCount, r.WinRate, r.Wins, r.Losses,
		r.TotalProfit, r.Status, r.CreatedAt, time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("save experiment: %w", err)
	}
	return nil
}

func (p *PostgresRepository) Get(ctx context.Context, id string) (Result, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT id, search_id, search_total, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
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
		SELECT id, search_id, search_total, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
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

func (p *PostgresRepository) ListBySearch(ctx context.Context, searchID string) ([]Result, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, search_id, search_total, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
		FROM experiments WHERE search_id = $1
	`, searchID)
	if err != nil {
		return nil, fmt.Errorf("list experiments by search: %w", err)
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
		&r.ID, &r.SearchID, &r.SearchTotal, &r.CandidateID, &strategies, &params, &r.Policy, &versions,
		&r.DatasetPeriod, &r.Return, &r.MDD, &r.TradeCount, &r.WinRate, &r.Wins, &r.Losses,
		&r.TotalProfit, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return Result{}, err
	}

	strategies, err = normalizeLegacyJSON(strategies)
	if err != nil {
		return Result{}, fmt.Errorf("decode strategies: %w", err)
	}
	params, err = normalizeLegacyJSON(params)
	if err != nil {
		return Result{}, fmt.Errorf("decode params: %w", err)
	}
	versions, err = normalizeLegacyJSON(versions)
	if err != nil {
		return Result{}, fmt.Errorf("decode strategy versions: %w", err)
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

// MarkStaleRunningFailed fails RUNNING rows that haven't been updated in
// olderThan. pg_try_advisory_xact_lock guards it: safe under Supabase's
// transaction-mode pooler (the lock is scoped to and released with this
// transaction, unlike session-level advisory locks), and safe for multiple
// server instances to call concurrently — whichever gets the lock does the
// sweep, everyone else sees it's held and returns immediately, no double work.
func (p *PostgresRepository) MarkStaleRunningFailed(ctx context.Context, olderThan time.Duration) ([]Result, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("mark stale running: begin: %w", err)
	}
	defer tx.Rollback()

	var locked bool
	if err := tx.QueryRowContext(ctx,
		`SELECT pg_try_advisory_xact_lock(hashtext('experiment_stale_sweep')::bigint)`,
	).Scan(&locked); err != nil {
		return nil, fmt.Errorf("mark stale running: lock: %w", err)
	}
	if !locked {
		return nil, nil
	}

	now := time.Now()
	rows, err := tx.QueryContext(ctx, `
		UPDATE experiments SET status = 'FAILED', updated_at = $1
		WHERE status = 'RUNNING' AND updated_at < $2
		RETURNING id, search_id, search_total, candidate_id, strategies, params, policy, strategy_versions,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
	`, now.UnixMilli(), now.Add(-olderThan).UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("mark stale running: update: %w", err)
	}

	var results []Result
	for rows.Next() {
		r, err := scanResult(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("mark stale running: scan: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("mark stale running: rows: %w", err)
	}
	rows.Close()

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("mark stale running: commit: %w", err)
	}
	return results, nil
}

// An earlier pgx integration wrote []byte parameters into TEXT columns. In
// simple protocol pgx encoded those values as PostgreSQL bytea literals such
// as \x5b224d41225d instead of the intended JSON text ["MA"]. Keep reads
// compatible while migration 0003 normalizes existing rows.
func normalizeLegacyJSON(value []byte) ([]byte, error) {
	text := string(value)
	if !strings.HasPrefix(text, `\x`) {
		return value, nil
	}
	decoded, err := hex.DecodeString(text[2:])
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
