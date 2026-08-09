package main

import (
	"context"

	pb "github.com/trunglq04/goride/shared/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedDriverServiceServer

	Service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler := &gRPCHandler{
		Service: service,
	}

	pb.RegisterDriverServiceServer(s, handler)
}

func (h *gRPCHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterDriver not implemented")
}
func (h *gRPCHandler) UnregisterDriver(ctx context.Context, req *pb.UnregisterDriverRequest) (*pb.UnregisterDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UnregisterDriver not implemented")
}
