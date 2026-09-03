package domain

import (
	"context"

	pbd "github.com/trunglq04/goride/shared/proto/driver"
)

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	CancelTrip(ctx context.Context, userID, tripID, reason string) error
	SaveRideFare(ctx context.Context, fare *RideFareModel) error
	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID, status string, driver *pbd.Driver, excludedDriverID *string) error
}
