package main

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu     sync.Mutex
	clients map[string]*client
}

// authRateLimiter returns a Gin middleware that applies per-IP rate limiting
// to auth endpoints (login, register, OTP). Limits to 5 req/s with a burst of 10.
func authRateLimiter() gin.HandlerFunc {
	rl := &RateLimiter{
		mu:     sync.Mutex{},
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

	return func(c *gin.Context) {
		ip := c.ClientIP()

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
				"path", c.Request.URL.Path,
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}

// requestLogger logs every incoming HTTP request with method, path, status,
// duration and client IP. Level depends on the response status.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		args := []any{
			"http.method", c.Request.Method,
			"http.path", path,
			"http.status", status,
			"duration", time.Since(start).String(),
			"client_ip", c.ClientIP(),
		}
		if rawQuery != "" {
			args = append(args, "http.query", rawQuery)
		}

		ctx := c.Request.Context()
		switch {
		case len(c.Errors) > 0:
			slog.Default().ErrorContext(ctx, "HTTP request failed", append(args, "errors", c.Errors.String())...)
		case status >= http.StatusInternalServerError:
			slog.Default().ErrorContext(ctx, "HTTP request", args...)
		case status >= http.StatusBadRequest:
			slog.Default().WarnContext(ctx, "HTTP request", args...)
		default:
			slog.Default().InfoContext(ctx, "HTTP request", args...)
		}
	}
}

func corsConfig(server *gin.Engine) {
	server.Use(cors.New(cors.Config{
		AllowAllOrigins: true,

		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"Cache-Control",
			"X-Requested-With",
		},

		AllowCredentials: false,

		MaxAge: 12 * time.Hour,
	}))

	server.Use(func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	})
}
