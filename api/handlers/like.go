// api/handlers/like.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/response"
)

type LikeHandler struct {
	likeService services.LikeService
}

func NewLikeHandler(ls services.LikeService) *LikeHandler {
	return &LikeHandler{
		likeService: ls,
	}
}

// @Summary 음식 좋아요
// @Description 특정 음식에 좋아요를 표시한다.
// @Tags Like
// @Accept json
// @Produce json
// @Param foodID path string true "음식 ID"
// @Success 200 {object} response.Response{data=string} "성공 메시지"
// @Failure 400 {object} response.Response "잘못된 음식 ID 형식"
// @Failure 401 {object} response.Response "인증 실패"
// @Failure 409 {object} response.Response "이미 좋아요한 음식"
// @Security BearerAuth
// @Router /foods/{foodID}/likes [post]
func (h *LikeHandler) LikeFood(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	fIDStr := c.Param("foodID")

	err = h.likeService.LikeFood(c.Request.Context(), userID, fIDStr)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    "food liked successfully",
	})
}

// @Summary 음식 좋아요 취소
// @Description 특정 음식의 좋아요를 취소한다.
// @Tags Like
// @Accept json
// @Produce json
// @Param foodID path string true "음식 ID"
// @Success 200 {object} response.Response{data=string} "성공 메시지"
// @Failure 404 {object} response.Response "좋아요 기록 없음"
// @Security BearerAuth
// @Router /foods/{foodID}/likes [delete]
func (h *LikeHandler) UnlikeFood(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	fIDStr := c.Param("foodID")

	err = h.likeService.UnlikeFood(c.Request.Context(), userID, fIDStr)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    "food unliked successfully",
	})
}

// @Summary 좋아요한 음식 목록 조회
// @Description 현재 유저가 좋아요한 음식 목록을 조회한다.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.StandardFood} "좋아요 음식 목록"
// @Security BearerAuth
// @Router /users/me/liked-foods [get]
func (h *LikeHandler) GetLikedFoods(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	foods, err := h.likeService.GetLikedFoods(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    foods,
	})
}
