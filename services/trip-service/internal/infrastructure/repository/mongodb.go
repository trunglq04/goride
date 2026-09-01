package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/db"
	pbd "github.com/trunglq04/goride/shared/proto/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRepository struct {
	db *mongo.Database
}

func NewMongoRepository(db *mongo.Database) *mongoRepository {
	return &mongoRepository{db: db}
}

// EnsureIndexes creates required indexes, including the TTL index on ride_fares.
// Call this once at application startup.
func (r *mongoRepository) EnsureIndexes(ctx context.Context) error {
	// TTL index: ride_fares documents expire 24 hours after createdAt
	ttlIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "createdAt", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(5 * 60).
			SetName("ride_fares_ttl"),
	}
	_, err := r.db.Collection(db.RideFaresCollection).Indexes().CreateOne(ctx, ttlIndex)
	return err
}

func (r *mongoRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	result, err := r.db.Collection(db.TripsCollection).InsertOne(ctx, trip)
	if err != nil {
		return nil, err
	}

	trip.ID = result.InsertedID.(primitive.ObjectID)

	return trip, nil
}

func (r *mongoRepository) CancelTrip(ctx context.Context, userID, tripID, reason string) error {
	_id, err := primitive.ObjectIDFromHex(tripID)
	if err != nil {
		return err // invalid ID
	}

	filter := bson.M{
		"_id":    _id,
		"userID": userID,
	}
	update := bson.M{
		"$set": bson.M{
			"status": "CANCELED",
		},
	}
	result, err := r.db.Collection(db.TripsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("trip not found or unauthorized to cancel")
	}
	return nil
}

func (r *mongoRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := r.db.Collection(db.TripsCollection).FindOne(ctx, bson.M{"_id": _id})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var trip domain.TripModel
	err = result.Decode(&trip)
	if err != nil {
		return nil, err
	}

	return &trip, nil
}

func (r *mongoRepository) UpdateTrip(ctx context.Context, tripID, status string, driver *pbd.Driver, excludedDriverID *string) error {
	_id, err := primitive.ObjectIDFromHex(tripID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	if driver != nil {
		update["$set"].(bson.M)["driver"] = driver
	}

	if excludedDriverID != nil {
		update["$addToSet"] = bson.M{
			"excludedDriverIDs": *excludedDriverID,
		}
	}

	result, err := r.db.Collection(db.TripsCollection).UpdateOne(ctx, bson.M{"_id": _id}, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("trip not found: %s", tripID)
	}

	return nil
}

func (r *mongoRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	fare.CreatedAt = time.Now().UTC()
	result, err := r.db.Collection(db.RideFaresCollection).InsertOne(ctx, fare)
	if err != nil {
		return err
	}

	fare.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (r *mongoRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := r.db.Collection(db.RideFaresCollection).FindOne(ctx, bson.M{"_id": _id})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var fare domain.RideFareModel
	err = result.Decode(&fare)
	if err != nil {
		return nil, err
	}

	return &fare, nil
}
