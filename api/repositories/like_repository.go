// api/repositories/like_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LikeRepository interface {
	Create(ctx context.Context, like *models.Like) error
	Delete(ctx context.Context, userID, meetingID primitive.ObjectID) (int64, error)
	FindFoodIDsByUserID(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
}

type likeRepository struct {
	collection *mongo.Collection
}

func NewLikeRepository(db *mongo.Database) LikeRepository {
	return &likeRepository{
		collection: db.Collection("likes"),
	}
}

func (r *likeRepository) Create(ctx context.Context, like *models.Like) error {
	_, err := r.collection.InsertOne(ctx, like)
	return err
}

func (r *likeRepository) Delete(ctx context.Context, userID, meetingID primitive.ObjectID) (int64, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"user_id":    userID,
		"meeting_id": meetingID,
	})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (r *likeRepository) FindFoodIDsByUserID(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var foodIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var like models.Like
		if err := cursor.Decode(&like); err != nil {
			continue
		}
		foodIDs = append(foodIDs, like.FoodID)
	}

	return foodIDs, nil
}
