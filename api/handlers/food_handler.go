// api/handlers/food_handler.go

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
)

type FoodHandler struct {
	foodService services.FoodService
}

func NewFoodHandler(foodService services.FoodService) *FoodHandler {
	return &FoodHandler{
		foodService: foodService,
	}
}

func (h *FoodHandler) GetStandardFoodByID(c *gin.Context) {
	foodID := c.Param("foodID")

	food, err := h.foodService.GetStandardByID(c, foodID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"standard_food": food})
}

func (h *FoodHandler) CreateStandardFoods(c *gin.Context) {
	var req []models.CreateStandardFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	newFoods, err := h.foodService.CreateStandards(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"standard_foods": newFoods})
}

func (h *FoodHandler) ResolveFoodItems(c *gin.Context) {
	var req models.ResolveFoodItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	foodItems, err := h.foodService.ResolveFoodItems(c, req.Names)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"food_items": foodItems})
}

func (h *FoodHandler) GetMainFeedFoods(c *gin.Context) {
	speed := c.Query("speed")

	foodCountStr := c.Query("count")
	foodCount, err := strconv.Atoi(foodCountStr)
	if err != nil {
		c.Error(apperr.BadRequest("invalid food count", err))
		return
	}

	selectedFoods, err := h.foodService.GetMainFeedFoods(speed, foodCount)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"foods": selectedFoods})
}
