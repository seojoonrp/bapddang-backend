// api/repositories/review.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
	Update(ctx context.Context, review *models.Review) error
	Delete(ctx context.Context, reviewID primitive.ObjectID) error
	FindByUserIDAndDay(ctx context.Context, userID primitive.ObjectID, day int) ([]models.Review, error)
	FindByID(ctx context.Context, reviewID primitive.ObjectID) (*models.Review, error)
}

type reviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewRepository{collection: db.Collection("reviews")}
}

func (r *reviewRepository) Create(ctx context.Context, review *models.Review) error {
	_, err := r.collection.InsertOne(ctx, review)
	return err
}

func (r *reviewRepository) Update(ctx context.Context, review *models.Review) error {
	filter := bson.M{"_id": review.ID}
	update := bson.M{
		"$set": bson.M{
			"meal_time":  review.MealTime,
			"image_url":  review.ImageURL,
			"comment":    review.Comment,
			"rating":     review.Rating,
			"updated_at": review.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *reviewRepository) Delete(ctx context.Context, reviewID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": reviewID})
	return err
}

func (r *reviewRepository) FindByUserIDAndDay(ctx context.Context, userID primitive.ObjectID, day int) ([]models.Review, error) {
	var reviews []models.Review

	filter := bson.M{"user_id": userID, "day": day}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *reviewRepository) FindByID(ctx context.Context, reviewID primitive.ObjectID) (*models.Review, error) {
	var review models.Review
	err := r.collection.FindOne(ctx, bson.M{"_id": reviewID}).Decode(&review)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &review, nil
}
