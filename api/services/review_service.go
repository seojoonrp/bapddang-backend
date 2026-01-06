// api/services/review_service.go

package services

import (
	"context"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewService interface {
	CreateReview(ctx context.Context, input models.ReviewInput, user models.User) (*models.Review, error)
	UpdateReview(ctx context.Context, reviewID primitive.ObjectID, input models.ReviewInput, user models.User) (*models.Review, int, error)
	GetMyReviewsByDay(ctx context.Context, userID primitive.ObjectID, day int) ([]models.Review, error)
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
	foodRepo   repositories.FoodRepository
}

func NewReviewService(reviewRepo repositories.ReviewRepository, foodRepo repositories.FoodRepository) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
		foodRepo:   foodRepo,
	}
}

func (s *reviewService) CreateReview(ctx context.Context, input models.ReviewInput, user models.User) (*models.Review, error) {
	newReview := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		Name:      input.Name,
		Foods:     input.Foods,
		Speed:     input.Speed,
		MealTime:  input.MealTime,
		Tags:      input.Tags,
		ImageURL:  input.ImageURL,
		Comment:   input.Comment,
		Rating:    input.Rating,
		Day:       user.Day,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.reviewRepo.SaveReview(ctx, &newReview)
	if err != nil {
		return nil, err
	}

	return &newReview, nil
}

func (s *reviewService) UpdateReview(ctx context.Context, reviewID primitive.ObjectID, input models.ReviewInput, user models.User) (*models.Review, int, error) {
	existingReview, err := s.reviewRepo.FindByIDAndUserID(ctx, reviewID, user.ID)
	if err != nil {
		return nil, 0, err
	}

	oldRating := existingReview.Rating

	existingReview.MealTime = input.MealTime
	existingReview.Tags = input.Tags
	existingReview.ImageURL = input.ImageURL
	existingReview.Comment = input.Comment
	existingReview.Rating = input.Rating
	existingReview.UpdatedAt = time.Now()

	err = s.reviewRepo.UpdateReview(ctx, existingReview)
	if err != nil {
		return nil, 0, err
	}

	return existingReview, oldRating, nil
}

func (s *reviewService) GetMyReviewsByDay(ctx context.Context, userID primitive.ObjectID, day int) ([]models.Review, error) {
	return s.reviewRepo.FindByUserIDAndDay(ctx, userID, day)
}
