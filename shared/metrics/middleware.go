// Package metrics — HTTP middleware for net/http.
package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Status() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

// MetricMiddleware returns an HTTP middleware that records Prometheus HTTP metrics.
// It must be registered after metrics.Init() is called.
// The /metrics path itself is excluded to avoid self-instrumentation noise.
func MetricMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if HTTPRequestsTotal == nil || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			path := r.URL.Path
			method := r.Method

			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			status := strconv.Itoa(rw.Status())
			duration := time.Since(start).Seconds()

			HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
			HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
		})
	}
}
