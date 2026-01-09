// api/repositories/food_repository.go

package repositories

import (
	"context"
	"errors"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodRepository interface {
	FindStandardByID(ctx context.Context, id primitive.ObjectID) (*models.StandardFood, error)
	FindStandardByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.StandardFood, error)
	FindStandardByName(ctx context.Context, name string) (*models.StandardFood, error)
	FindCustomByName(ctx context.Context, name string) (*models.CustomFood, error)

	CreateStandard(ctx context.Context, food *models.StandardFood) error
	CreateCustom(ctx context.Context, food *models.CustomFood) error
	AddUserToCustom(ctx context.Context, foodID, userID primitive.ObjectID) (bool, error)

	GetRandomStandard(ctx context.Context, speed string, count int) ([]*models.StandardFood, error)

	UpdateCreatedReviewStats(ctx context.Context, foodID []primitive.ObjectID, rating int) error
	UpdateModifiedReviewStats(ctx context.Context, foodID []primitive.ObjectID, oldRating, newRating int) error

	IncrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error
	DecrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error
}

type foodRepository struct {
	standardFoodCollection *mongo.Collection
	customFoodCollection   *mongo.Collection
}

func NewFoodRepository(db *mongo.Database) FoodRepository {
	return &foodRepository{
		standardFoodCollection: db.Collection("standard_foods"),
		customFoodCollection:   db.Collection("custom_foods"),
	}
}

func (r *foodRepository) FindStandardByID(ctx context.Context, id primitive.ObjectID) (*models.StandardFood, error) {
	var food models.StandardFood
	err := r.standardFoodCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&food)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) FindStandardByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.StandardFood, error) {
	if len(ids) == 0 {
		return []*models.StandardFood{}, nil
	}

	cursor, err := r.standardFoodCollection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var foods []*models.StandardFood
	if err := cursor.All(ctx, &foods); err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *foodRepository) FindStandardByName(ctx context.Context, name string) (*models.StandardFood, error) {
	var food models.StandardFood
	err := r.standardFoodCollection.FindOne(ctx, bson.M{"name": name}).Decode(&food)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) FindCustomByName(ctx context.Context, name string) (*models.CustomFood, error) {
	var food models.CustomFood
	err := r.customFoodCollection.FindOne(ctx, bson.M{"name": name}).Decode(&food)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) CreateStandard(ctx context.Context, food *models.StandardFood) error {
	_, err := r.standardFoodCollection.InsertOne(ctx, food)
	return err
}

func (r *foodRepository) CreateCustom(ctx context.Context, food *models.CustomFood) error {
	_, err := r.customFoodCollection.InsertOne(ctx, food)
	return err
}

// 이미 있으면 true, 없으면 false
func (r *foodRepository) AddUserToCustom(ctx context.Context, foodID, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": foodID}
	update := bson.M{"$addToSet": bson.M{"using_user_ids": userID}}
	result, err := r.customFoodCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	if result.MatchedCount == 0 {
		return false, errors.New("custom food not found")
	}
	if result.ModifiedCount == 0 {
		return true, nil
	}
	return false, nil
}

func (r *foodRepository) GetRandomStandard(ctx context.Context, speed string, count int) ([]*models.StandardFood, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"speed": speed}}},
		{{Key: "$sample", Value: bson.M{"size": count}}},
	}

	cursor, err := r.standardFoodCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var foods []*models.StandardFood
	for cursor.Next(ctx) {
		var food models.StandardFood
		if err := cursor.Decode(&food); err != nil {
			return nil, err
		}
		foods = append(foods, &food)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *foodRepository) UpdateCreatedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, rating int) error {
	filter := bson.M{"_id": bson.M{"$in": foodIDs}}

	incMap := bson.M{"review_count": 1}
	incMap["total_rating"] = rating

	update := bson.M{"$inc": incMap}
	_, err := r.standardFoodCollection.UpdateMany(ctx, filter, update)
	return err
}

func (r *foodRepository) UpdateModifiedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, oldRating, newRating int) error {
	filter := bson.M{"_id": bson.M{"$in": foodIDs}}

	ratingDiff := newRating - oldRating

	incMap := bson.M{}
	if ratingDiff != 0 {
		incMap["total_rating"] = ratingDiff
	} else {
		return nil
	}

	update := bson.M{"$inc": incMap}
	_, err := r.standardFoodCollection.UpdateMany(ctx, filter, update)
	return err
}

func (r *foodRepository) IncrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error {
	result, err := r.standardFoodCollection.UpdateOne(
		ctx,
		bson.M{"_id": foodID},
		bson.M{"$inc": bson.M{"like_count": 1}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("food not found")
	}
	return nil
}

func (r *foodRepository) DecrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error {
	result, err := r.standardFoodCollection.UpdateOne(
		ctx,
		bson.M{"_id": foodID, "like_count": bson.M{"$gt": 0}},
		bson.M{"$inc": bson.M{"like_count": -1}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("food not found or like count is already zero")
	}
	return nil
}
