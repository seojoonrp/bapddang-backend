// api/handlers/review.go

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

// @Summary 리뷰 생성
// @Description 새로운 리뷰를 생성한다.
// @Tags Review
// @Accept json
// @Produce json
// @Param review body models.CreateReviewRequest true "리뷰 정보"
// @Success 201 {object} response.Response{data=models.Review} "리뷰 생성 성공"
// @Failure 400 {object} response.Response "잘못된 요청 본문"
// @Failure 401 {object} response.Response "인증 실패"
// @Security BearerAuth
// @Router /reviews [post]
func (h *ReviewHandler) Create(c *gin.Context) {
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

	newReview, err := h.reviewService.Create(c, req, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    newReview,
	})
}

// @Summary 리뷰 수정
// @Description 기존 리뷰 정보를 수정한다.
// @Tags Review
// @Accept json
// @Produce json
// @Param reviewID path string true "리뷰 ID"
// @Param review body models.UpdateReviewRequest true "리뷰 수정 정보"
// @Success 200 {object} response.Response{data=models.Review} "리뷰 수정 성공"
// @Failure 401 {object} response.Response "수정 권한 없음"
// @Security BearerAuth
// @Router /reviews/{reviewID} [patch]
func (h *ReviewHandler) Update(c *gin.Context) {
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

	updatedReview, err := h.reviewService.Update(c, reviewID, userID, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    updatedReview,
	})
}

// @Summary 리뷰 삭제
// @Description 특정 리뷰를 삭제한다.
// @Tags Review
// @Accept json
// @Produce json
// @Param reviewID path string true "리뷰 ID"
// @Success 200 {object} response.Response{data=string} "리뷰 삭제 성공"
// @Failure 401 {object} response.Response "삭제 권한 없음"
// @Security BearerAuth
// @Router /reviews/{reviewID} [delete]
func (h *ReviewHandler) Delete(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	reviewID := c.Param("reviewID")

	err = h.reviewService.Delete(c, reviewID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    "Review deleted successfully",
	})
}

// @Summary 날짜별 리뷰 조회
// @Description 특정 날짜에 작성한 리뷰 목록을 조회한다.
// @Tags Review
// @Accept json
// @Produce json
// @Param day query int true "조회할 날짜"
// @Success 200 {object} response.Response{data=[]models.Review} "리뷰 목록 조회 성공"
// @Security BearerAuth
// @Router /users/me/reviews [get]
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

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    reviews,
	})
}
