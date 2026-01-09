// api/handlers/food_handler.go

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodHandler struct {
	foodService services.FoodService
}

func NewFoodHandler(foodService services.FoodService) *FoodHandler {
	return &FoodHandler{
		foodService: foodService,
	}
}

func (h *FoodHandler) GetStandardFoodByID(ctx *gin.Context) {
	foodIDStr := ctx.Param("foodID")

	food, err := h.foodService.GetStandardFoodByID(ctx, foodIDStr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID"})
		return
	}

	ctx.JSON(http.StatusOK, food)
}

func (h *FoodHandler) CreateStandardFood(ctx *gin.Context) {
	var input models.NewStandardFoodInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	newFood, err := h.foodService.CreateStandardFood(ctx, input)
	if err != nil {
		if err.Error() == "food already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Food already exists"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create food"})
		return
	}

	ctx.JSON(http.StatusCreated, newFood)
}

func (h *FoodHandler) FindOrCreateCustomFood(ctx *gin.Context) {
	var input models.NewCustomFoodInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(models.User)

	customFood, err := h.foodService.FindOrCreateCustomFood(ctx, input, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create custom food"})
		return
	}

	ctx.JSON(http.StatusOK, customFood)
}

func (h *FoodHandler) GetMainFeedFoods(ctx *gin.Context) {
	foodType := ctx.Query("type")
	if foodType != "meal" && foodType != "dessert" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food type"})
		return
	}
	speed := ctx.Query("speed")
	if speed != "fast" && speed != "slow" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid speed"})
		return
	}
	foodCountStr := ctx.Query("count")
	if foodCountStr == "" {
		foodCountStr = "10"
	}
	foodCount, err := strconv.Atoi(foodCountStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid count"})
		return
	}

	selectedFoods, err := h.foodService.GetMainFeedFoods(foodType, speed, foodCount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get main feed foods"})
		return
	}

	ctx.JSON(http.StatusOK, selectedFoods)
}
