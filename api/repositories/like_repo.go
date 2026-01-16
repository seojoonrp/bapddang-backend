// api/repositories/like_repo.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LikeRepository interface {
	Create(ctx context.Context, like *models.Like) error
	Delete(ctx context.Context, userID, foodID primitive.ObjectID) (int64, error)
	CheckLikedStatus(ctx context.Context, userID primitive.ObjectID, foodIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error)
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

func (r *likeRepository) Delete(ctx context.Context, userID, foodID primitive.ObjectID) (int64, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"user_id": userID,
		"food_id": foodID,
	})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (r *likeRepository) CheckLikedStatus(ctx context.Context, userID primitive.ObjectID, foodIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	filter := bson.M{
		"user_id": userID,
		"food_id": bson.M{"$in": foodIDs},
	}

	opts := options.Find().SetProjection(bson.M{"food_id": 1, "_id": 0})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	likedMap := make(map[primitive.ObjectID]bool)
	for cursor.Next(ctx) {
		var result struct {
			FoodID primitive.ObjectID `bson:"food_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		likedMap[result.FoodID] = true
	}

	return likedMap, nil
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
