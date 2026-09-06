package experiment_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

// sequentialGenerator produces a distinct candidate every call (varying
// param values), so dedup never has to retry in tests that don't care about it.
type sequentialGenerator struct{ n int }

func (g *sequentialGenerator) Generate() strategy.CandidateStrategy {
	g.n++
	return strategy.CandidateStrategy{
		ID: fmt.Sprintf("cand-%d", g.n),
		Instances: []strategy.StrategyInstance{
			{Type: "MA", Params: map[string]any{"maShortWindow": g.n}},
		},
		Policy: "majority",
	}
}

// constantGenerator always returns the exact same candidate — used to prove
// dedup forces retries.
type constantGenerator struct{ calls int }

func (g *constantGenerator) Generate() strategy.CandidateStrategy {
	g.calls++
	return strategy.CandidateStrategy{
		ID: "same",
		Instances: []strategy.StrategyInstance{
			{Type: "MA", Params: map[string]any{"maShortWindow": 5}},
		},
		Policy: "majority",
	}
}

type fakePool struct {
	mu        sync.Mutex
	observers map[int]func(experiment.Result)
	nextID    int
}

func (p *fakePool) AddObserver(f func(experiment.Result)) func() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.observers == nil {
		p.observers = map[int]func(experiment.Result){}
	}
	id := p.nextID
	p.nextID++
	p.observers[id] = f
	return func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		delete(p.observers, id)
	}
}

func (p *fakePool) fire(r experiment.Result) {
	p.mu.Lock()
	obs := make([]func(experiment.Result), 0, len(p.observers))
	for _, o := range p.observers {
		obs = append(obs, o)
	}
	p.mu.Unlock()
	for _, o := range obs {
		o(r)
	}
}

func drainAll(t *testing.T, q *experiment.InMemoryQueue) []experiment.BacktestJob {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var jobs []experiment.BacktestJob
	for {
		job, err := q.Dequeue(ctx)
		if err != nil {
			return jobs
		}
		jobs = append(jobs, job)
	}
}

func TestRunSearchLoop_StopsAtMaxCandidates(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(100)
	pool := &fakePool{}
	registry := strategy.NewRegistry()
	if err := registry.RegisterPlugin(strategy.NewMAPlugin(strategy.NewMAStrategy(20, 50))); err != nil {
		t.Fatal(err)
	}

	experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", TimeFrame: "1h", From: 1, To: 2,
		StartingCapital: 1000, DatasetPeriod: "1-2",
		MaxCandidates: 5, VersionResolver: registry,
	}, &sequentialGenerator{}, pool, repo, queue, nil)

	jobs := drainAll(t, queue)
	if len(jobs) != 5 {
		t.Fatalf("enqueued %d jobs, want 5", len(jobs))
	}
	for _, j := range jobs {
		if j.SearchID != "s1" || j.SearchTotal != 5 || j.Pair != "BTCUSDT" || j.Timeframe != "1h" || j.DatasetPeriod != "1-2" {
			t.Fatalf("job %+v missing SearchID/SearchTotal", j)
		}
		if j.StrategyVersions["MA"] != strategy.MAStrategyVersion {
			t.Fatalf("job versions = %v, want descriptor-owned MA version", j.StrategyVersions)
		}
	}
}

func TestRunSearchLoop_DedupesWithinOneRun(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(100)
	pool := &fakePool{}
	gen := &constantGenerator{}

	experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 3,
	}, gen, pool, repo, queue, nil)

	// Every candidate is identical, so dedup retries up to 10 times each
	// and then gives up and accepts the duplicate — still enqueues exactly
	// MaxCandidates jobs, just burns retries doing it.
	jobs := drainAll(t, queue)
	if len(jobs) != 3 {
		t.Fatalf("enqueued %d jobs, want 3", len(jobs))
	}
	if gen.calls < 3 {
		t.Fatalf("generator called %d times, want at least 3 (dedup should force retries)", gen.calls)
	}
}

func TestRunSearchLoop_StopsAtMaxDuration(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(1_000_000)
	pool := &fakePool{}

	start := time.Now()
	experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 1_000_000, MaxDuration: 30 * time.Millisecond,
	}, &sequentialGenerator{}, pool, repo, queue, nil)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("took %s, want to stop near MaxDuration (30ms)", elapsed)
	}
	jobs := drainAll(t, queue)
	if len(jobs) >= 1_000_000 {
		t.Fatalf("enqueued %d jobs, want far fewer than MaxCandidates (proves duration cut it short)", len(jobs))
	}
}

func TestRunSearchLoop_StopsAtNoImprovementLimit(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(1) // buffer=1 forces the generator to
	// pace with dequeues, making completion-vs-generation ordering deterministic.
	pool := &fakePool{}

	done := make(chan struct{})
	go func() {
		experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
			SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
			MaxCandidates: 1000, NoImprovementLimit: 3,
		}, &sequentialGenerator{}, pool, repo, queue, nil)
		close(done)
	}()

	// Candidate 0: a genuine improvement (resets the counter). Candidates
	// after that: no improvement (each increments it). Once the loop
	// decides to stop it may do so without ever producing another job, so
	// each dequeue attempt uses a short timeout and rechecks done rather
	// than blocking forever on a job that isn't coming.
	deadline := time.Now().Add(3 * time.Second)
	first := true
	for {
		select {
		case <-done:
			return // loop stopped on its own — success
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("loop did not stop within the deadline")
		}

		dctx, dcancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		job, err := queue.Dequeue(dctx)
		dcancel()
		if err != nil {
			continue
		}
		result := experiment.Result{ID: job.ID, SearchID: "s1", Status: "COMPLETED"}
		if first {
			result.Return, result.WinRate, result.MDD = 100, 100, 0 // high score
			first = false
		} else {
			result.Return, result.WinRate, result.MDD = 1, 1, 50 // low score, never beats the first
		}
		pool.fire(result)
	}
}

