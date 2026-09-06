package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/google/uuid"
)

func readiness(check func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if check != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := check(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "database": "unavailable"})
				return
			}
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ready", "database": "available"})
	}
}

func operationalMetrics(metrics *experiment.RuntimeMetrics, queue experiment.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := readQueueStats(r.Context(), queue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.Snapshot(stats))
	}
}

func prometheusMetrics(metrics *experiment.RuntimeMetrics, queue experiment.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := readQueueStats(r.Context(), queue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		snapshot := metrics.Snapshot(stats)
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		var body strings.Builder
		writeMetric := func(name, help, metricType string, value any) {
			fmt.Fprintf(&body, "# HELP %s %s\n# TYPE %s %s\n%s %v\n", name, help, name, metricType, name, value)
		}
		writeMetric("crypto_strategy_workers", "Configured backtest worker count.", "gauge", snapshot.WorkerCount)
		writeMetric("crypto_strategy_jobs_completed_total", "Completed jobs observed since process start.", "counter", snapshot.JobsCompleted)
		writeMetric("crypto_strategy_jobs_failed_total", "Failed jobs observed since process start.", "counter", snapshot.JobsFailed)
		writeMetric("crypto_strategy_jobs_per_minute", "Average terminal job throughput since process start.", "gauge", snapshot.JobsPerMinute)
		writeMetric("crypto_strategy_queue_jobs", "Jobs currently waiting in the durable queue.", "gauge", snapshot.Queue.Queued)
		writeMetric("crypto_strategy_running_jobs", "Jobs currently claimed by workers.", "gauge", snapshot.Queue.Running)
		writeMetric("crypto_strategy_dead_jobs", "Jobs retained after exhausting retries.", "gauge", snapshot.Queue.Failed)
		writeMetric("crypto_strategy_queue_wait_p50_milliseconds", "Process-lifetime p50 queue wait.", "gauge", snapshot.QueueWaitP50Ms)
		writeMetric("crypto_strategy_queue_wait_p95_milliseconds", "Process-lifetime p95 queue wait.", "gauge", snapshot.QueueWaitP95Ms)
		writeMetric("crypto_strategy_execution_p50_milliseconds", "Process-lifetime p50 backtest execution time.", "gauge", snapshot.ExecutionTimeP50Ms)
		writeMetric("crypto_strategy_execution_p95_milliseconds", "Process-lifetime p95 backtest execution time.", "gauge", snapshot.ExecutionTimeP95Ms)
		_, _ = w.Write([]byte(body.String()))
	}
}

func readQueueStats(parent context.Context, queue experiment.Queue) (experiment.QueueStats, error) {
	provider, ok := queue.(experiment.QueueStatsProvider)
	if !ok {
		return experiment.QueueStats{}, nil
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return provider.Stats(ctx)
}

func logJobTransition(result experiment.Result) {
	if result.Status != "RUNNING" && result.Status != "COMPLETED" && result.Status != "FAILED" {
		return
	}
	log.Printf("experiment search_id=%s job_id=%s candidate_id=%s status=%s trades=%d return_pct=%.4f", result.SearchID, result.ID, result.CandidateID, result.Status, result.TradeCount, result.Return)
}

func withRequestTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" || len(requestID) > 128 {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		started := time.Now()
		if r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			return
		}
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		log.Printf("http request_id=%s method=%s path=%s status=%d bytes=%d duration_ms=%.3f", requestID, r.Method, r.URL.Path, recorder.status, recorder.bytes, float64(time.Since(started).Microseconds())/1000)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(payload []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	written, err := r.ResponseWriter.Write(payload)
	r.bytes += written
	return written, err
}
