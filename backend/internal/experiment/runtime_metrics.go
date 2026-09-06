package experiment

import (
	"math"
	"sort"
	"sync"
	"time"
)

type RuntimeMetrics struct {
	mu                 sync.Mutex
	startedAt          time.Time
	workerCount        int
	runningSince       map[string]time.Time
	completed          uint64
	failed             uint64
	queueWaitSamples   uint64
	totalQueueWait     time.Duration
	queueWaitSamplesMs []float64
	executionSamples   uint64
	totalExecutionTime time.Duration
	executionSamplesMs []float64
}

type RuntimeMetricsSnapshot struct {
	UptimeSeconds          float64    `json:"uptimeSeconds"`
	WorkerCount            int        `json:"workerCount"`
	JobsCompleted          uint64     `json:"jobsCompleted"`
	JobsFailed             uint64     `json:"jobsFailed"`
	JobsPerMinute          float64    `json:"jobsPerMinute"`
	AverageQueueWaitMs     float64    `json:"averageQueueWaitMs"`
	QueueWaitP50Ms         float64    `json:"queueWaitP50Ms"`
	QueueWaitP95Ms         float64    `json:"queueWaitP95Ms"`
	AverageExecutionTimeMs float64    `json:"averageExecutionTimeMs"`
	ExecutionTimeP50Ms     float64    `json:"executionTimeP50Ms"`
	ExecutionTimeP95Ms     float64    `json:"executionTimeP95Ms"`
	Queue                  QueueStats `json:"queue"`
}

func NewRuntimeMetrics(workerCount int) *RuntimeMetrics {
	return &RuntimeMetrics{startedAt: time.Now(), workerCount: workerCount, runningSince: make(map[string]time.Time)}
}

// Observe consumes the same worker state transitions used by WebSocket
// progress, keeping instrumentation outside the backtest business logic.
func (m *RuntimeMetrics) Observe(result Result) {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	switch result.Status {
	case "RUNNING":
		if _, exists := m.runningSince[result.ID]; exists {
			return
		}
		m.runningSince[result.ID] = now
		if result.CreatedAt > 0 {
			wait := now.Sub(time.UnixMilli(result.CreatedAt))
			if wait >= 0 {
				m.totalQueueWait += wait
				m.queueWaitSamples++
				m.queueWaitSamplesMs = appendBounded(m.queueWaitSamplesMs, durationMs(wait))
			}
		}
	case "COMPLETED", "FAILED":
		if started, exists := m.runningSince[result.ID]; exists {
			duration := now.Sub(started)
			m.totalExecutionTime += duration
			m.executionSamples++
			m.executionSamplesMs = appendBounded(m.executionSamplesMs, durationMs(duration))
			delete(m.runningSince, result.ID)
		}
		if result.Status == "COMPLETED" {
			m.completed++
		} else {
			m.failed++
		}
	}
}

func (m *RuntimeMetrics) Snapshot(queue QueueStats) RuntimeMetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	uptime := time.Since(m.startedAt)
	snapshot := RuntimeMetricsSnapshot{
		UptimeSeconds: uptime.Seconds(), WorkerCount: m.workerCount,
		JobsCompleted: m.completed, JobsFailed: m.failed, Queue: queue,
	}
	if uptime > 0 {
		snapshot.JobsPerMinute = float64(m.completed+m.failed) / uptime.Minutes()
	}
	if m.queueWaitSamples > 0 {
		snapshot.AverageQueueWaitMs = float64(m.totalQueueWait.Microseconds()) / 1000 / float64(m.queueWaitSamples)
		snapshot.QueueWaitP50Ms = percentile(m.queueWaitSamplesMs, 0.50)
		snapshot.QueueWaitP95Ms = percentile(m.queueWaitSamplesMs, 0.95)
	}
	if m.executionSamples > 0 {
		snapshot.AverageExecutionTimeMs = float64(m.totalExecutionTime.Microseconds()) / 1000 / float64(m.executionSamples)
		snapshot.ExecutionTimeP50Ms = percentile(m.executionSamplesMs, 0.50)
		snapshot.ExecutionTimeP95Ms = percentile(m.executionSamplesMs, 0.95)
	}
	return snapshot
}

const maxLatencySamples = 1024

func appendBounded(samples []float64, value float64) []float64 {
	if len(samples) == maxLatencySamples {
		copy(samples, samples[1:])
		samples[len(samples)-1] = value
		return samples
	}
	return append(samples, value)
}

func durationMs(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000
}

func percentile(samples []float64, quantile float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	ordered := append([]float64(nil), samples...)
	sort.Float64s(ordered)
	index := int(math.Ceil(float64(len(ordered))*quantile)) - 1
	if index < 0 {
		index = 0
	}
	return ordered[index]
}
