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

func (h *gRPCHanler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	// 1. Fetch and validate the fare.
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate fare: %v", err)
	}

	// 2. Call create trip.
	trip, err := h.service.CreateTrip(ctx, rideFare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}
	// 3. Initialize an empty driver to the trip.
	// 4. Add comment at the end of the func to publish an event on the Async Comms module

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}

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

	route, err := h.service.GetRoute(ctx, pickup, destination, true)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	// 1. Estimate the ride fares prices based on the route (ex: distance)
	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)
	// 2. Store the ride fares for the create trip to fetch and validate
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, rq.UserID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}
