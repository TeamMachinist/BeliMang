package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// =============================================================================
// HTTP Metrics (Automatic via Middleware)
// OPTIMIZED: Reduced cardinality for high RPS
// =============================================================================

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "belimang",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "belimang",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds",
			// OPTIMIZED: Fewer buckets = less memory + faster lookup
			// Focus on relevant latencies for your SLA
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5}, // 5ms to 2.5s
		},
		[]string{"method", "endpoint"},
	)

	// OPTIONAL: Consider removing if not critical
	// At 60k RPS, these add ~120k metric updates/sec
	// HTTPRequestSize = promauto.NewHistogramVec(
	// 	prometheus.HistogramOpts{
	// 		Namespace: "belimang",
	// 		Subsystem: "http",
	// 		Name:      "request_size_bytes",
	// 		Help:      "HTTP request size in bytes",
	// 		Buckets:   []float64{100, 1000, 10000, 100000, 1000000}, // Simplified buckets
	// 	},
	// 	[]string{"method", "endpoint"},
	// )

	// HTTPResponseSize = promauto.NewHistogramVec(
	// 	prometheus.HistogramOpts{
	// 		Namespace: "belimang",
	// 		Subsystem: "http",
	// 		Name:      "response_size_bytes",
	// 		Help:      "HTTP response size in bytes",
	// 		Buckets:   []float64{100, 1000, 10000, 100000, 1000000}, // Simplified buckets
	// 	},
	// 	[]string{"method", "endpoint"},
	// )

	HTTPActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "http",
			Name:      "active_requests",
			Help:      "Number of currently active HTTP requests",
		},
	)
)

// =============================================================================
// Database Metrics (Automatic via pgx tracer)
// OPTIMIZED: Keep essential metrics only
// =============================================================================

var (
	// Connection Pool Metrics - KEEP (low cardinality)
	DBConnectionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "connections_total",
			Help:      "Total number of database connections in the pool",
		},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "connections_active",
			Help:      "Number of active (in-use) database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "connections_idle",
			Help:      "Number of idle database connections",
		},
	)

	DBMaxOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "max_open_connections",
			Help:      "Maximum number of open connections configured",
		},
	)

	DBConnectionWaitDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "connection_wait_duration_seconds",
			Help:      "Time spent waiting for a database connection",
			// OPTIMIZED: Fewer buckets for sub-millisecond precision
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5},
		},
	)

	// Query Metrics - CONSIDER DISABLING table label if too many tables
	// High cardinality = memory pressure
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "query_duration_seconds",
			Help:      "Database query execution duration in seconds",
			// OPTIMIZED: Focus on query performance range
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		// OPTION 1: Keep table name (good for debugging, but higher cardinality)
		[]string{"query_type", "table"},
		// OPTION 2: Remove table name if you have >20 tables (lower cardinality)
		// []string{"query_type"},
	)

	DBQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "query_total",
			Help:      "Total number of database queries executed",
		},
		[]string{"query_type", "table"}, // Or just {"query_type"}
	)

	DBQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "query_errors_total",
			Help:      "Total number of database query errors",
		},
		[]string{"query_type", "error_code"},
	)

	// Transaction Metrics - SIMPLIFIED
	// Remove if you don't use WithTransaction helper
	DBTransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "transaction_duration_seconds",
			Help:      "Database transaction duration in seconds",
			Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"status"}, // commit, rollback
	)

	DBTransactionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "belimang",
			Subsystem: "database",
			Name:      "transaction_total",
			Help:      "Total number of database transactions",
		},
		[]string{"status"}, // commit, rollback, error
	)

	// OPTIONAL: Remove if health checks are rare
	// DBHealthCheckErrors = promauto.NewCounterVec(
	// 	prometheus.CounterOpts{
	// 		Namespace: "belimang",
	// 		Subsystem: "database",
	// 		Name:      "health_check_errors_total",
	// 		Help:      "Total number of database health check errors",
	// 	},
	// 	[]string{"check_type"},
	// )
)

// =============================================================================
// Application Metrics (Low frequency updates)
// =============================================================================

var (
	AppUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "app",
			Name:      "uptime_seconds",
			Help:      "Application uptime in seconds",
		},
	)

	AppInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "app",
			Name:      "info",
			Help:      "Application information",
		},
		[]string{"version", "env", "go_version"},
	)
)

// =============================================================================
// CRITICAL: Add this for Redis metrics (currently missing!)
// =============================================================================

var (
	// Redis metrics - OPTIONAL but recommended
	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "belimang",
			Subsystem: "redis",
			Name:      "operation_duration_seconds",
			Help:      "Redis operation duration in seconds",
			Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
		},
		[]string{"operation"}, // get, set, del, etc
	)

	RedisOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "belimang",
			Subsystem: "redis",
			Name:      "operation_total",
			Help:      "Total number of Redis operations",
		},
		[]string{"operation", "status"}, // status: ok, error
	)

	RedisPoolStats = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "belimang",
			Subsystem: "redis",
			Name:      "pool_connections",
			Help:      "Redis connection pool statistics",
		},
		[]string{"state"}, // total, idle, active
	)
)
