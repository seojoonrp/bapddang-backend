// api/services/review.go

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
	Create(ctx context.Context, input models.CreateReviewRequest, userID string) (*models.Review, error)
	Update(ctx context.Context, reviewID string, userID string, input models.UpdateReviewRequest) (*models.Review, error)
	Delete(ctx context.Context, reviewID string, userID string) error
	GetMyReviewsByDay(ctx context.Context, userID string, day int) ([]models.Review, error)
	GetRecentWithStandardFood(ctx context.Context, userID string, count int) ([]models.RecentReviewResponse, error)
}

type reviewService struct {
	reviewRepo      repositories.ReviewRepository
	foodRepo        repositories.FoodRepository
	userRepo        repositories.UserRepository
	marshmallowRepo repositories.MarshmallowRepository
}

func NewReviewService(rr repositories.ReviewRepository, fr repositories.FoodRepository, ur repositories.UserRepository, mr repositories.MarshmallowRepository) ReviewService {
	return &reviewService{
		reviewRepo:      rr,
		foodRepo:        fr,
		userRepo:        ur,
		marshmallowRepo: mr,
	}
}

func (s *reviewService) Create(ctx context.Context, req models.CreateReviewRequest, userID string) (*models.Review, error) {
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

	commentLen := len([]rune(req.Comment))
	if commentLen > 50 || commentLen <= 0 {
		return nil, apperr.BadRequest("comment must be between 1 and 50 characters", nil)
	}

	user, err := s.userRepo.FindByID(ctx, uID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch user", err)
	}

	newReview := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    uID,
		Name:      req.Name,
		Foods:     req.Foods,
		MealTime:  req.MealTime,
		ImageURL:  req.ImageURL,
		Comment:   req.Comment,
		Rating:    req.Rating,
		Day:       user.Day,
		Week:      user.Week,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.reviewRepo.Create(ctx, &newReview)
	if err != nil {
		return nil, apperr.InternalServerError("failed to save review", err)
	}

	var standardFoodIDs []primitive.ObjectID
	var customFoodIDs []primitive.ObjectID

	for _, foodItem := range newReview.Foods {
		foodID, err := primitive.ObjectIDFromHex(foodItem.FoodID)
		if err != nil {
			continue
		}
		if foodItem.Type == models.FoodTypeStandard {
			standardFoodIDs = append(standardFoodIDs, foodID)
		}
		if foodItem.Type == models.FoodTypeCustom {
			customFoodIDs = append(customFoodIDs, foodID)
		}
	}

	if len(standardFoodIDs) > 0 {
		err = s.foodRepo.UpdateStandardCreatedReviewStats(ctx, standardFoodIDs, newReview.Rating)
		if err != nil {
			return nil, apperr.InternalServerError("failed to update food review stats", err)
		}
	}
	if len(customFoodIDs) > 0 {
		err = s.foodRepo.UpdateCustomCreatedReviewStats(ctx, customFoodIDs)
		if err != nil {
			return nil, apperr.InternalServerError("failed to update custom food review stats", err)
		}
	}

	marshmallow, err := s.marshmallowRepo.FindByUserIDAndWeek(ctx, user.ID, user.Week)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch marshmallow", err)
	}
	if marshmallow == nil {
		return nil, apperr.InternalServerError("marshmallow not found for current week", nil)
	}
	err = s.marshmallowRepo.AddReviewData(ctx, marshmallow.ID, newReview.Rating)
	if err != nil {
		return nil, apperr.InternalServerError("failed to update marshmallow status", err)
	}

	return &newReview, nil
}

