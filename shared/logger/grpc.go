package logger

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcUnaryServerInterceptor returns a gRPC unary interceptor that logs every
// incoming RPC with its method, duration and status code.
func GrpcUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		logGrpcCall(ctx, start, info.FullMethod, status.Code(err), err)
		return resp, err
	}
}

// GrpcStreamServerInterceptor returns a gRPC stream interceptor that logs
// every incoming streaming RPC when the stream finishes.
func GrpcStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		logGrpcCall(ss.Context(), start, info.FullMethod, status.Code(err), err)
		return err
	}
}

func logGrpcCall(ctx context.Context, start time.Time, method string, code codes.Code, err error) {
	args := []any{
		"grpc.method", method,
		"grpc.code", code.String(),
		"duration", time.Since(start).String(),
	}

	switch {
	case err == nil:
		slog.Default().InfoContext(ctx, "gRPC request", args...)
	case code == codes.Internal || code == codes.Unknown:
		slog.Default().ErrorContext(ctx, "gRPC request failed", append(args, "err", err)...)
	default:
		slog.Default().WarnContext(ctx, "gRPC request rejected", append(args, "err", err)...)
	}
}
