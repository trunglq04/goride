package tracing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func WrapHandler(operation string, handler gin.HandlerFunc) gin.HandlerFunc {
	tracer := otel.GetTracerProvider().Tracer(operation)
	return func(c *gin.Context) {
		ctx, span := tracer.Start(
			c.Request.Context(),
			operation,
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Request.Method),
				semconv.HTTPRouteKey.String(c.FullPath()),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer func() {
			if r := recover(); r != nil {
				span.RecordError(fmt.Errorf("panic: %v", r))
				span.SetStatus(codes.Error, "panic")
				span.End()
				panic(r) // re-panic so gin.Recovery() still handles the HTTP response
			}
			span.End()
		}()

		// Inject span context back so the handler (and downstream calls) can use it
		c.Request = c.Request.WithContext(ctx)

		handler(c)

		// Record response status after handler finishes
		status := c.Writer.Status()
		span.SetAttributes(semconv.HTTPResponseStatusCodeKey.Int(status))
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
	}
}
