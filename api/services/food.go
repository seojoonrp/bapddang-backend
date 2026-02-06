// api/services/food.go

package services

import (
	"context"
	"log"
	"math/rand"
	"sort"
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

	GetMainFeedFoods(ctx context.Context, userID string, speed string, count int) ([]models.FoodLikeResponse, error)
	GetFoodsByCategories(ctx context.Context, userID string, speed string, categories []string, count int) ([]models.FoodLikeResponse, error)

	IncrementLikeCount(ctx context.Context, foodID string) error
	DecrementLikeCount(ctx context.Context, foodID string) error
}

type foodService struct {
	foodRepo       repositories.FoodRepository
	likeRepo       repositories.LikeRepository
	recHistoryRepo repositories.RecHistoryRepository
	cacheLock      sync.RWMutex
}

func NewFoodService(
	ctx context.Context,
	fr repositories.FoodRepository,
	lr repositories.LikeRepository,
	rhr repositories.RecHistoryRepository,
) FoodService {
	return &foodService{
		foodRepo:       fr,
		likeRepo:       lr,
		recHistoryRepo: rhr,
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

func (s *foodService) getRecommendedFoods(ctx context.Context, userID primitive.ObjectID, speed string, count int) ([]models.StandardFood, error) {
	candidateCount := count * 7
	candidates, err := s.foodRepo.GetRandomStandards(ctx, speed, nil, candidateCount)
	if err != nil {
		return nil, apperr.InternalServerError("failed to get candidate foods", err)
	}

	historyMap, err := s.recHistoryRepo.GetRecentFoodIDsMap(ctx, userID, 2)
	if err != nil {
		return nil, apperr.InternalServerError("failed to get recommendation history", err)
	}

	type scoredFood struct {
		food  models.StandardFood
		score float64
	}
	scoredList := make([]scoredFood, 0, len(candidates))
	now := time.Now()

	for _, food := range candidates {
		weight := 1.0
		if lastSeen, ok := historyMap[food.ID]; ok {
			hoursSinceSeen := now.Sub(lastSeen).Hours()
			if hoursSinceSeen < 1 {
				weight = 0.01
			} else if hoursSinceSeen < 6 {
				weight = 0.1
			} else if hoursSinceSeen < 18 {
				weight = 0.4
			} else {
				weight = 0.8
			}
		}

		scoredList = append(scoredList, scoredFood{
			food:  food,
			score: weight * (0.8 + rand.Float64()*0.4),
		})
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})

	var finalFoods []models.StandardFood

	excludedParents, err := s.recHistoryRepo.GetLatestParents(ctx, userID, 7)
	if err != nil {
		return nil, apperr.InternalServerError("failed to get latest parents from history", err)
	}

	usedParents := make(map[string]bool)
	for _, parent := range excludedParents {
		usedParents[parent] = true
	}

	for _, sf := range scoredList {
		if len(finalFoods) >= count {
			break
		}

		isOverlap := false
		for _, parent := range sf.food.Parents {
			if usedParents[parent] {
				isOverlap = true
				break
			}
		}

		if !isOverlap {
			finalFoods = append(finalFoods, sf.food)
			for _, parent := range sf.food.Parents {
				usedParents[parent] = true
			}
		}
	}

	// 혹시라도 부족할 시 부모 상관없이 채우기
	if len(finalFoods) < count {
		for _, sf := range scoredList {
			if len(finalFoods) >= count {
				break
			}

			alreadyAdded := false
			for _, f := range finalFoods {
				if f.ID == sf.food.ID {
					alreadyAdded = true
					break
				}
			}

			if !alreadyAdded {
				finalFoods = append(finalFoods, sf.food)
			}
		}
	}

	finalIDs := make([]primitive.ObjectID, 0, len(finalFoods))
	finalParents := make([]string, 0)
	for _, food := range finalFoods {
		finalIDs = append(finalIDs, food.ID)
		finalParents = append(finalParents, food.Parents...)
	}
	newHistory := models.RecHistory{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		FoodIDs:   finalIDs,
		Parents:   finalParents,
		CreatedAt: now,
	}

	go func(h models.RecHistory) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.recHistoryRepo.SaveHistory(bgCtx, h); err != nil {
			log.Printf("[WARNING] failed to save recommendation history: %v", err)
		}
	}(newHistory)

	return finalFoods, nil
}

func (s *foodService) wrapWithLikeStatus(ctx context.Context, userID primitive.ObjectID, foods []models.StandardFood) ([]models.FoodLikeResponse, error) {
	foodIDs := make([]primitive.ObjectID, 0, len(foods))
	for _, food := range foods {
		foodIDs = append(foodIDs, food.ID)
	}

	likedMap, err := s.likeRepo.CheckLikedStatus(ctx, userID, foodIDs)
	if err != nil {
		return nil, apperr.InternalServerError("failed to check liked status", err)
	}

	var responses []models.FoodLikeResponse
	for _, food := range foods {
		responses = append(responses, models.FoodLikeResponse{
			Food:    food,
			IsLiked: likedMap[food.ID],
		})
	}

	return responses, nil
}

func (s *foodService) GetMainFeedFoods(ctx context.Context, userID string, speed string, count int) ([]models.FoodLikeResponse, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	if count <= 0 || count > 10 {
		return nil, apperr.BadRequest("invalid food count", nil)
	}

	if speed != models.SpeedFast && speed != models.SpeedSlow {
		return nil, apperr.BadRequest("invalid speed type", nil)
	}

	foods, err := s.getRecommendedFoods(ctx, uID, speed, count)
	if err != nil {
		return nil, err
	}

	return s.wrapWithLikeStatus(ctx, uID, foods)
}

func (s *foodService) GetFoodsByCategories(ctx context.Context, userID string, speed string, categories []string, count int) ([]models.FoodLikeResponse, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	if count <= 0 || count > 10 {
		return nil, apperr.BadRequest("invalid food count", nil)
	}

	if speed != models.SpeedFast && speed != models.SpeedSlow {
		return nil, apperr.BadRequest("invalid speed type", nil)
	}

	foods, err := s.foodRepo.GetRandomStandards(ctx, speed, categories, count)
	if err != nil {
		return nil, apperr.InternalServerError("failed to get foods by categories", err)
	}

	return s.wrapWithLikeStatus(ctx, uID, foods)
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
