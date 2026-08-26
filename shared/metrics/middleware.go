// Package metrics — HTTP middleware for Gin.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// It must be registered after metrics.Init() is called.
// The /metrics path itself is excluded to avoid self-instrumentation noise.
func MetricMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if HTTPRequestsTotal == nil || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
