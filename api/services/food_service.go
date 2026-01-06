// api/services/food_service.go

package services

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/models"
	"github.com/seojoonrp/bapddang-server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodService interface {
	GetStandardFoodByID(ctx context.Context, id string) (*models.StandardFood, error)
	GetStandardFoodsByIDs(ids []primitive.ObjectID) ([]*models.StandardFood, error)
	CreateStandardFood(ctx context.Context, input models.NewStandardFoodInput) (*models.StandardFood, error)
	FindOrCreateCustomFood(ctx context.Context, input models.NewCustomFoodInput, user models.User) (*models.CustomFood, error)

	GetMainFeedFoods(foodType, speed string, foodCount int) ([]*models.StandardFood, error)
	ValidateFoods(ctx context.Context, names []string, userID primitive.ObjectID) ([]models.ValidationResult, error)

	UpdateCreatedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, rating int) error
	UpdateModifiedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, oldRating, newRating int) error
	UpdateLikeStats(ctx context.Context, foodID primitive.ObjectID, increment int) error

	SyncRatingStatsCache(foodID primitive.ObjectID, oldRating, newRating int) error
}

type foodService struct {
	foodRepo          repositories.FoodRepository
	standardFoodCache []*models.StandardFood
	customFoodCache   []*models.CustomFood
	cacheLock         sync.RWMutex
}

func NewFoodService(ctx context.Context, foodRepo repositories.FoodRepository) FoodService {
	allStandardFoods, err := foodRepo.GetAllStandardFoods(ctx)
	if err != nil {
		log.Fatal("FATAL: Failed to load standard food cache: ", err)
	}
	allCustomFoods, err := foodRepo.GetAllCustomFoods(ctx)
	if err != nil {
		log.Fatal("FATAL: Failed to load custom food cache: ", err)
	}

	log.Printf("Successfully loaded %d standard foods into cache", len(allStandardFoods))
	log.Printf("Successfully loaded %d custom foods into cache", len(allCustomFoods))

	return &foodService{
		foodRepo:          foodRepo,
		standardFoodCache: allStandardFoods,
		customFoodCache:   allCustomFoods,
		cacheLock:         sync.RWMutex{},
	}
}

func (s *foodService) GetStandardFoodByID(ctx context.Context, id string) (*models.StandardFood, error) {
	foodID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("Invalid food id")
	}

	food, err := s.foodRepo.FindStandardFoodByID(ctx, foodID)
	if err != nil {
		return nil, err
	}

	return food, nil
}

func (s *foodService) GetStandardFoodsByIDs(ids []primitive.ObjectID) ([]*models.StandardFood, error) {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()

	results := make([]*models.StandardFood, 0, len(ids))

	foodMap := make(map[primitive.ObjectID]*models.StandardFood)
	for _, food := range s.standardFoodCache {
		foodMap[food.ID] = food
	}

	for _, id := range ids {
		if food, exists := foodMap[id]; exists {
			results = append(results, food)
		}
	}

	return results, nil
}

func (s *foodService) CreateStandardFood(ctx context.Context, input models.NewStandardFoodInput) (*models.StandardFood, error) {
	_, err := s.foodRepo.FindStandardFoodByName(ctx, input.Name)
	if err == nil {
		return nil, errors.New("food already exists")
	}

	newFood := &models.StandardFood{
		ID:          primitive.NewObjectID(),
		Name:        input.Name,
		ImageURL:    input.ImageURL,
		Speed:       input.Speed,
		Type:        input.Type,
		Categories:  input.Categories,
		LikeCount:   0,
		ReviewCount: 0,
		TotalRating: 0,
	}
	err = s.foodRepo.SaveStandardFood(ctx, newFood)
	if err != nil {
		return nil, err
	}

	s.cacheLock.Lock()
	s.standardFoodCache = append(s.standardFoodCache, newFood)
	s.cacheLock.Unlock()

	return newFood, nil
}

