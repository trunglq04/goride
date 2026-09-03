package domain

import (
	"context"

	tripTypes "github.com/trunglq04/goride/services/trip-service/pkg/types"
	pbd "github.com/trunglq04/goride/shared/proto/driver"
	types "github.com/trunglq04/goride/shared/types"
)

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	CancelTrip(ctx context.Context, userID, tripID, reason string) error
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate, useOSMApi bool) (*tripTypes.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*RideFareModel // price for each package slugs
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*RideFareModel, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver, excludedDriverID *string) error
}
