package tracing

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

// WrapHandler wraps a standard http.HandlerFunc with an OpenTelemetry span.
func WrapHandler(operation string, handler http.HandlerFunc) http.HandlerFunc {
	tracer := otel.GetTracerProvider().Tracer(operation)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(
			r.Context(),
			operation,
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.HTTPRouteKey.String(r.URL.Path),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer func() {
			if rec := recover(); rec != nil {
				span.RecordError(fmt.Errorf("panic: %v", rec))
				span.SetStatus(codes.Error, "panic")
				span.End()
				panic(rec) // re-panic so recovery middleware can handle the HTTP response
			}
			span.End()
		}()

		// Inject span context back so the handler (and downstream calls) can use it
		rw := &statusResponseWriter{ResponseWriter: w}
		handler(rw, r.WithContext(ctx))

		// Record response status after handler finishes
		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}
		span.SetAttributes(semconv.HTTPResponseStatusCodeKey.Int(status))
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
	}
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// statusResponseWriter captures the HTTP status code for span recording.
type statusResponseWriter struct {
	http.ResponseWriter
	status int
}
