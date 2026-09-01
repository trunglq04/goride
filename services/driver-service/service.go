package main

import (
	"context"
	"encoding/json"
	"log/slog"
	math "math/rand/v2"

	"github.com/mmcloughlin/geohash"

	"sync"

	"github.com/trunglq04/goride/shared/messaging"
	pb "github.com/trunglq04/goride/shared/proto/driver"
	"github.com/trunglq04/goride/shared/util"
)

type DriverEventPublisher interface {
	PublishTripRequest(ctx context.Context, driverID string, payloadBytes []byte) error
	PublishTripCanceled(ctx context.Context, driverID string, payloadBytes []byte) error
	PublishNoDriversFound(ctx context.Context, userID string, payloadBytes []byte) error
}

type Service struct {
	drivers      map[string]*driverInMap
	tripToDriver map[string]string
	mu           sync.RWMutex
	publisher    DriverEventPublisher
}

type driverInMap struct {
	Driver *pb.Driver
	// TODO: route
}

func NewService(publisher DriverEventPublisher) *Service {
	return &Service{
		drivers:      make(map[string]*driverInMap),
		tripToDriver: make(map[string]string),
		publisher:    publisher,
	}
}

func (s *Service) FindAvailableDrivers(packageType string, excludedDriverIDs []string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	excluded := make(map[string]bool, len(excludedDriverIDs))
	for _, id := range excludedDriverIDs {
		excluded[id] = true
	}

	var matchingDrivers []string
	for _, driver := range s.drivers {
		if driver.Driver.PackageSlug == packageType && !excluded[driver.Driver.Id] {
			matchingDrivers = append(matchingDrivers, driver.Driver.Id)
		}
	}

	if len(matchingDrivers) == 0 {
		return []string{}
	}

	return matchingDrivers
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	randomIndex := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndex]

	randomPlate := GenerateRandomPlate()
	randomAvatar := util.GetRandomAvatar(randomIndex)

	// we can ignore this property for now, but it must be sent to the frontend.
	geohashStr := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Id:             driverId,
		Geohash:        geohashStr,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Happy Driver",
		PackageSlug:    packageSlug,
		ProfilePicture: randomAvatar,
		CarPlate:       randomPlate,
	}

	s.mu.Lock()
	s.drivers[driverId] = &driverInMap{
		Driver: driver,
	}
	s.mu.Unlock()

	slog.Info("Driver registered",
		"driver_id", driver.Id,
		"package_slug", driver.PackageSlug,
	)

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.drivers, driverId)
	if _, ok := s.drivers[driverId]; !ok {
		slog.Info("Driver unregistered", "driver_id", driverId)
	}
}

func (s *Service) UpdateDriverLocation(driverId string, location *pb.Location, geohash string) (*pb.Driver, error) {
	driver := &pb.Driver{
		Id:             driverId,
		Geohash:        geohash,
		Location:       location,
		Name:           s.drivers[driverId].Driver.Name,
		PackageSlug:    s.drivers[driverId].Driver.PackageSlug,
		ProfilePicture: s.drivers[driverId].Driver.ProfilePicture,
		CarPlate:       s.drivers[driverId].Driver.CarPlate,
	}

	s.mu.Lock()
	s.drivers[driverId] = &driverInMap{
		Driver: driver,
	}

	s.mu.Unlock()

	slog.Info("Driver location updated",
		"driver_id", driver.Id,
		"package_slug", driver.PackageSlug,
		"latitude", location.GetLatitude(),
		"longitude", location.GetLongitude(),
		"geohash", geohash,
	)

	return driver, nil
}

func (s *Service) HandleTripCreated(ctx context.Context, payload messaging.TripEventData) error {
	packageSlug := payload.Trip.GetSelectedFare().GetPackageSlug()

	suitableIDs := s.FindAvailableDrivers(
		packageSlug,
		payload.ExcludeDriverIDs,
	)

	slog.InfoContext(ctx, "Searching for suitable drivers",
		"package_slug", packageSlug,
		"found", len(suitableIDs),
		"excluded_drivers", len(payload.ExcludeDriverIDs),
	)

	if len(suitableIDs) == 0 {
		slog.WarnContext(ctx, "No suitable drivers found, notifying rider",
			"package_slug", packageSlug,
			"trip_id", payload.Trip.GetId(),
		)
		
		if err := s.publisher.PublishNoDriversFound(ctx, payload.Trip.UserID, nil); err != nil {
			slog.ErrorContext(ctx, "Failed to publish no-drivers-found event",
				"trip_id", payload.Trip.GetId(),
				"err", err,
			)
			return err
		}
		return nil
	}

	randIndex := math.IntN(len(suitableIDs))
	suitableDriverID := suitableIDs[randIndex]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal trip event",
			"trip_id", payload.Trip.GetId(),
			"err", err,
		)
		return err
	}

	s.mu.Lock()
	s.tripToDriver[payload.Trip.GetId()] = suitableDriverID
	s.mu.Unlock()

	// Ask drivers if they confirm the trip request
	if err := s.publisher.PublishTripRequest(ctx, suitableDriverID, marshalledEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to publish trip request to driver",
			"driver_id", suitableDriverID,
			"trip_id", payload.Trip.GetId(),
			"err", err,
		)
		return err
	}

	slog.InfoContext(ctx, "Trip request sent to driver",
		"driver_id", suitableDriverID,
		"trip_id", payload.Trip.GetId(),
		"package_slug", packageSlug,
	)

	return nil
}

func (s *Service) HandleTripCanceled(ctx context.Context, payload messaging.TripEventData) error {
	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal trip event payload", "err", err)
		return err
	}

	// 1. If a driver is already assigned, notify them specifically
	if payload.Trip.GetDriver() != nil && payload.Trip.GetDriver().GetId() != "" {
		return s.publisher.PublishTripCanceled(ctx, payload.Trip.GetDriver().GetId(), marshalledEvent)
	}

	s.mu.RLock()
	lastDriverID, exists := s.tripToDriver[payload.Trip.GetId()]
	s.mu.RUnlock()

	if exists {
		return s.publisher.PublishTripCanceled(ctx, lastDriverID, marshalledEvent)
	}

	// 2. Fallback to ExcludeDriverIDs if cache miss
	if len(payload.ExcludeDriverIDs) > 0 {
		lastDriverID := payload.ExcludeDriverIDs[len(payload.ExcludeDriverIDs)-1]
		return s.publisher.PublishTripCanceled(ctx, lastDriverID, marshalledEvent)
	}

	return nil
}
