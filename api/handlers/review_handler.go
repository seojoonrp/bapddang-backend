// api/handlers/review_handler.go

// 리뷰 관련 로직(리뷰 생성, 수정, 삭제, 조회 등)을 처리하는 API 핸들러

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
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

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	newReview, err := h.reviewService.CreateReview(c, req, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"review": newReview})
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	reviewID := c.Param("reviewID")

	updatedReview, err := h.reviewService.UpdateReview(c, reviewID, userID, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": updatedReview})
}

func (h *ReviewHandler) GetMyReviewsByDay(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	dayStr := c.Query("day")
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		c.Error(apperr.BadRequest("invalid day query parameter", err))
		return
	}

	reviews, err := h.reviewService.GetMyReviewsByDay(c, userID, day)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}
