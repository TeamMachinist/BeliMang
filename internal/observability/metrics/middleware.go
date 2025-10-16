package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		HTTPActiveRequests.Inc()
		defer HTTPActiveRequests.Dec()

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "not_found"
		}

		status := statusCategory(c.Writer.Status())

		endpoint := path

		HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

// statusCategory converts HTTP status code to category for lower cardinality
// This reduces metric combinations significantly
func statusCategory(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func statusExact(code int) string {
	return strconv.Itoa(code)
}
