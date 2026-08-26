package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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
