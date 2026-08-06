package repository

import (
	"context"
	"sync"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
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
