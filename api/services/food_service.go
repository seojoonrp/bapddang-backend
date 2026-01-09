// api/services/food_service.go

package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodService interface {
	GetStandardByID(ctx context.Context, id string) (*models.StandardFood, error)
	CreateStandard(ctx context.Context, req models.CreateStandardFoodRequest) (*models.StandardFood, error)
	FindOrCreateCustom(ctx context.Context, req models.CreateCustomFoodRequest, userID string) (*models.CustomFood, error)

	GetMainFeedFoods(speed string, foodCount int) ([]*models.StandardFood, error)

	IncrementLikeCount(ctx context.Context, foodID string) error
	DecrementLikeCount(ctx context.Context, foodID string) error
}

type foodService struct {
	foodRepo  repositories.FoodRepository
	cacheLock sync.RWMutex
}

func NewFoodService(ctx context.Context, foodRepo repositories.FoodRepository) FoodService {
	return &foodService{
		foodRepo: foodRepo,
	}
}

func (s *foodService) GetStandardByID(ctx context.Context, foodID string) (*models.StandardFood, error) {
	fID, err := primitive.ObjectIDFromHex(foodID)
	if err != nil {
		return nil, apperr.BadRequest("invalid food ID format", err)
	}

	food, err := s.foodRepo.FindStandardByID(ctx, fID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch standard food", err)
	}
	if food == nil {
		return nil, apperr.NotFound("food not found", nil)
	}

	return food, nil
}

func (s *foodService) CreateStandard(ctx context.Context, req models.CreateStandardFoodRequest) (*models.StandardFood, error) {
	food, err := s.foodRepo.FindStandardByName(ctx, req.Name)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch standard food", err)
	}
	if food == nil {
		return nil, apperr.Conflict("food already exists", nil)
	}

	newFood := &models.StandardFood{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		ImageURL:    req.ImageURL,
		Speed:       req.Speed,
		Categories:  req.Categories,
		LikeCount:   0,
		ReviewCount: 0,
		TotalRating: 0,
	}

	err = s.foodRepo.CreateStandard(ctx, newFood)
	if err != nil {
		return nil, apperr.InternalServerError("failed to create standard food", err)
	}

	return newFood, nil
}

func (s *foodService) FindOrCreateCustom(ctx context.Context, req models.CreateCustomFoodRequest, userID string) (*models.CustomFood, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	existingFood, err := s.foodRepo.FindCustomByName(ctx, req.Name)
	if err == mongo.ErrNoDocuments {
		newFood := &models.CustomFood{
			ID:           primitive.NewObjectID(),
			Name:         req.Name,
			UsingUserIDs: []primitive.ObjectID{uID},
			CreatedAt:    time.Now(),
		}
		err := s.foodRepo.CreateCustom(ctx, newFood)
		if err != nil {
			return nil, apperr.InternalServerError("failed to create custom food", err)
		}

		return newFood, nil
	}
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch custom food", err)
	}

	existed, err := s.foodRepo.AddUserToCustom(ctx, existingFood.ID, uID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to add user to custom food", err)
	}
	if existed {
		return existingFood, nil
	}

	existingFood.UsingUserIDs = append(existingFood.UsingUserIDs, uID)
	return existingFood, nil
}

func (s *foodService) GetMainFeedFoods(speed string, foodCount int) ([]*models.StandardFood, error) {
	if foodCount <= 0 {
		return nil, apperr.BadRequest("food count must be positive", nil)
	}
	if foodCount > 10 {
		log.Println("Someone requested too many foods for main feed:", foodCount)
		foodCount = 10
	}

	if speed != models.SpeedFast && speed != models.SpeedSlow {
		return nil, apperr.BadRequest("invalid speed type", nil)
	}

	// 임시로 그냥 랜덤 뽑기 설정
	foods, err := s.foodRepo.GetRandomStandard(context.Background(), speed, foodCount)
	if err != nil {
		return nil, apperr.InternalServerError("failed to get main feed foods", err)
	}

	return foods, nil
}

func (s *foodService) IncrementLikeCount(ctx context.Context, foodID string) error {
	fID, err := primitive.ObjectIDFromHex(foodID)
	if err != nil {
		return apperr.BadRequest("invalid food ID format", err)
	}

	err = s.foodRepo.IncrementLikeCount(ctx, fID)
	if err != nil {
		return apperr.InternalServerError("failed to increment like count", err)
	}

	return nil
}

func (s *foodService) DecrementLikeCount(ctx context.Context, foodID string) error {
	fID, err := primitive.ObjectIDFromHex(foodID)
	if err != nil {
		return apperr.BadRequest("invalid food ID format", err)
	}

	err = s.foodRepo.DecrementLikeCount(ctx, fID)
	if err != nil {
		return apperr.InternalServerError("failed to decrement like count", err)
	}

	return nil
}
