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

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type PostgresRepository struct {
	db *sql.DB
}

const DefaultLeaderboardLimit = 100

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) Save(ctx context.Context, r Result) error {
	return saveResult(ctx, p.db, r)
}

type resultExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func saveResult(ctx context.Context, executor resultExecer, r Result) error {
	searchID := r.SearchID
	if searchID == "" {
		searchID = r.ID
	}
	searchTotal := r.SearchTotal
	if searchTotal < 1 {
		searchTotal = 1
	}
	instances, err := json.Marshal(r.Instances)
	if err != nil {
		return fmt.Errorf("marshal instances: %w", err)
	}
	versions, err := json.Marshal(r.StrategyVersions)
	if err != nil {
		return fmt.Errorf("marshal strategy versions: %w", err)
	}
	models := r.SentimentModels
	if models == nil {
		models = []strategy.SentimentModelIdentity{}
	}
	sentimentModels, err := json.Marshal(models)
	if err != nil {
		return fmt.Errorf("marshal sentiment models: %w", err)
	}

	_, err = executor.ExecContext(ctx, `
		INSERT INTO experiments (
			id, search_id, search_total, candidate_id, instances, policy, strategy_versions, sentiment_models,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (id) DO UPDATE SET
			search_id = EXCLUDED.search_id,
			search_total = EXCLUDED.search_total,
			candidate_id = EXCLUDED.candidate_id,
			instances = EXCLUDED.instances,
			policy = EXCLUDED.policy,
			strategy_versions = EXCLUDED.strategy_versions,
			sentiment_models = EXCLUDED.sentiment_models,
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
	`, r.ID, searchID, searchTotal, r.CandidateID, string(instances), r.Policy, string(versions), string(sentimentModels),
		r.DatasetPeriod, r.Return, r.MDD, r.TradeCount, r.WinRate, r.Wins, r.Losses,
		r.TotalProfit, r.Status, r.CreatedAt, time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("save experiment: %w", err)
	}
	return nil
}

func (p *PostgresRepository) Get(ctx context.Context, id string) (Result, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT id, search_id, search_total, candidate_id, instances, policy, strategy_versions, sentiment_models,
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
		SELECT id, search_id, search_total, candidate_id, instances, policy, strategy_versions, sentiment_models,
			dataset_period, return_pct, mdd, trade_count, win_rate, wins, losses,
			total_profit, status, created_at, updated_at
		FROM experiments
		ORDER BY
			CASE WHEN status = 'COMPLETED' THEN 0 ELSE 1 END,
			(0.50 * COALESCE(return_pct, 0) + 0.30 * COALESCE(win_rate, 0) - 0.20 * COALESCE(mdd, 0)) DESC,
			created_at DESC
		LIMIT $1
	`, DefaultLeaderboardLimit)
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
		SELECT id, search_id, search_total, candidate_id, instances, policy, strategy_versions, sentiment_models,
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

func (p *PostgresRepository) SaveTrades(ctx context.Context, experimentID string, trades []Trade) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save trades: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM experiment_trades WHERE experiment_id = $1`, experimentID); err != nil {
		return fmt.Errorf("clear existing trades: %w", err)
	}
	for _, t := range trades {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO experiment_trades (
				experiment_id, pair, entry_time, direction, volume_usd, entry_price,
				stop_loss, take_profit, exit_price, exit_time, transaction_cost, slippage, profit
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		`, experimentID, t.Pair, t.EntryTime, string(t.Direction), t.VolumeUSD, t.EntryPrice,
			t.StopLoss, t.TakeProfit, t.ExitPrice, t.ExitTime, t.TransactionCost, t.Slippage, t.Profit); err != nil {
			return fmt.Errorf("insert trade: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit save trades: %w", err)
	}
	return nil
}

func (p *PostgresRepository) ListTrades(ctx context.Context, experimentID string) ([]Trade, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT pair, entry_time, direction, volume_usd, entry_price, stop_loss, take_profit,
			exit_price, exit_time, transaction_cost, slippage, profit
		FROM experiment_trades WHERE experiment_id = $1 ORDER BY entry_time ASC
	`, experimentID)
	if err != nil {
		return nil, fmt.Errorf("list trades: %w", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var t Trade
		var direction string
		if err := rows.Scan(&t.Pair, &t.EntryTime, &direction, &t.VolumeUSD, &t.EntryPrice,
			&t.StopLoss, &t.TakeProfit, &t.ExitPrice, &t.ExitTime, &t.TransactionCost, &t.Slippage, &t.Profit); err != nil {
			return nil, fmt.Errorf("scan trade: %w", err)
		}
		t.Direction = Direction(direction)
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanResult(s scanner) (Result, error) {
	var r Result
	var instances, versions, sentimentModels []byte

	err := s.Scan(
		&r.ID, &r.SearchID, &r.SearchTotal, &r.CandidateID, &instances, &r.Policy, &versions, &sentimentModels,
		&r.DatasetPeriod, &r.Return, &r.MDD, &r.TradeCount, &r.WinRate, &r.Wins, &r.Losses,
		&r.TotalProfit, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return Result{}, err
	}

	instances, err = normalizeLegacyJSON(instances)
	if err != nil {
		return Result{}, fmt.Errorf("decode instances: %w", err)
	}
	versions, err = normalizeLegacyJSON(versions)
	if err != nil {
		return Result{}, fmt.Errorf("decode strategy versions: %w", err)
	}
	sentimentModels, err = normalizeLegacyJSON(sentimentModels)
	if err != nil {
		return Result{}, fmt.Errorf("decode sentiment models: %w", err)
	}

	if err := json.Unmarshal(instances, &r.Instances); err != nil {
		return Result{}, fmt.Errorf("unmarshal instances: %w", err)
	}
	if err := json.Unmarshal(versions, &r.StrategyVersions); err != nil {
		return Result{}, fmt.Errorf("unmarshal strategy versions: %w", err)
	}
	if err := json.Unmarshal(sentimentModels, &r.SentimentModels); err != nil {
		return Result{}, fmt.Errorf("unmarshal sentiment models: %w", err)
	}
	return r, nil
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
