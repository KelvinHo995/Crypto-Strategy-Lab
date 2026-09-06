package experiment

import (
	"testing"
	"time"
)

func TestRuntimeMetricsTracksWorkerLifecycle(t *testing.T) {
	metrics := NewRuntimeMetrics(3)
	createdAt := time.Now().Add(-20 * time.Millisecond).UnixMilli()
	metrics.Observe(Result{ID: "job-1", Status: "RUNNING", CreatedAt: createdAt})
	time.Sleep(time.Millisecond)
	metrics.Observe(Result{ID: "job-1", Status: "COMPLETED", CreatedAt: createdAt})
	metrics.Observe(Result{ID: "job-2", Status: "FAILED", CreatedAt: createdAt})

	snapshot := metrics.Snapshot(QueueStats{Queued: 2, Running: 1, Failed: 4})
	if snapshot.WorkerCount != 3 || snapshot.JobsCompleted != 1 || snapshot.JobsFailed != 1 {
		t.Fatalf("unexpected counters: %+v", snapshot)
	}
	if snapshot.AverageQueueWaitMs <= 0 || snapshot.QueueWaitP50Ms <= 0 || snapshot.QueueWaitP95Ms <= 0 ||
		snapshot.AverageExecutionTimeMs <= 0 || snapshot.ExecutionTimeP50Ms <= 0 || snapshot.ExecutionTimeP95Ms <= 0 || snapshot.JobsPerMinute <= 0 {
		t.Fatalf("expected positive measurements: %+v", snapshot)
	}
	if snapshot.Queue.Queued != 2 || snapshot.Queue.Running != 1 || snapshot.Queue.Failed != 4 {
		t.Fatalf("unexpected queue stats: %+v", snapshot.Queue)
	}
}
