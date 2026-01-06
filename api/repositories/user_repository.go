// api/repositories/user_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Save(ctx context.Context, user *models.User) error
	AddLikedFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error)
	RemoveLikedFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error)
	GetLikedFoodIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(coll *mongo.Collection) UserRepository {
	return &userRepository{collection: coll}
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Save(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) AddLikedFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": userID}
	update := bson.M{"$addToSet": bson.M{"liked_food_ids": foodID}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *userRepository) RemoveLikedFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"liked_food_ids": foodID}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *userRepository) GetLikedFoodIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.LikedFoodIDs, nil
}
