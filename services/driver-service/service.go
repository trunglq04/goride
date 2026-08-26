package main

import (
	"log/slog"
	math "math/rand/v2"

	"github.com/mmcloughlin/geohash"

	"sync"

	pb "github.com/trunglq04/goride/shared/proto/driver"
	"github.com/trunglq04/goride/shared/util"
)

type Service struct {
	drivers map[string]*driverInMap
	mu      sync.RWMutex
}

type driverInMap struct {
	Driver *pb.Driver
	// TODO: route
}

func NewService() *Service {
	return &Service{
		drivers: make(map[string]*driverInMap),
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
