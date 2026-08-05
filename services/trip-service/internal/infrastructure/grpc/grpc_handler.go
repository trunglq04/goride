package grpc

import (
	"context"
	"log"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	pb "github.com/trunglq04/goride/shared/proto/trip"
	types "github.com/trunglq04/goride/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHanler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHanler {
	handler := &gRPCHanler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHanler) PreviewTrip(ctx context.Context, rq *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := rq.GetStartLocation()
	destination := rq.GetEndLocation()

	pickupCoord := &types.Coordinate{
		Latitude:  pickup.Latitude,
		Longitude: pickup.Longitude,
	}
	destinationCoord := &types.Coordinate{
		Latitude:  destination.Latitude,
		Longitude: destination.Longitude,
	}

	r, err := h.service.GetRoute(ctx, pickupCoord, destinationCoord, true)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     r.ToProto(),
		RideFares: []*pb.RideFare{},
	}, nil
}
