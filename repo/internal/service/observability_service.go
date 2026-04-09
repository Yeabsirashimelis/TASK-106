package service

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
)

// Metrics tracks request-level metrics in memory
type Metrics struct {
	mu             sync.RWMutex
	requestCount   int64
	errorCount     int64
	activeSessions int64
	latencies      []time.Duration
	latencyIdx     int
	startTime      time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		latencies: make([]time.Duration, 1000),
		startTime: time.Now(),
	}
}

func (m *Metrics) RecordRequest(duration time.Duration, isError bool) {
	atomic.AddInt64(&m.requestCount, 1)
	if isError {
		atomic.AddInt64(&m.errorCount, 1)
	}
	m.mu.Lock()
	m.latencies[m.latencyIdx%len(m.latencies)] = duration
	m.latencyIdx++
	m.mu.Unlock()
}

func (m *Metrics) IncrSessions() {
	atomic.AddInt64(&m.activeSessions, 1)
}

func (m *Metrics) DecrSessions() {
	atomic.AddInt64(&m.activeSessions, -1)
}

func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := m.latencyIdx
	if count > len(m.latencies) {
		count = len(m.latencies)
	}

	var latencies []time.Duration
	for i := 0; i < count; i++ {
		if m.latencies[i] > 0 {
			latencies = append(latencies, m.latencies[i])
		}
	}

	sortDurations(latencies)

	var p50, p95, p99 time.Duration
	if len(latencies) > 0 {
		p50 = latencies[len(latencies)*50/100]
		p95 = latencies[len(latencies)*95/100]
		p99 = latencies[len(latencies)*99/100]
	}

	return MetricsSnapshot{
		RequestCount:   atomic.LoadInt64(&m.requestCount),
		ErrorCount:     atomic.LoadInt64(&m.errorCount),
		ActiveSessions: atomic.LoadInt64(&m.activeSessions),
		P50Latency:     p50,
		P95Latency:     p95,
		P99Latency:     p99,
		Uptime:         time.Since(m.startTime),
	}
}

type MetricsSnapshot struct {
	RequestCount   int64
	ErrorCount     int64
	ActiveSessions int64
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	Uptime         time.Duration
}

func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j] < d[j-1]; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}

// ObservabilityService provides health and metrics
type ObservabilityService struct {
	db      *sqlx.DB
	metrics *Metrics
}

func NewObservabilityService(db *sqlx.DB, metrics *Metrics) *ObservabilityService {
	return &ObservabilityService{db: db, metrics: metrics}
}

type HealthStatus struct {
	Status     string            `json:"status"`
	Uptime     string            `json:"uptime"`
	Checks     map[string]string `json:"checks"`
	GoRoutines int               `json:"goroutines"`
}

func (s *ObservabilityService) HealthCheck(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Status:     "ok",
		Uptime:     time.Since(s.metrics.startTime).Round(time.Second).String(),
		Checks:     make(map[string]string),
		GoRoutines: runtime.NumGoroutine(),
	}

	if err := s.db.PingContext(ctx); err != nil {
		status.Status = "degraded"
		status.Checks["database"] = "error: " + err.Error()
	} else {
		status.Checks["database"] = "ok"
	}

	dbStats := s.db.Stats()
	status.Checks["db_open_connections"] = fmt.Sprintf("%d", dbStats.OpenConnections)
	status.Checks["db_in_use"] = fmt.Sprintf("%d", dbStats.InUse)
	status.Checks["db_idle"] = fmt.Sprintf("%d", dbStats.Idle)
	status.Checks["disk"] = checkDiskSpace()

	return status
}

type MetricsResponse struct {
	Requests       int64  `json:"request_count"`
	Errors         int64  `json:"error_count"`
	ActiveSessions int64  `json:"active_sessions"`
	P50LatencyMs   int64  `json:"p50_latency_ms"`
	P95LatencyMs   int64  `json:"p95_latency_ms"`
	P99LatencyMs   int64  `json:"p99_latency_ms"`
	UptimeSeconds  int64  `json:"uptime_seconds"`
	DBOpenConns    int    `json:"db_open_connections"`
	DBInUse        int    `json:"db_in_use"`
	DBIdle         int    `json:"db_idle"`
	GoRoutines     int    `json:"goroutines"`
	MemAllocMB     uint64 `json:"mem_alloc_mb"`
}

func (s *ObservabilityService) GetMetrics() MetricsResponse {
	snap := s.metrics.GetSnapshot()
	dbStats := s.db.Stats()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MetricsResponse{
		Requests:       snap.RequestCount,
		Errors:         snap.ErrorCount,
		ActiveSessions: snap.ActiveSessions,
		P50LatencyMs:   snap.P50Latency.Milliseconds(),
		P95LatencyMs:   snap.P95Latency.Milliseconds(),
		P99LatencyMs:   snap.P99Latency.Milliseconds(),
		UptimeSeconds:  int64(snap.Uptime.Seconds()),
		DBOpenConns:    dbStats.OpenConnections,
		DBInUse:        dbStats.InUse,
		DBIdle:         dbStats.Idle,
		GoRoutines:     runtime.NumGoroutine(),
		MemAllocMB:     memStats.Alloc / 1024 / 1024,
	}
}

func (s *ObservabilityService) GetMetricsCollector() *Metrics {
	return s.metrics
}

func checkDiskSpace() string {
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./storage"
	}
	info, err := os.Stat(storagePath)
	if err != nil {
		return "storage path not accessible"
	}
	if !info.IsDir() {
		return "storage path is not a directory"
	}
	return "ok"
}
