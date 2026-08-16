package main

import (
	"log"
	math "math/rand/v2"

	"github.com/mmcloughlin/geohash"

	"sync"

	pb "github.com/trunglq04/goride/shared/proto/driver"
	"github.com/trunglq04/goride/shared/util"
)

type Service struct {
	drivers []*driverInMap
	mu      sync.RWMutex
}

type driverInMap struct {
	Driver *pb.Driver
	// Index int
	// TODO: route
}

func NewService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
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
	s.drivers = append(s.drivers, &driverInMap{
		Driver: driver,
	})
	s.mu.Unlock()

	log.Printf("Succesfully REGISTER a DRIVER: %v, packageSlug: %v\n", driver.Id, driver.PackageSlug)

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driver := range s.drivers {
		if driver.Driver.Id == driverId {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
			log.Println("Succesfully UNREGISTER a DRIVER:", driverId)
		}
	}

}
