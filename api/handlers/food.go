// api/handlers/food.go

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"github.com/seojoonrp/bapddang-server/response"
)

type FoodHandler struct {
	foodService services.FoodService
}

func NewFoodHandler(foodService services.FoodService) *FoodHandler {
	return &FoodHandler{
		foodService: foodService,
	}
}

// @Summary 표준 음식 조회
// @Description 음식 ID를 통해 표준 음식 정보를 조회한다.
// @Tags Food
// @Accept json
// @Produce json
// @Param foodID path string true "음식 ID"
// @Success 200 {object} response.Response{data=models.StandardFood} "음식 조회 성공"
// @Failure 400 {object} response.Response "잘못된 ID 형식"
// @Failure 404 {object} response.Response "음식을 찾을 수 없음"
// @Router /foods/{foodID} [get]
func (h *FoodHandler) GetStandardFoodByID(c *gin.Context) {
	foodID := c.Param("foodID")

	food, err := h.foodService.GetStandardByID(c, foodID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    food,
	})
}

// @Summary 표준 음식 생성
// @Description 관리자 권한으로 새로운 표준 음식들을 생성한다.
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body []models.CreateStandardFoodRequest true "음식 정보 목록"
// @Success 201 {object} response.Response{data=[]models.StandardFood} "음식 생성 성공"
// @Router /admin/standard-foods [post]
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

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    newFoods,
	})
}

// @Summary 리뷰용 음식 반환
// @Description 음식 이름들을 받아 리뷰용 음식 항목들을 생성 및 반환한다.
// @Tags Food
// @Accept json
// @Produce json
// @Param request body models.ResolveFoodItemsRequest true "음식 이름 목록"
// @Success 200 {object} response.Response{data=[]models.ReviewFoodItem} "음식 항목 반환 성공"
// @Security BearerAuth
// @Router /foods/resolve [post]
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

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    foodItems,
	})
}

// @Summary 카뉴 음식 조회
// @Description 유저에게 추천한 기록을 바탕으로 카뉴 음식 목록을 가져온다.
// @Tags Food
// @Accept json
// @Produce json
// @Param speed query string true "속도 (fast/slow)"
// @Param count query int true "조회 개수 (최대 10개)"
// @Success 200 {object} response.Response{data=[]models.MainFeedResponse} "조회 성공"
// @Security BearerAuth
// @Router /foods/main-feed [get]
func (h *FoodHandler) GetMainFeedFoods(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	speed := c.Query("speed")

	foodCountStr := c.Query("count")
	foodCount, err := strconv.Atoi(foodCountStr)
	if err != nil {
		c.Error(apperr.BadRequest("invalid food count", err))
		return
	}

	result, err := h.foodService.GetMainFeedFoods(c, userID, speed, foodCount)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    result,
	})
}
