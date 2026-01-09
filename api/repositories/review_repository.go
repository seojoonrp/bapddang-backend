// api/repositories/review_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepository interface {
	SaveReview(ctx context.Context, review *models.Review) error
	UpdateReview(ctx context.Context, review *models.Review) error
	FindByUserIDAndDay(ctx context.Context, userID primitive.ObjectID, day int) ([]models.Review, error)
	FindByIDAndUserID(ctx context.Context, reviewID, userID primitive.ObjectID) (*models.Review, error)
}

type reviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewRepository{collection: db.Collection("reviews")}
}

func (r *reviewRepository) SaveReview(ctx context.Context, review *models.Review) error {
	_, err := r.collection.InsertOne(ctx, review)
	return err
}

func (r *reviewRepository) UpdateReview(ctx context.Context, review *models.Review) error {
	filter := bson.M{"_id": review.ID}
	update := bson.M{
		"$set": bson.M{
			"meal_time":  review.MealTime,
			"tags":       review.Tags,
			"image_url":  review.ImageURL,
			"comment":    review.Comment,
			"rating":     review.Rating,
			"updated_at": review.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
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

func (r *reviewRepository) FindByIDAndUserID(ctx context.Context, reviewID, userID primitive.ObjectID) (*models.Review, error) {
	var review models.Review

	filter := bson.M{"_id": reviewID, "user_id": userID}
	err := r.collection.FindOne(ctx, filter).Decode(&review)
	if err != nil {
		return nil, err
	}

	return &review, nil
}
