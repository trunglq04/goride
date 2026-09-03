package domain

import (
	"context"

	pb "github.com/trunglq04/goride/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	UserID            string             `bson:"userID"`
	Status            string             `bson:"status"`
	RideFare          *RideFareModel     `bson:"rideFare"`
	Driver            *pb.TripDriver     `bson:"driver"`
	ExcludedDriverIDs []string           `bson:"excludedDriverIDs"`
}

func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id:           t.ID.Hex(),
		UserID:       t.UserID,
		SelectedFare: t.RideFare.ToProto(),
		Status:       t.Status,
		Driver:       t.Driver,
		Route:        t.RideFare.Route.ToProto(),
	}
}

type TripEventPublisher interface {
	PublishTripCreated(ctx context.Context, trip *TripModel) error
	PublishTripCanceled(ctx context.Context, trip *TripModel) error
}
