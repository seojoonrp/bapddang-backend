// api/handlers/like_handler.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
)

type LikeHandler struct {
	likeService services.LikeService
}

func NewLikeHandler(ls services.LikeService) *LikeHandler {
	return &LikeHandler{
		likeService: ls,
	}
}

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

	c.JSON(http.StatusOK, gin.H{"message": "food liked successfully"})
}

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

	c.JSON(http.StatusOK, gin.H{"message": "food unliked successfully"})
}

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

	c.JSON(http.StatusOK, gin.H{"liked_foods": foods})
}