func (s *foodService) FindOrCreateCustomFood(ctx context.Context, input models.NewCustomFoodInput, user models.User) (*models.CustomFood, error) {
	existingFood, err := s.foodRepo.FindCustomFoodByName(ctx, input.Name)

	if err == mongo.ErrNoDocuments {
		newFood := &models.CustomFood{
			ID:           primitive.NewObjectID(),
			Name:         input.Name,
			UsingUserIDs: []primitive.ObjectID{user.ID},
			CreatedAt:    time.Now(),
		}
		err := s.foodRepo.SaveCustomFood(ctx, newFood)
		if err != nil {
			return nil, err
		}

		s.cacheLock.Lock()
		s.customFoodCache = append(s.customFoodCache, newFood)
		s.cacheLock.Unlock()

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
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()

	candidates := make([]*models.StandardFood, 0)
	for _, food := range s.standardFoodCache {
		if food.Type == foodType && food.Speed == speed {
			candidates = append(candidates, food)
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	resultList := make([]*models.StandardFood, 0, foodCount)
	usedCategories := make(map[string]bool)

	for _, food := range candidates {
		if len(resultList) >= foodCount {
			break
		}
		if len(food.Categories) == 0 {
			resultList = append(resultList, food)
			continue
		}

		uniqueCategory := true
		for _, category := range food.Categories {
			if usedCategories[category] {
				uniqueCategory = false
				break
			}
		}
		if uniqueCategory {
			resultList = append(resultList, food)
			for _, category := range food.Categories {
				usedCategories[category] = true
			}
		}
	}

	log.Printf("Selected %d main feed foods", len(resultList))

	return resultList, nil
}

type matchCandidates struct {
	Score  float64
	Output models.ValidationOutput
}

func (s *foodService) ValidateFoods(ctx context.Context, names []string, userID primitive.ObjectID) ([]models.ValidationResult, error) {
	results := make([]models.ValidationResult, 0, len(names))

	const similarityThreshold = 0.75
	const maxSuggestions = 3

	for _, name := range names {
		result := models.ValidationResult{OriginalName: name}

		standardFood, err := s.foodRepo.FindStandardFoodByName(ctx, name)
		if err == nil {
			result.Status = "ok"
			result.OkOutput = &models.ValidationOutput{
				ID:   standardFood.ID,
				Name: standardFood.Name,
				Type: "standard",
			}
			results = append(results, result)
			continue
		}

		customFood, err := s.foodRepo.FindCustomFoodByName(ctx, name)
		if err == nil {
			result.Status = "ok"
			result.OkOutput = &models.ValidationOutput{
				ID:   customFood.ID,
				Name: customFood.Name,
				Type: "custom",
			}
			results = append(results, result)
			continue
		}

		var candidates []matchCandidates

		s.cacheLock.RLock()

		for _, food := range s.standardFoodCache {
			score := utils.Score(name, food.Name)
			if score >= similarityThreshold {
				candidates = append(candidates, matchCandidates{
					Score: score,
					Output: models.ValidationOutput{
						ID:   food.ID,
						Name: food.Name,
						Type: "standard",
					},
				})
			}
		}

		for _, food := range s.customFoodCache {
			score := utils.Score(name, food.Name)
			if score >= similarityThreshold {
				candidates = append(candidates, matchCandidates{
					Score: score,
					Output: models.ValidationOutput{
						ID:   food.ID,
						Name: food.Name,
						Type: "custom",
					},
				})
			}
		}

		s.cacheLock.RUnlock()

		if len(candidates) == 0 {
			s.cacheLock.Lock()

			// TODO : Lock 걸면서 중복 생성됐는지 이중체크

			newID := primitive.NewObjectID()
			newCustomFood := &models.CustomFood{
				ID:           newID,
				Name:         name,
				UsingUserIDs: []primitive.ObjectID{userID},
				CreatedAt:    time.Now(),
			}

			err := s.foodRepo.SaveCustomFood(ctx, newCustomFood)
			if err != nil {
				s.cacheLock.Unlock()
				return nil, err
			}

			s.customFoodCache = append(s.customFoodCache, newCustomFood)

			result.Status = "new"
			result.NewOutput = &models.ValidationOutput{
				ID:   newID,
				Name: name,
				Type: "new",
			}
			results = append(results, result)

			s.cacheLock.Unlock()
			continue
		}

		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})

		limit := min(len(candidates), maxSuggestions)
		for i := range limit {
			result.SuggestionOutputs = append(result.SuggestionOutputs, candidates[i].Output)
		}

		result.Status = "suggestion"
		results = append(results, result)
	}

	return results, nil
}

func (s *foodService) UpdateCreatedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, rating int) error {
	err := s.foodRepo.UpdateCreatedReviewStats(ctx, foodIDs, rating)
	if err != nil {
		return err
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	for _, foodID := range foodIDs {
		err := s.SyncRatingStatsCache(foodID, 0, rating)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *foodService) UpdateModifiedReviewStats(ctx context.Context, foodIDs []primitive.ObjectID, oldRating, newRating int) error {
	err := s.foodRepo.UpdateModifiedReviewStats(ctx, foodIDs, oldRating, newRating)
	if err != nil {
		return err
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	for _, foodID := range foodIDs {
		err := s.SyncRatingStatsCache(foodID, oldRating, newRating)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *foodService) UpdateLikeStats(ctx context.Context, foodID primitive.ObjectID, increment int) error {
	var err error
	if increment > 0 {
		err = s.foodRepo.IncrementLikeCount(ctx, foodID)
	} else if increment < 0 {
		err = s.foodRepo.DecrementLikeCount(ctx, foodID)
	}
	if err != nil {
		return err
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	for _, food := range s.standardFoodCache {
		if food.ID == foodID {
			food.LikeCount += increment
			if food.LikeCount < 0 {
				food.LikeCount = 0
			}
			break
		}
	}

	return nil
}

func (s *foodService) SyncRatingStatsCache(foodID primitive.ObjectID, oldRating, newRating int) error {
	for _, food := range s.standardFoodCache {
		if food.ID == foodID {
			food.TotalRating = food.TotalRating - oldRating + newRating
			break
		}
	}

	return nil
}
