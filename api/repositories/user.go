// api/repositories/user.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	FindByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, userID primitive.ObjectID) error
	UpdateAppleRefreshToken(ctx context.Context, userID primitive.ObjectID, refreshToken string) error
	UpdateDayAndWeek(ctx context.Context, userID primitive.ObjectID, newDay int, newWeek int) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{collection: db.Collection("users")}
}

func (r *userRepository) FindByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
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

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) Delete(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": userID})
	return err
}

func (r *userRepository) UpdateAppleRefreshToken(ctx context.Context, userID primitive.ObjectID, refreshToken string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"apple_refresh_token": refreshToken}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) UpdateDayAndWeek(ctx context.Context, userID primitive.ObjectID, newDay int, newWeek int) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"day": newDay, "week": newWeek}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
