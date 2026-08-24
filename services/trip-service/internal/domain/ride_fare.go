package domain

import (
	"time"

	tripTypes "github.com/trunglq04/goride/services/trip-service/pkg/types"
	pb "github.com/trunglq04/goride/shared/proto/trip"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID         `bson:"_id,omitempty"`
	UserID            string                     `bson:"userID"`
	PackageSlug       string                     `bson:"packageSlug"` // ex: sedan, suv, van, luxury
	TotalPriceInCents float64                    `bson:"totalPriceInCents"`
	Route             *tripTypes.OsrmApiResponse `bson:"route"`
	CreatedAt         time.Time                  `bson:"createdAt"`
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	var pbFares []*pb.RideFare
	for _, f := range fares {
		pbFares = append(pbFares, f.ToProto())
	}

	return pbFares
}
