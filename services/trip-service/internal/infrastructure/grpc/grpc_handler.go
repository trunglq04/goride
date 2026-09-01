package grpc

import (
	"context"
	"time"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/logger"
	pb "github.com/trunglq04/goride/shared/proto/trip"
	types "github.com/trunglq04/goride/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer

	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
}

func (h *gRPCHandler) CancelTrip(ctx context.Context, req *pb.CancelTripRequest) (*pb.CancelTripResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := h.service.CancelTrip(ctx, req.UserID, req.TripID, req.Reason)
	if err != nil {
		logger.L().ErrorContext(ctx, "Failed to cancel trip",
			"user_id", req.UserID,
			"trip_id", req.TripID,
			"err", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to cancel trip: %v", err)
	}

	logger.L().InfoContext(ctx, "Trip canceled",
		"user_id", req.UserID,
		"trip_id", req.TripID,
	)
	return &pb.CancelTripResponse{
		TripID: req.TripID,
		Status: "CANCELED",
	}, nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	// 1. Fetch and validate the fare.
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to validate fare: %v", err)
	}

	// 2. Call create trip.
	trip, err := h.service.CreateTrip(ctx, rideFare)
	if err != nil {
		logger.L().ErrorContext(ctx, "Failed to create trip",
			"user_id", userID,
			"fare_id", fareID,
			"err", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to create trip: %v", err)
	}
	logger.L().InfoContext(ctx, "Trip created",
		"trip_id", trip.ID.Hex(),
		"user_id", userID,
		"fare_id", fareID,
	)

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, rq *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
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

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	route, err := h.service.GetRoute(ctx, pickup, destination, true)
	if err != nil {
		logger.L().ErrorContext(ctx, "Failed to get route",
			"user_id", rq.UserID,
			"pickup", pickup,
			"destination", destination,
			"err", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to get route: %v", err)
	}

	// 1. Estimate the ride fares prices based on the route (ex: distance)
	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)
	// 2. Store the ride fares for the create trip to fetch and validate
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, rq.UserID, route)
	if err != nil {
		logger.L().ErrorContext(ctx, "Failed to generate trip fares",
			"user_id", rq.UserID,
			"err", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to get route: %v", err)
	}

	logger.L().InfoContext(ctx, "Trip preview generated",
		"user_id", rq.UserID,
		"distance", route.Routes[0].Distance,
		"duration", route.Routes[0].Duration,
		"fares", len(fares),
	)

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}
