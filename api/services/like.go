// api/services/like.go

package services

import (
	"context"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LikeService interface {
	LikeFood(ctx context.Context, userID, foodID string) error
	UnlikeFood(ctx context.Context, userID, foodID string) error
	GetLikedFoods(ctx context.Context, userID string) ([]*models.LikedFoodResponse, error)
}

type likeService struct {
	likeRepo repositories.LikeRepository
	foodRepo repositories.FoodRepository
}

func NewLikeService(likeRepo repositories.LikeRepository, foodRepo repositories.FoodRepository) LikeService {
	return &likeService{
		likeRepo: likeRepo,
		foodRepo: foodRepo,
	}
}

func (s *likeService) LikeFood(ctx context.Context, userID, foodID string) error {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return apperr.InternalServerError("invalid user ID in token", err)
	}

	fID, err := primitive.ObjectIDFromHex(foodID)
	if err != nil {
		return apperr.BadRequest("invalid food ID format", err)
	}

	err = s.likeRepo.Create(ctx, &models.Like{
		UserID:    uID,
		FoodID:    fID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperr.Conflict("food already liked", err)
		}
		return apperr.InternalServerError("failed to like food", err)
	}

	err = s.foodRepo.IncrementLikeCount(ctx, fID)
	if err != nil {
		return apperr.InternalServerError("failed to increment like count", err)
	}

	return nil
}

func (s *likeService) UnlikeFood(ctx context.Context, userID, foodID string) error {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return apperr.InternalServerError("invalid user ID in token", err)
	}

	fID, err := primitive.ObjectIDFromHex(foodID)
	if err != nil {
		return apperr.BadRequest("invalid food ID format", err)
	}

	deletedCount, err := s.likeRepo.Delete(ctx, uID, fID)
	if err != nil {
		return apperr.InternalServerError("failed to delete like", err)
	}
	if deletedCount == 0 {
		return apperr.NotFound("like not found", nil)
	}

	err = s.foodRepo.DecrementLikeCount(ctx, fID)
	if err != nil {
		return apperr.InternalServerError("failed to decrement like count", err)
	}

	return nil
}

func (s *likeService) GetLikedFoods(ctx context.Context, userID string) ([]*models.LikedFoodResponse, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	likes, err := s.likeRepo.FindLikesByUserID(ctx, uID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch liked food IDs", err)
	}

	if len(likes) == 0 {
		return []*models.LikedFoodResponse{}, nil
	}

	foodIDs := make([]primitive.ObjectID, 0, len(likes))
	for _, like := range likes {
		foodIDs = append(foodIDs, like.FoodID)
	}

	foods, err := s.foodRepo.FindStandardByIDs(ctx, foodIDs)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch liked foods", err)
	}

	foodMap := make(map[primitive.ObjectID]models.StandardFood)
	for _, food := range foods {
		foodMap[food.ID] = *food
	}

	likedFoods := make([]*models.LikedFoodResponse, 0, len(likes))
	for _, like := range likes {
		if food, exists := foodMap[like.FoodID]; exists {
			likedFoods = append(likedFoods, &models.LikedFoodResponse{
				Food:    food,
				LikedAt: like.CreatedAt,
			})
		}
	}

	return likedFoods, nil
}
