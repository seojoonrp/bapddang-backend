// api/services/food_service.go

package services

import (
	"context"
	"sync"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodService interface {
	GetStandardFoodByID(ctx context.Context, id string) (*models.StandardFood, error)
	CreateStandardFood(ctx context.Context, input models.CreateStandardFoodRequest) (*models.StandardFood, error)
	FindOrCreateCustomFood(ctx context.Context, input models.CreateCustomFoodRequest, user models.User) (*models.CustomFood, error)

	GetMainFeedFoods(foodType, speed string, foodCount int) ([]*models.StandardFood, error)

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

func (s *foodService) GetStandardFoodByID(ctx context.Context, foodID string) (*models.StandardFood, error) {
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

func (s *foodService) CreateStandardFood(ctx context.Context, req models.CreateStandardFoodRequest) (*models.StandardFood, error) {
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
		Type:        req.Type,
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

func (s *foodService) FindOrCreateCustomFood(ctx context.Context, input models.CreateCustomFoodRequest, user models.User) (*models.CustomFood, error) {
	existingFood, err := s.foodRepo.FindCustomByName(ctx, input.Name)

	if err == mongo.ErrNoDocuments {
		newFood := &models.CustomFood{
			ID:           primitive.NewObjectID(),
			Name:         input.Name,
			UsingUserIDs: []primitive.ObjectID{user.ID},
			CreatedAt:    time.Now(),
		}
		err := s.foodRepo.CreateCustorm(ctx, newFood)
		if err != nil {
			return nil, err
		}

		return newFood, nil
	}

	if err != nil {
		return nil, err
	}

	err = s.foodRepo.AddUserToCustomFood(ctx, existingFood.ID, user.ID)
	if err != nil {
		return nil, err
	}

	alreadyExists := false
	for _, uid := range existingFood.UsingUserIDs {
		if uid == user.ID {
			alreadyExists = true
			break
		}
	}

	if !alreadyExists {
		existingFood.UsingUserIDs = append(existingFood.UsingUserIDs, user.ID)
	}

	return existingFood, nil
}

func (s *foodService) GetMainFeedFoods(foodType, speed string, foodCount int) ([]*models.StandardFood, error) {
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
