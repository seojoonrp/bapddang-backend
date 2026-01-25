// api/repositories/marshmallow.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MarshmallowRepository interface {
	Create(ctx context.Context, marshmallow models.Marshmallow) error
	AddReviewData(ctx context.Context, marshmallowID primitive.ObjectID, rating int) error
	UpdateReviewData(ctx context.Context, marshmallowID primitive.ObjectID, oldRating int, newRating int) error
	CompleteMarshmallow(ctx context.Context, marshmallowID primitive.ObjectID, status int) error
	FindByUserIDAndWeek(ctx context.Context, userID primitive.ObjectID, week int) (*models.Marshmallow, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Marshmallow, error)
}

type marshmallowRepository struct {
	collection *mongo.Collection
}

func NewMarshmallowRepository(db *mongo.Database) MarshmallowRepository {
	return &marshmallowRepository{
		collection: db.Collection("marshmallows"),
	}
}

func (r *marshmallowRepository) Create(ctx context.Context, marshmallow models.Marshmallow) error {
	_, err := r.collection.InsertOne(ctx, marshmallow)
	return err
}

func (r *marshmallowRepository) AddReviewData(ctx context.Context, marshmallowID primitive.ObjectID, rating int) error {
	filter := primitive.M{"_id": marshmallowID}
	update := primitive.M{
		"$inc": primitive.M{
			"review_count": 1,
			"total_rating": rating,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *marshmallowRepository) UpdateReviewData(ctx context.Context, marshmallowID primitive.ObjectID, oldRating int, newRating int) error {
	filter := primitive.M{"_id": marshmallowID}
	update := primitive.M{
		"$inc": primitive.M{
			"total_rating": newRating - oldRating,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *marshmallowRepository) CompleteMarshmallow(ctx context.Context, marshmallowID primitive.ObjectID, status int) error {
	filter := primitive.M{"_id": marshmallowID}
	update := primitive.M{
		"$set": primitive.M{
			"status":      status,
			"is_complete": true,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *marshmallowRepository) FindByUserIDAndWeek(ctx context.Context, userID primitive.ObjectID, week int) (*models.Marshmallow, error) {
	filter := primitive.M{
		"user_id": userID,
		"week":    week,
	}

	var marshmallow models.Marshmallow
	err := r.collection.FindOne(ctx, filter).Decode(&marshmallow)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &marshmallow, nil
}

func (r *marshmallowRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Marshmallow, error) {
	marshmallows := make([]models.Marshmallow, 0)

	filter := primitive.M{"user_id": userID}

	findOptions := options.Find().SetSort(primitive.M{"week": 1})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &marshmallows); err != nil {
		return nil, err
	}

	return marshmallows, nil
}
