package domain

import (
	tripTypes "github.com/trunglq04/goride/services/trip-service/pkg/types"
	pb "github.com/trunglq04/goride/shared/proto/trip"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string // ex: van, luxury, sedan
	TotalPriceInCents float64
	Route             *tripTypes.OsrmApiResponse
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