// Proves the generator paces itself against NoImprovementLimit rather than
// relying on the queue to provide backpressure — TestRunSearchLoop_StopsAtNoImprovementLimit
// above uses a buffer of 1, which incidentally forces pacing on its own and
// would pass even without this fix. Here the buffer is enormous (mirroring
// the production Postgres queue, which has no bound at all), and nothing
// ever completes, so a correct implementation must stop generating once
// NoImprovementLimit candidates are outstanding and simply wait, instead of
// racing ahead to MaxCandidates.
func TestRunSearchLoop_NoImprovementCapsInFlightCandidates(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(1_000_000)
	pool := &fakePool{}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	experiment.RunSearchLoop(ctx, experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 1000, NoImprovementLimit: 3,
	}, &sequentialGenerator{}, pool, repo, queue, nil)

	jobs := drainAll(t, queue)
	if len(jobs) > 3 {
		t.Fatalf("enqueued %d jobs with zero completions and NoImprovementLimit=3 — generator should cap in-flight candidates at the limit instead of racing ahead", len(jobs))
	}
}

func TestRunSearchLoop_RespectsContextCancellation(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(1)
	pool := &fakePool{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before the loop even starts

	done := make(chan struct{})
	go func() {
		experiment.RunSearchLoop(ctx, experiment.LoopParams{
			SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
			MaxCandidates: 1000,
		}, &sequentialGenerator{}, pool, repo, queue, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunSearchLoop did not return after context cancellation")
	}
}

// onStopped exists so a caller (the HTTP layer) can tell a consumer the
// search-run is done early — it must fire on every early-stop path and
// never fire when the loop simply generates its full MaxCandidates, since
// the existing tested==total mechanism already covers that case.
func TestRunSearchLoop_CallsOnStoppedForEarlyStop(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(1_000_000)
	pool := &fakePool{}

	var status, reason string
	onStopped := func(s, r string) { status, reason = s, r }

	experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 1_000_000, MaxDuration: 30 * time.Millisecond,
	}, &sequentialGenerator{}, pool, repo, queue, onStopped)

	if status != "STOPPED" || reason != "max-duration" {
		t.Fatalf("onStopped(status=%q, reason=%q), want STOPPED/max-duration", status, reason)
	}
}

func TestRunSearchLoop_OnStoppedReportsFailedWhenNothingEnqueued(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(1)
	pool := &fakePool{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before a single candidate can be enqueued

	var status, reason string
	onStopped := func(s, r string) { status, reason = s, r }

	experiment.RunSearchLoop(ctx, experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 1000,
	}, &sequentialGenerator{}, pool, repo, queue, onStopped)

	if status != "FAILED" || reason != "cancelled" {
		t.Fatalf("onStopped(status=%q, reason=%q), want FAILED/cancelled", status, reason)
	}
}

func TestRunSearchLoop_DoesNotCallOnStoppedWhenCompletingNormally(t *testing.T) {
	repo := newMemRepo()
	queue := experiment.NewInMemoryQueue(100)
	pool := &fakePool{}

	called := false
	onStopped := func(string, string) { called = true }

	experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 5,
	}, &sequentialGenerator{}, pool, repo, queue, onStopped)

	if called {
		t.Fatal("onStopped fired for a loop that generated its full MaxCandidates normally")
	}
}

// pendingQueue implements experiment.PendingQueue on top of InMemoryQueue so
// tests can prove RunSearchLoop prefers the atomic path when available,
// instead of the crash-window-prone Save-then-Enqueue fallback.
type pendingQueue struct {
	*experiment.InMemoryQueue
	pendingCalls int
}

func (q *pendingQueue) EnqueuePending(ctx context.Context, job experiment.BacktestJob, _ experiment.Result) error {
	q.pendingCalls++
	return q.InMemoryQueue.Enqueue(ctx, job)
}

func TestRunSearchLoop_UsesAtomicEnqueuePendingWhenAvailable(t *testing.T) {
	repo := newMemRepo()
	queue := &pendingQueue{InMemoryQueue: experiment.NewInMemoryQueue(100)}
	pool := &fakePool{}

	experiment.RunSearchLoop(context.Background(), experiment.LoopParams{
		SearchID: "s1", Pair: "BTCUSDT", StartingCapital: 1000,
		MaxCandidates: 3,
	}, &sequentialGenerator{}, pool, repo, queue, nil)

	if queue.pendingCalls != 3 {
		t.Fatalf("EnqueuePending called %d times, want 3", queue.pendingCalls)
	}
	if len(repo.results) != 0 {
		t.Fatalf("repo.Save called directly %d times, want 0 — EnqueuePending should own writing the pending row", len(repo.results))
	}
}
