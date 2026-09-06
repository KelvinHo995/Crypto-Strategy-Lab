package experiment_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
	"github.com/joho/godotenv"
)

func openQueueRuntimeDB(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("RUN_POSTGRES_QUEUE_INTEGRATION") != "1" {
		t.Skip("set RUN_POSTGRES_QUEUE_INTEGRATION=1 to run destructive queue integration tests against an idle backend database")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		_ = godotenv.Load("../../.env")
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured")
	}
	db, err := experiment.OpenDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestPostgresQueueAtomicEnqueueClaimAndAck(t *testing.T) {
	db := openQueueRuntimeDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	id := fmt.Sprintf("queue-runtime-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 20*time.Second)
		defer stop()
		if _, err := db.ExecContext(cleanup, `DELETE FROM experiment_jobs WHERE id=$1`, id); err != nil {
			t.Errorf("cleanup job %s: %v", id, err)
		}
		if _, err := db.ExecContext(cleanup, `DELETE FROM experiments WHERE id=$1`, id); err != nil {
			t.Errorf("cleanup experiment %s: %v", id, err)
		}
	})

	queue := experiment.NewPostgresQueue(db)
	job := experiment.BacktestJob{ID: id, SearchID: id, SearchTotal: 1, Pair: "BTCUSDT", Timeframe: "1h", From: 1, To: 2,
		Candidate: strategy.CandidateStrategy{ID: "candidate", Instances: []strategy.StrategyInstance{{Type: "MA"}}, Policy: "majority"},
		Config:    experiment.Config{Pair: "BTCUSDT", StartingCapital: 1000}, DatasetPeriod: "1-2", EnqueuedAt: time.Now().UnixMilli()}
	pending := experiment.Result{ID: id, SearchID: id, SearchTotal: 1, CandidateID: "candidate", Instances: []strategy.StrategyInstance{{Type: "MA"}}, Policy: "majority", StrategyVersions: map[string]string{"MA": "v1"}, DatasetPeriod: "1-2", Status: "PENDING", CreatedAt: job.EnqueuedAt}
	if err := queue.EnqueuePending(ctx, job, pending); err != nil {
		t.Fatalf("enqueue pending: %v; apply migrations through 0014", err)
	}
	stored, err := experiment.NewPostgresRepository(db).Get(ctx, id)
	if err != nil || stored.Status != "PENDING" {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	claimed, err := queue.Dequeue(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.ID != id || claimed.SearchID != id || claimed.SearchTotal != 1 || claimed.Pair != "BTCUSDT" || len(claimed.Candles) != 0 {
		t.Fatalf("claimed=%+v", claimed)
	}
	if err := queue.Ack(ctx, id); err != nil {
		t.Fatal(err)
	}
	var jobs int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM experiment_jobs WHERE id=$1`, id).Scan(&jobs); err != nil || jobs != 0 {
		t.Fatalf("jobs=%d err=%v", jobs, err)
	}
}

func TestPostgresQueueExhaustedRetryFailsExperiment(t *testing.T) {
	db := openQueueRuntimeDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	id := fmt.Sprintf("queue-retry-runtime-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 20*time.Second)
		defer stop()
		if _, err := db.ExecContext(cleanup, `DELETE FROM experiment_jobs WHERE id=$1`, id); err != nil {
			t.Errorf("cleanup job %s: %v", id, err)
		}
		if _, err := db.ExecContext(cleanup, `DELETE FROM experiments WHERE id=$1`, id); err != nil {
			t.Errorf("cleanup experiment %s: %v", id, err)
		}
	})

	queue := experiment.NewPostgresQueue(db)
	job := experiment.BacktestJob{ID: id, SearchID: id, SearchTotal: 1, Pair: "BTCUSDT", Timeframe: "1h", From: 1, To: 2,
		Candidate: strategy.CandidateStrategy{ID: "candidate", Instances: []strategy.StrategyInstance{{Type: "MA"}}, Policy: "majority"},
		Config:    experiment.Config{Pair: "BTCUSDT", StartingCapital: 1000}, DatasetPeriod: "1-2", EnqueuedAt: time.Now().UnixMilli()}
	pending := experiment.Result{ID: id, SearchID: id, SearchTotal: 1, CandidateID: "candidate", Instances: []strategy.StrategyInstance{{Type: "MA"}}, Policy: "majority", StrategyVersions: map[string]string{"MA": "v1"}, DatasetPeriod: "1-2", Status: "PENDING", CreatedAt: job.EnqueuedAt}
	if err := queue.EnqueuePending(ctx, job, pending); err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt <= experiment.DefaultQueueMaxAttempts; attempt++ {
		if _, err := queue.Dequeue(ctx); err != nil {
			t.Fatalf("dequeue attempt %d: %v", attempt, err)
		}
		if err := queue.Nack(ctx, id, fmt.Errorf("attempt %d failed", attempt)); err != nil {
			t.Fatalf("nack attempt %d: %v", attempt, err)
		}
		if attempt < experiment.DefaultQueueMaxAttempts {
			if _, err := db.ExecContext(ctx, `UPDATE experiment_jobs SET available_at=0 WHERE id=$1`, id); err != nil {
				t.Fatal(err)
			}
		}
	}

	var jobStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM experiment_jobs WHERE id=$1`, id).Scan(&jobStatus); err != nil {
		t.Fatal(err)
	}
	stored, err := experiment.NewPostgresRepository(db).Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if jobStatus != "FAILED" || stored.Status != "FAILED" {
		t.Fatalf("job status=%s experiment status=%s", jobStatus, stored.Status)
	}
}
