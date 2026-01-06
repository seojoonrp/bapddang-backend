// api/handlers/review_handler.go

// 리뷰 관련 로직(리뷰 생성, 수정, 삭제, 조회 등)을 처리하는 API 핸들러

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewHandler struct {
	reviewService services.ReviewService
	foodService   services.FoodService
}

func NewReviewHandler(reviewService services.ReviewService, foodService services.FoodService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
		foodService:   foodService,
	}
}

func (h *ReviewHandler) CreateReview(ctx *gin.Context) {
	var input models.ReviewInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review request format"})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(models.User)

	newReview, err := h.reviewService.CreateReview(ctx, input, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create review"})
		return
	}

	standardFoods := make([]primitive.ObjectID, 0)
	for _, foodItem := range newReview.Foods {
		if foodItem.FoodType == "standard" {
			standardFoods = append(standardFoods, foodItem.FoodID)
		}
	}
	if len(standardFoods) > 0 {
		go h.foodService.UpdateCreatedReviewStats(ctx, standardFoods, input.Rating)
	}

	ctx.JSON(http.StatusCreated, newReview)
}

func (h *ReviewHandler) UpdateReview(ctx *gin.Context) {
	var input models.ReviewInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review request format"})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(models.User)

	reviewIDHex := ctx.Param("reviewID")
	reviewID, err := primitive.ObjectIDFromHex(reviewIDHex)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID format"})
		return
	}

	updatedReview, oldRating, err := h.reviewService.UpdateReview(ctx, reviewID, input, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	standardFoods := make([]primitive.ObjectID, 0)
	for _, foodItem := range updatedReview.Foods {
		if foodItem.FoodType == "standard" {
			standardFoods = append(standardFoods, foodItem.FoodID)
		}
	}
	if len(standardFoods) > 0 && oldRating != input.Rating {
		go h.foodService.UpdateModifiedReviewStats(ctx, standardFoods, oldRating, input.Rating)
	}

	ctx.JSON(http.StatusOK, updatedReview)
}

func (h *ReviewHandler) GetMyReviewsByDay(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(models.User)

	dayStr := ctx.Query("day")
	if dayStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Day query parameter is required"})
		return
	}

	day, err := strconv.Atoi(dayStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day query parameter"})
		return
	}

	reviews, err := h.reviewService.GetMyReviewsByDay(ctx, user.ID, day)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch reviews"})
		return
	}

	ctx.JSON(http.StatusOK, reviews)
}
