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
)

type FoodService interface {
	GetStandardByID(ctx context.Context, id string) (*models.StandardFood, error)
	CreateStandards(ctx context.Context, req []models.CreateStandardFoodRequest) ([]*models.StandardFood, error)
	ResolveFoodItems(ctx context.Context, names []string) ([]models.ReviewFoodItem, error)

	GetMainFeedFoods(ctx context.Context, userID string, speed string, foodCount int) ([]models.MainFeedResponse, error)

	IncrementLikeCount(ctx context.Context, foodID string) error
	DecrementLikeCount(ctx context.Context, foodID string) error
}

type foodService struct {
	foodRepo  repositories.FoodRepository
	likeRepo  repositories.LikeRepository
	cacheLock sync.RWMutex
}

func NewFoodService(ctx context.Context, fr repositories.FoodRepository, lr repositories.LikeRepository) FoodService {
	return &foodService{
		foodRepo: fr,
		likeRepo: lr,
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

func (s *foodService) CreateStandards(ctx context.Context, req []models.CreateStandardFoodRequest) ([]*models.StandardFood, error) {
	var newFoods []*models.StandardFood
	var docs []interface{}

	for _, foodReq := range req {
		newFood := &models.StandardFood{
			ID:          primitive.NewObjectID(),
			Name:        foodReq.Name,
			ImageURL:    foodReq.ImageURL,
			Speed:       foodReq.Speed,
			Parents:     foodReq.Parents,
			Categories:  foodReq.Categories,
			LikeCount:   0,
			ReviewCount: 0,
			TotalRating: 0,
		}
		newFoods = append(newFoods, newFood)
		docs = append(docs, newFood)
	}

	err := s.foodRepo.CreateStandards(ctx, docs)
	if err != nil {
		return nil, apperr.InternalServerError("failed to create standard foods", err)
	}

	return newFoods, nil
}

func (s *foodService) findOrCreateCustom(ctx context.Context, name string) (*models.CustomFood, error) {
	existingFood, err := s.foodRepo.FindCustomByName(ctx, name)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch custom food", err)
	}
	if existingFood == nil {
		newFood := models.CustomFood{
			ID:          primitive.NewObjectID(),
			Name:        name,
			ReviewCount: 0,
			CreatedAt:   time.Now(),
		}
		err := s.foodRepo.CreateCustom(ctx, newFood)
		if err != nil {
			return nil, apperr.InternalServerError("failed to create custom food", err)
		}
		return &newFood, nil
	}

	return existingFood, nil
}

func (s *foodService) ResolveFoodItems(ctx context.Context, names []string) ([]models.ReviewFoodItem, error) {
	if len(names) == 0 {
		return nil, apperr.BadRequest("names list cannot be empty", nil)
	}

	var result []models.ReviewFoodItem

	for _, name := range names {
		standardFood, err := s.foodRepo.FindStandardByName(ctx, name)
		if err != nil {
			return nil, apperr.InternalServerError("failed to fetch standard food by name", err)
		}

		if standardFood != nil {
			result = append(result, models.ReviewFoodItem{
				FoodID:   standardFood.ID.Hex(),
				FoodName: standardFood.Name,
				Type:     models.FoodTypeStandard,
			})
			continue
		}

		customFood, err := s.findOrCreateCustom(ctx, name)
		if err != nil {
			return nil, err
		}

		result = append(result, models.ReviewFoodItem{
			FoodID:   customFood.ID.Hex(),
			FoodName: customFood.Name,
			Type:     models.FoodTypeCustom,
		})
	}

	return result, nil
}

func (s *foodService) GetMainFeedFoods(ctx context.Context, userID string, speed string, foodCount int) ([]models.MainFeedResponse, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	if foodCount <= 0 {
		return nil, apperr.BadRequest("food count must be positive", nil)
	}
	if foodCount > 10 {
		log.Println("[WARNING] Someone requested too many foods for main feed:", foodCount)
		foodCount = 10
	}

	if speed != models.SpeedFast && speed != models.SpeedSlow {
		return nil, apperr.BadRequest("invalid speed type", nil)
	}

	// 임시로 그냥 랜덤 뽑기 설정
	foods, err := s.foodRepo.GetRandomStandards(ctx, speed, foodCount)
	if err != nil {
		return nil, apperr.InternalServerError("failed to get main feed foods", err)
	}

	foodIDs := make([]primitive.ObjectID, 0, len(foods))
	for _, food := range foods {
		foodIDs = append(foodIDs, food.ID)
	}

	likedMap, err := s.likeRepo.CheckLikedStatus(ctx, uID, foodIDs)
	if err != nil {
		return nil, apperr.InternalServerError("failed to check liked status", err)
	}

	responses := make([]models.MainFeedResponse, 0, len(foods))
	for _, food := range foods {
		responses = append(responses, models.MainFeedResponse{
			Food:    food,
			IsLiked: likedMap[food.ID],
		})
	}

	return responses, nil
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
