package experiment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type PostgresQueue struct {
	db           *sql.DB
	pollInterval time.Duration
	lease        time.Duration
	maxAttempts  int
	notify       chan struct{}
}

func NewPostgresQueue(db *sql.DB) *PostgresQueue {
	return &PostgresQueue{db: db, pollInterval: DefaultQueuePollInterval, lease: DefaultQueueLease, maxAttempts: DefaultQueueMaxAttempts, notify: make(chan struct{}, 1)}
}

func (q *PostgresQueue) Enqueue(ctx context.Context, job BacktestJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal experiment job: %w", err)
	}
	_, err = q.db.ExecContext(ctx, `
		INSERT INTO experiment_jobs (id, payload, status, attempts, available_at, created_at)
		VALUES ($1, $2, 'QUEUED', 0, $3, $3)
		ON CONFLICT (id) DO NOTHING
	`, job.ID, string(payload), time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("enqueue experiment job: %w", err)
	}
	q.signal()
	return nil
}

func (q *PostgresQueue) EnqueuePending(ctx context.Context, job BacktestJob, pending Result) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal experiment job: %w", err)
	}
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("enqueue pending experiment: begin: %w", err)
	}
	defer tx.Rollback()
	if err := saveResult(ctx, tx, pending); err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO experiment_jobs (id, payload, status, attempts, available_at, created_at)
		VALUES ($1, $2, 'QUEUED', 0, $3, $3)
		ON CONFLICT (id) DO NOTHING
	`, job.ID, string(payload), now); err != nil {
		return fmt.Errorf("enqueue pending experiment: insert job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("enqueue pending experiment: commit: %w", err)
	}
	q.signal()
	return nil
}

func (q *PostgresQueue) Dequeue(ctx context.Context) (BacktestJob, error) {
	for {
		job, err := q.claim(ctx)
		if err == nil {
			return job, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return BacktestJob{}, err
		}
		timer := time.NewTimer(q.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return BacktestJob{}, ctx.Err()
		case <-q.notify:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (q *PostgresQueue) claim(ctx context.Context) (BacktestJob, error) {
	now := time.Now()
	row := q.db.QueryRowContext(ctx, `
		WITH exhausted AS (
			UPDATE experiment_jobs
			SET status = 'FAILED', locked_at = NULL,
				last_error = COALESCE(last_error, 'worker lease expired after maximum attempts')
			WHERE status = 'RUNNING' AND attempts >= $1 AND locked_at < $3
			RETURNING id
		), failed_results AS (
			UPDATE experiments SET status = 'FAILED', updated_at = $2
			WHERE id IN (SELECT id FROM exhausted) AND status IN ('PENDING', 'RUNNING')
		), next_job AS (
			SELECT id
			FROM experiment_jobs
			WHERE attempts < $1
			  AND ((status = 'QUEUED' AND available_at <= $2)
			       OR (status = 'RUNNING' AND locked_at < $3))
			ORDER BY available_at, created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE experiment_jobs AS jobs
		SET status = 'RUNNING', attempts = jobs.attempts + 1, locked_at = $2, last_error = NULL
		FROM next_job
		WHERE jobs.id = next_job.id
		RETURNING jobs.payload
	`, q.maxAttempts, now.UnixMilli(), now.Add(-q.lease).UnixMilli())
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return BacktestJob{}, err
	}
	var job BacktestJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return BacktestJob{}, fmt.Errorf("decode experiment job: %w", err)
	}
	return job, nil
}

func (q *PostgresQueue) Ack(ctx context.Context, id string) error {
	if _, err := q.db.ExecContext(ctx, `DELETE FROM experiment_jobs WHERE id = $1`, id); err != nil {
		return fmt.Errorf("ack experiment job: %w", err)
	}
	return nil
}

func (q *PostgresQueue) Nack(ctx context.Context, id string, cause error) error {
	message := "worker failure"
	if cause != nil {
		message = cause.Error()
	}
	now := time.Now().UnixMilli()
	var affected int
	err := q.db.QueryRowContext(ctx, `
		WITH retried AS (
			UPDATE experiment_jobs
			SET status = CASE WHEN attempts >= $2 THEN 'FAILED' ELSE 'QUEUED' END,
				available_at = $3::bigint + LEAST(attempts::bigint * 1000, 30000),
				locked_at = NULL,
				last_error = $4
			WHERE id = $1
			RETURNING id, status
		), failed_result AS (
			UPDATE experiments SET status = 'FAILED', updated_at = $3
			WHERE id IN (SELECT id FROM retried WHERE status = 'FAILED')
			  AND status IN ('PENDING', 'RUNNING')
		)
		SELECT COUNT(*) FROM retried
	`, id, q.maxAttempts, now, message).Scan(&affected)
	if err != nil {
		return fmt.Errorf("nack experiment job: %w", err)
	}
	if affected > 0 {
		q.signal()
	}
	return nil
}

func (q *PostgresQueue) Heartbeat(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	if _, err := q.db.ExecContext(ctx, `
		WITH touched AS (
			UPDATE experiment_jobs SET locked_at = $2
			WHERE id = $1 AND status = 'RUNNING'
			RETURNING id
		)
		UPDATE experiments SET updated_at = $2
		WHERE id IN (SELECT id FROM touched) AND status = 'RUNNING'
	`, id, now); err != nil {
		return fmt.Errorf("heartbeat experiment job: %w", err)
	}
	return nil
}

func (q *PostgresQueue) Stats(ctx context.Context) (QueueStats, error) {
	var stats QueueStats
	err := q.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'QUEUED'),
			COUNT(*) FILTER (WHERE status = 'RUNNING'),
			COUNT(*) FILTER (WHERE status = 'FAILED')
		FROM experiment_jobs
	`).Scan(&stats.Queued, &stats.Running, &stats.Failed)
	if err != nil {
		return QueueStats{}, fmt.Errorf("read experiment queue stats: %w", err)
	}
	return stats, nil
}

func (q *PostgresQueue) signal() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}
