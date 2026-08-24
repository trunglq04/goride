package main

import (
	"context"

	pb "github.com/trunglq04/goride/shared/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type driverGrpcHandler struct {
	pb.UnimplementedDriverServiceServer

	service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler := &driverGrpcHandler{
		service: service,
	}

	pb.RegisterDriverServiceServer(s, handler)
}

func (h *driverGrpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.service.RegisterDriver(req.GetDriverID(), req.GetPackageSlug())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to register driver")
	}

	return &pb.RegisterDriverResponse{
		Driver: driver,
	}, nil
}

func (h *driverGrpcHandler) UnregisterDriver(ctx context.Context, req *pb.UnregisterDriverRequest) (*pb.UnregisterDriverResponse, error) {
	h.service.UnregisterDriver(req.GetDriverID())

	return &pb.UnregisterDriverResponse{
		Driver: &pb.Driver{
			Id: req.GetDriverID(),
		},
	}, nil
}

func (h *driverGrpcHandler) UpdateDriverLocation(ctx context.Context, req *pb.UpdateDriverLocationRequest) (*pb.UpdateDriverLocationResponse, error) {
	driver, err := h.service.UpdateDriverLocation(req.GetDriverID(), req.GetLocation(), req.GetGeohash())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update driver location")
	}

	return &pb.UpdateDriverLocationResponse{
		Driver: driver,
	}, nil
}
