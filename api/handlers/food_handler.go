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

func (h *FoodHandler) CreateStandardFood(c *gin.Context) {
	var req models.CreateStandardFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	newFood, err := h.foodService.CreateStandard(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"standard_food": newFood})
}

func (h *FoodHandler) FindOrCreateCustomFood(c *gin.Context) {
	var req models.CreateCustomFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	newFood, err := h.foodService.FindOrCreateCustom(c, req, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"custom_food": newFood})
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
