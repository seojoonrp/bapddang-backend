// api/repositories/food_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodRepository interface {
	FindStandardFoodByID(ctx context.Context, id primitive.ObjectID) (*models.StandardFood, error)
	FindStandardFoodByName(ctx context.Context, name string) (*models.StandardFood, error)
	FindCustomFoodByName(ctx context.Context, name string) (*models.CustomFood, error)
	GetAllStandardFoods(ctx context.Context) ([]*models.StandardFood, error)
	GetAllCustomFoods(ctx context.Context) ([]*models.CustomFood, error)
	SaveCustomFood(ctx context.Context, food *models.CustomFood) error
	SaveStandardFood(ctx context.Context, food *models.StandardFood) error

	AddUserToCustomFood(ctx context.Context, foodID, userID primitive.ObjectID) error
	UpdateCreatedReviewStats(ctx context.Context, foodID []primitive.ObjectID, rating int) error
	UpdateModifiedReviewStats(ctx context.Context, foodID []primitive.ObjectID, oldRating, newRating int) error
	IncrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error
	DecrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error
}

type foodRepository struct {
	standardFoodCollection *mongo.Collection
	customFoodCollection   *mongo.Collection
}

func NewFoodRepository(standardColl *mongo.Collection, customColl *mongo.Collection) FoodRepository {
	return &foodRepository{
		standardFoodCollection: standardColl,
		customFoodCollection:   customColl,
	}
}

func (r *foodRepository) FindStandardFoodByID(ctx context.Context, id primitive.ObjectID) (*models.StandardFood, error) {
	var food models.StandardFood
	err := r.standardFoodCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) FindStandardFoodByName(ctx context.Context, name string) (*models.StandardFood, error) {
	var food models.StandardFood
	err := r.standardFoodCollection.FindOne(ctx, bson.M{"name": name}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) FindCustomFoodByName(ctx context.Context, name string) (*models.CustomFood, error) {
	var food models.CustomFood
	err := r.customFoodCollection.FindOne(ctx, bson.M{"name": name}).Decode(&food)
	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) GetAllStandardFoods(ctx context.Context) ([]*models.StandardFood, error) {
	var foods []*models.StandardFood

	filter := bson.M{}
	cursor, err := r.standardFoodCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &foods); err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *foodRepository) GetAllCustomFoods(ctx context.Context) ([]*models.CustomFood, error) {
	var foods []*models.CustomFood

	filter := bson.M{}
	cursor, err := r.customFoodCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &foods); err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *foodRepository) SaveStandardFood(ctx context.Context, food *models.StandardFood) error {
	_, err := r.standardFoodCollection.InsertOne(ctx, food)
	return err
}

func (r *foodRepository) SaveCustomFood(ctx context.Context, food *models.CustomFood) error {
	_, err := r.customFoodCollection.InsertOne(ctx, food)
	return err
}

func (r *foodRepository) AddUserToCustomFood(ctx context.Context, foodID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": foodID}
	update := bson.M{"$addToSet": bson.M{"using_user_ids": userID}}
	_, err := r.customFoodCollection.UpdateOne(ctx, filter, update)
	return err
}

func (r *foodRepository) UpdateCreatedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, rating int) error {
	if len(foodIDs) == 0 {
		return nil
	}
	if rating <= 0 || rating > 5 {
		return nil
	}

	filter := bson.M{"_id": bson.M{"$in": foodIDs}}

	incMap := bson.M{"review_count": 1}
	incMap["total_rating"] = rating

	update := bson.M{"$inc": incMap}
	_, err := r.standardFoodCollection.UpdateMany(ctx, filter, update)
	return err
}

func (r *foodRepository) UpdateModifiedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, oldRating, newRating int) error {
	if len(foodIDs) == 0 {
		return nil
	}
	if newRating <= 0 || newRating > 5 {
		return nil
	}

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
	filter := bson.M{"_id": foodID}
	update := bson.M{"$inc": bson.M{"like_count": 1}}
	_, err := r.standardFoodCollection.UpdateOne(ctx, filter, update)
	return err
}

func (r *foodRepository) DecrementLikeCount(ctx context.Context, foodID primitive.ObjectID) error {
	filter := bson.M{"_id": foodID}
	update := bson.M{"$inc": bson.M{"like_count": -1}}
	_, err := r.standardFoodCollection.UpdateOne(ctx, filter, update)
	return err
}
