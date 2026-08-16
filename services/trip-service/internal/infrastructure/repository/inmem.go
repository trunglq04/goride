package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	pbd "github.com/trunglq04/goride/shared/proto/driver"
	pb "github.com/trunglq04/goride/shared/proto/trip"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
	mu        sync.RWMutex
}

func NewInmemReposity() *inmemRepository {
	return &inmemRepository{
		trips:     map[string]*domain.TripModel{},
		rideFares: map[string]*domain.RideFareModel{},
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	r.rideFares[f.ID.Hex()] = f

	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	f, exist := r.rideFares[id]
	if !exist {
		return nil, fmt.Errorf("fare does not exist with ID: %s", id)
	}

	return f, nil
}

func (r *inmemRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	trip, ok := r.trips[id]
	if !ok {
		return nil, nil
	}
	return trip, nil
}

func (r *inmemRepository) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver, excludedDriverID *string) error {
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found with ID: %s", tripID)
	}

	trip.Status = status

	if driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			CarPlate:       driver.CarPlate,
			ProfilePicture: driver.ProfilePicture,
		}
	}

	if excludedDriverID != nil {
		trip.ExcludedDriverIDs = append(trip.ExcludedDriverIDs, *excludedDriverID)
	}
	return nil
}
