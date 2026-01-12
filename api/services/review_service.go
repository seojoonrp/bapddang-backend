// api/services/review_service.go

package services

import (
	"context"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewService interface {
	CreateReview(ctx context.Context, input models.CreateReviewRequest, userID string) (*models.Review, error)
	UpdateReview(ctx context.Context, reviewID string, userID string, input models.UpdateReviewRequest) (*models.Review, error)
	GetMyReviewsByDay(ctx context.Context, userID string, day int) ([]models.Review, error)
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

func (s *reviewService) CreateReview(ctx context.Context, req models.CreateReviewRequest, userID string) (*models.Review, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	if len(req.Foods) == 0 {
		return nil, apperr.BadRequest("at least one food item is required", nil)
	}

	if req.Rating <= 0 || req.Rating > 5 {
		return nil, apperr.BadRequest("rating must be between 1 and 5", nil)
	}

	newReview := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    uID,
		Name:      req.Name,
		Foods:     req.Foods,
		Speed:     req.Speed,
		MealTime:  req.MealTime,
		Tags:      req.Tags,
		ImageURL:  req.ImageURL,
		Comment:   req.Comment,
		Rating:    req.Rating,
		Day:       1, // TODO
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.reviewRepo.CreateReview(ctx, &newReview)
	if err != nil {
		return nil, apperr.InternalServerError("failed to save review", err)
	}

	var standardFoodIDs []primitive.ObjectID
	for _, foodItem := range newReview.Foods {
		if foodItem.Type == models.FoodTypeStandard {
			foodID, err := primitive.ObjectIDFromHex(foodItem.FoodID)
			if err != nil {
				continue
			}
			standardFoodIDs = append(standardFoodIDs, foodID)
		}
	}

	if len(standardFoodIDs) > 0 {
		err = s.foodRepo.UpdateCreatedReviewStats(ctx, standardFoodIDs, newReview.Rating)
		if err != nil {
			return nil, apperr.InternalServerError("failed to update food review stats", err)
		}
	}

	return &newReview, nil
}

func (s *reviewService) UpdateReview(ctx context.Context, reviewID string, userID string, req models.UpdateReviewRequest) (*models.Review, error) {
	rID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		return nil, apperr.BadRequest("invalid review ID format", err)
	}

	review, err := s.reviewRepo.FindByID(ctx, rID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch review", err)
	}

	if review.UserID.Hex() != userID {
		return nil, apperr.Unauthorized("you are not the owner of this review", nil)
	}

	if req.Rating != nil {
		if *req.Rating <= 0 || *req.Rating > 5 {
			return nil, apperr.BadRequest("rating must be between 1 and 5", nil)
		}
	}

	oldRating := review.Rating

	review.MealTime = req.MealTime
	if req.Tags != nil {
		review.Tags = *req.Tags
	}
	if req.ImageURL != nil {
		review.ImageURL = *req.ImageURL
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}
	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	review.UpdatedAt = time.Now()

	err = s.reviewRepo.UpdateReview(ctx, review)
	if err != nil {
		return nil, apperr.InternalServerError("failed to update review", err)
	}

	var standardFoodIDs []primitive.ObjectID
	for _, foodItem := range review.Foods {
		if foodItem.Type == models.FoodTypeStandard {
			foodID, err := primitive.ObjectIDFromHex(foodItem.FoodID)
			if err != nil {
				continue
			}
			standardFoodIDs = append(standardFoodIDs, foodID)
		}
	}

	if len(standardFoodIDs) > 0 && oldRating != review.Rating {
		err = s.foodRepo.UpdateModifiedReviewStats(ctx, standardFoodIDs, oldRating, review.Rating)
		if err != nil {
			return nil, apperr.InternalServerError("failed to update food review stats", err)
		}
	}

	return review, nil
}

func (s *reviewService) GetMyReviewsByDay(ctx context.Context, userID string, day int) ([]models.Review, error) {
	if day <= 0 {
		return nil, apperr.BadRequest("day must be a positive integer", nil)
	}

	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	reviews, err := s.reviewRepo.FindByUserIDAndDay(ctx, uID, day)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch reviews", err)
	}

	return reviews, nil
}
