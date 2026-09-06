package main

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/trunglq04/goride/shared/util"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
}

// authRateLimiter returns an HTTP middleware that applies per-IP rate limiting
// to auth endpoints (login, register, OTP). Limits to 5 req/s with a burst of 10.
func authRateLimiter() func(http.Handler) http.Handler {
	rl := &RateLimiter{
		mu:      sync.Mutex{},
		clients: make(map[string]*client),
	}

	// Clean up stale entries every 3 minutes
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			rl.mu.Lock()
			for ip, c := range rl.clients {
				if time.Since(c.lastSeen) > 5*time.Minute {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)

			rl.mu.Lock()
			if _, found := rl.clients[ip]; !found {
				rl.clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(5), 10), // 5 req/s, burst 10
				}
			}
			rl.clients[ip].lastSeen = time.Now()
			limiter := rl.clients[ip].limiter
			rl.mu.Unlock()

			if !limiter.Allow() {
				slog.Warn("Rate limit exceeded for auth endpoint",
					"client_ip", ip,
					"path", r.URL.Path,
				)
				util.WriteError(w, http.StatusTooManyRequests, "Too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// requestLogger logs every incoming HTTP request with method, path, status,
// duration and client IP. Level depends on the response status.
func requestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			rawQuery := r.URL.RawQuery

			rw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			status := rw.status
			args := []any{
				"http.method", r.Method,
				"http.path", path,
				"http.status", status,
				"duration", time.Since(start).String(),
				"client_ip", realIP(r),
			}
			if rawQuery != "" {
				args = append(args, "http.query", rawQuery)
			}

			ctx := r.Context()
			switch {
			case status >= http.StatusInternalServerError:
				slog.Default().ErrorContext(ctx, "HTTP request", args...)
			case status >= http.StatusBadRequest:
				slog.Default().WarnContext(ctx, "HTTP request", args...)
			default:
				slog.Default().InfoContext(ctx, "HTTP request", args...)
			}
		})
	}
}

// corsMiddleware adds CORS headers and handles preflight OPTIONS requests.
func corsMiddleware() func(http.Handler) http.Handler {
	allowedMethods := "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD"
	allowedHeaders := "Origin, Content-Type, Accept, Authorization, Cache-Control, X-Requested-With"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// realIP returns the best-guess client IP from common proxy headers.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the list
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// loggingResponseWriter captures the status code for the request logger.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

// chain applies a list of middlewares to a handler, outermost first.
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// recoveryMiddleware catches panics and returns a 500 response.
func recoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					slog.Error("Panic recovered", "panic", rec, "path", r.URL.Path)
					util.WriteError(w, http.StatusInternalServerError, "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
