package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

func TestReadinessReportsDependencyFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	readiness(func(context.Context) error { return errors.New("database unavailable") }).ServeHTTP(response, req)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "not_ready") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestOperationalMetricsExposeWorkerAndQueueState(t *testing.T) {
	metrics := experiment.NewRuntimeMetrics(3)
	queue := experiment.NewInMemoryQueue(4)
	if err := queue.Enqueue(context.Background(), experiment.BacktestJob{ID: "queued"}); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	response := httptest.NewRecorder()
	operationalMetrics(metrics, queue).ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"workerCount":3`) || !strings.Contains(response.Body.String(), `"queued":1`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestPrometheusMetricsExposeScrapeFormat(t *testing.T) {
	metrics := experiment.NewRuntimeMetrics(3)
	metrics.Observe(experiment.Result{ID: "completed", Status: "COMPLETED"})
	request := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	response := httptest.NewRecorder()
	prometheusMetrics(metrics, experiment.NewInMemoryQueue(1)).ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("response = %d content-type=%q", response.Code, response.Header().Get("Content-Type"))
	}
	body := response.Body.String()
	for _, expected := range []string{"# TYPE crypto_strategy_jobs_completed_total counter", "crypto_strategy_jobs_completed_total 1", "crypto_strategy_queue_wait_p95_milliseconds"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics body missing %q:\n%s", expected, body)
		}
	}
}

func TestRequestTracePreservesOrCreatesID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("X-Request-ID", "caller-id")
	response := httptest.NewRecorder()
	withRequestTrace(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)
	if got := response.Header().Get("X-Request-ID"); got != "caller-id" {
		t.Fatalf("X-Request-ID = %q", got)
	}
}
