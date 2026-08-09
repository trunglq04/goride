package main

import pb "github.com/trunglq04/goride/shared/proto/driver"

type Service struct {
	drivers []*driverInMap
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

// TODO: Register and unregister methods