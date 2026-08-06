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

func (h *gRPCHanler) CreateTrip(ctx context.Context, req *pb.)


func (h *gRPCHanler) PreviewTrip(ctx context.Context, rq *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pbPickup := rq.GetStartLocation()
	pbDestination := rq.GetEndLocation()

	pickup := &types.Coordinate{
		Latitude:  pbPickup.Latitude,
		Longitude: pbPickup.Longitude,
	}
	destination := &types.Coordinate{
		Latitude:  pbDestination.Latitude,
		Longitude: pbDestination.Longitude,
	}

	r, err := h.service.GetRoute(ctx, pickup, destination, true)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	// 1. Estimate the ride fares prices based on the route (ex: distance)
	estimatedFares := h.service.EstimatePackagesPriceWithRoute(r)
	// 2. Store the ride fares for the create trip to fetch and validate
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, rq.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     r.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}
