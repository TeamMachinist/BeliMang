package database

import (
	"context"
	"strings"
	"sync"
	"time"

	"belimang/internal/observability/metrics"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queryContext struct {
	startTime time.Time
	sql       string
}

// Pool untuk reuse queryContext objects - reduce GC pressure
var queryContextPool = sync.Pool{
	New: func() interface{} {
		return &queryContext{}
	},
}

// MetricsTracer implements pgx.QueryTracer for automatic metrics collection
type MetricsTracer struct{}

// TraceQueryStart is called when a query begins
func (t *MetricsTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// Get from pool instead of allocating new
	qc := queryContextPool.Get().(*queryContext)
	qc.startTime = time.Now()
	qc.sql = data.SQL

	return context.WithValue(ctx, "query_context", qc)
}

// OPTIMIZED: TraceQueryEnd with less overhead
func (t *MetricsTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	qc, ok := ctx.Value("query_context").(*queryContext)
	if !ok {
		return
	}

	duration := time.Since(qc.startTime).Seconds()

	// Get query metadata
	queryType := getQueryType(qc.sql)

	// Only extract table name if needed (not for all query types)
	tableName := "unknown"
	if needsTableName(queryType) {
		tableName = extractTableName(qc.sql)
	}

	// Record metrics
	metrics.DBQueryTotal.WithLabelValues(queryType, tableName).Inc()
	metrics.DBQueryDuration.WithLabelValues(queryType, tableName).Observe(duration)

	// Only record errors (excluding ErrNoRows which is normal)
	if data.Err != nil && data.Err != pgx.ErrNoRows {
		errorCode := "unknown"
		if pgErr, ok := data.Err.(*pgconn.PgError); ok {
			errorCode = pgErr.Code
		}
		metrics.DBQueryErrors.WithLabelValues(queryType, errorCode).Inc()
	}

	// Return queryContext to pool
	queryContextPool.Put(qc)
}

// OPTIMIZED: Check if we need table name extraction
func needsTableName(queryType string) bool {
	switch queryType {
	case "select", "insert", "update", "delete":
		return true
	default:
		return false
	}
}

// OPTIMIZED: Faster query type detection
func getQueryType(sql string) string {
	if len(sql) < 6 {
		return "other"
	}

	// Check first non-whitespace characters
	i := 0
	for i < len(sql) && (sql[i] == ' ' || sql[i] == '\t' || sql[i] == '\n') {
		i++
	}

	if i >= len(sql) {
		return "other"
	}

	// Convert first 6 chars to uppercase for comparison
	var prefix [6]byte
	end := i + 6
	if end > len(sql) {
		end = len(sql)
	}

	for j := i; j < end; j++ {
		c := sql[j]
		if c >= 'a' && c <= 'z' {
			c -= 32 // Convert to uppercase
		}
		prefix[j-i] = c
	}

	// Fast comparison using first characters
	switch prefix[0] {
	case 'S': // SELECT
		if prefix[1] == 'E' {
			return "select"
		}
	case 'I': // INSERT
		if prefix[1] == 'N' {
			return "insert"
		}
	case 'U': // UPDATE
		if prefix[1] == 'P' {
			return "update"
		}
	case 'D': // DELETE
		if prefix[1] == 'E' {
			return "delete"
		}
	case 'B': // BEGIN
		if prefix[1] == 'E' {
			return "begin"
		}
	case 'C': // COMMIT
		if prefix[1] == 'O' {
			return "commit"
		}
	case 'R': // ROLLBACK
		if prefix[1] == 'O' {
			return "rollback"
		}
	case 'W': // WITH (CTE)
		if prefix[1] == 'I' {
			// For WITH queries, check for INSERT/UPDATE/DELETE
			upper := strings.ToUpper(sql)
			if strings.Contains(upper, "INSERT") {
				return "insert"
			} else if strings.Contains(upper, "UPDATE") {
				return "update"
			} else if strings.Contains(upper, "DELETE") {
				return "delete"
			}
			return "select"
		}
	}

	return "other"
}

// OPTIMIZED: Faster table name extraction
func extractTableName(sql string) string {
	upper := strings.ToUpper(sql)

	// Define search patterns with their keyword lengths
	patterns := []struct {
		keyword string
		length  int
	}{
		{"FROM ", 5},
		{"INTO ", 5},
		{"UPDATE ", 7},
		{"TABLE ", 6},
		{"JOIN ", 5},
	}

	for _, pattern := range patterns {
		idx := strings.Index(upper, pattern.keyword)
		if idx == -1 {
			continue
		}

		start := idx + pattern.length
		if start >= len(sql) {
			continue
		}

		// Skip whitespace
		for start < len(sql) && (sql[start] == ' ' || sql[start] == '\t' || sql[start] == '\n') {
			start++
		}

		if start >= len(sql) {
			continue
		}

		// Find end of table name
		end := start
		for end < len(sql) {
			c := sql[end]
			if c == ' ' || c == '\n' || c == '\t' || c == ',' || c == ';' || c == '(' || c == ')' {
				break
			}
			end++
		}

		if end > start {
			tableName := sql[start:end]

			// Remove quotes if present
			tableName = strings.Trim(tableName, "\"'`")

			// If schema.table format, get just table name
			if dotIdx := strings.LastIndex(tableName, "."); dotIdx != -1 {
				tableName = tableName[dotIdx+1:]
			}

			// Filter out keywords
			if tableName != "" && tableName != "ONLY" && !isKeyword(strings.ToUpper(tableName)) {
				return tableName
			}
		}
	}

	return "unknown"
}

// Helper to check common SQL keywords
func isKeyword(s string) bool {
	switch s {
	case "SELECT", "FROM", "WHERE", "AND", "OR", "NOT", "IN", "EXISTS",
		"INNER", "OUTER", "LEFT", "RIGHT", "FULL", "CROSS", "NATURAL",
		"ONLY", "AS", "ON", "USING":
		return true
	}
	return false
}