func (s *reviewService) Update(ctx context.Context, reviewID string, userID string, req models.UpdateReviewRequest) (*models.Review, error) {
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

	if req.Rating <= 0 || req.Rating > 5 {
		return nil, apperr.BadRequest("rating must be in between 1 and 5", nil)
	}

	commentLen := len([]rune(req.Comment))
	if commentLen > 50 || commentLen <= 0 {
		return nil, apperr.BadRequest("comment must be between 1 and 50 characters", nil)
	}

	user, err := s.userRepo.FindByID(ctx, review.UserID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch user", err)
	}

	if review.Week != user.Week {
		return nil, apperr.BadRequest("cannot update review from previous weeks", nil)
	}

	oldRating := review.Rating

	review.MealTime = req.MealTime
	review.ImageURL = req.ImageURL
	review.Comment = req.Comment
	review.Rating = req.Rating
	review.UpdatedAt = time.Now()

	err = s.reviewRepo.Update(ctx, review)
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
		err = s.foodRepo.UpdateStandardModifiedReviewStats(ctx, standardFoodIDs, oldRating, review.Rating)
		if err != nil {
			return nil, apperr.InternalServerError("failed to update food review stats", err)
		}
	}

	if oldRating != review.Rating {
		marshmallow, err := s.marshmallowRepo.FindByUserIDAndWeek(ctx, user.ID, user.Week)
		if err != nil {
			return nil, apperr.InternalServerError("failed to fetch marshmallow", err)
		}
		if marshmallow == nil {
			return nil, apperr.InternalServerError("marshmallow not found for current week", nil)
		}
		err = s.marshmallowRepo.UpdateReviewData(ctx, marshmallow.ID, oldRating, review.Rating)
		if err != nil {
			return nil, apperr.InternalServerError("failed to update marshmallow status", err)
		}
	}

	return review, nil
}

func (s *reviewService) Delete(ctx context.Context, reviewID string, userID string) error {
	rID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		return apperr.BadRequest("invalid review ID format", err)
	}

	review, err := s.reviewRepo.FindByID(ctx, rID)
	if err != nil {
		return apperr.InternalServerError("failed to fetch review", err)
	}

	if review.UserID.Hex() != userID {
		return apperr.Unauthorized("you are not the owner of this review", nil)
	}

	var standardFoodIDs []primitive.ObjectID
	var customFoodIDs []primitive.ObjectID

	for _, foodItem := range review.Foods {
		foodID, err := primitive.ObjectIDFromHex(foodItem.FoodID)
		if err != nil {
			continue
		}
		if foodItem.Type == models.FoodTypeStandard {
			standardFoodIDs = append(standardFoodIDs, foodID)
		}
		if foodItem.Type == models.FoodTypeCustom {
			customFoodIDs = append(customFoodIDs, foodID)
		}
	}

	if len(standardFoodIDs) > 0 {
		err = s.foodRepo.UpdateStandardDeletedReviewStats(ctx, standardFoodIDs, review.Rating)
		if err != nil {
			return apperr.InternalServerError("failed to update food review stats", err)
		}
	}
	if len(customFoodIDs) > 0 {
		err = s.foodRepo.UpdateCustomDeletedReviewStats(ctx, customFoodIDs)
		if err != nil {
			return apperr.InternalServerError("failed to update custom food review stats", err)
		}
	}

	err = s.reviewRepo.Delete(ctx, rID)
	if err != nil {
		return apperr.InternalServerError("failed to delete review", err)
	}

	return nil
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

func (s *reviewService) GetRecentWithStandardFood(ctx context.Context, userID string, count int) ([]models.RecentReviewResponse, error) {
	if count <= 0 {
		return nil, apperr.BadRequest("count must be a positive integer", nil)
	}
	if count > 3 {
		count = 3
	}

	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	reviews, err := s.reviewRepo.FindRecentWithStandardFood(ctx, uID, int64(count))
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch recent reviews", err)
	}

	result := make([]models.RecentReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		var firstStandardFood models.StandardFood

		for _, f := range r.Foods {
			if f.Type == models.FoodTypeStandard {
				fID, err := primitive.ObjectIDFromHex(f.FoodID)
				if err != nil {
					return nil, apperr.InternalServerError("invalid food ID in review", err)
				}
				food, err := s.foodRepo.FindStandardByID(ctx, fID)
				if err != nil {
					return nil, apperr.InternalServerError("failed to fetch food details", err)
				}
				if food == nil {
					return nil, apperr.InternalServerError("food not found for ID in review", nil)
				}

				firstStandardFood = *food
				break
			}
		}

		result = append(result, models.RecentReviewResponse{
			Comment:   r.Comment,
			Rating:    r.Rating,
			CreatedAt: r.CreatedAt,
			Food:      firstStandardFood,
		})
	}

	return result, nil
}
