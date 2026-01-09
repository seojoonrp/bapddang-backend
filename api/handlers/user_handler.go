// api/handlers/user_handler.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService services.UserService
	foodService services.FoodService
}

func NewUserHandler(userService services.UserService, foodService services.FoodService) *UserHandler {
	return &UserHandler{
		userService: userService,
		foodService: foodService,
	}
}

func (h *UserHandler) CheckUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.Error(apperr.BadRequest("username query parameter is required", nil))
		return
	}

	exists, err := h.userService.CheckUsernameExists(c, username)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var req models.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	err := h.userService.SignUp(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, user, err := h.userService.Login(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   false,
	})
}

func (h *UserHandler) GoogleLogin(c *gin.Context) {
	var req struct {
		IDToken string `json:"idToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	isNew, token, user, err := h.userService.LoginWithGoogle(c, req.IDToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   isNew,
	})
}

func (h *UserHandler) KakaoLogin(c *gin.Context) {
	var req struct {
		AccessToken string `json:"accessToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	isNew, token, user, err := h.userService.LoginWithKakao(c, req.AccessToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   isNew,
	})
}

func (h *UserHandler) AppleLogin(c *gin.Context) {
	var req struct {
		IdentityToken string `json:"identityToken" binding:"required"`
		FullName      struct {
			GivenName  string `json:"givenName"`
			FamilyName string `json:"familyName"`
		} `json:"fullName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	isNew, token, user, err := h.userService.LoginWithApple(c, req.IdentityToken)

	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   isNew,
	})
}

func (h *UserHandler) GetMe(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	ctx.JSON(http.StatusOK, userCtx)
}

func (h *UserHandler) LikeFood(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userID := userCtx.(models.User).ID

	foodIDHex := ctx.Param("foodID")
	foodID, err := primitive.ObjectIDFromHex(foodIDHex)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID format"})
		return
	}

	wasAdded, err := h.userService.LikeFood(ctx, userID, foodID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if wasAdded {
		go h.foodService.UpdateLikeStats(ctx, foodID, 1)

		ctx.JSON(http.StatusOK, gin.H{"message": "Food liked successfully"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food is already liked"})
	}
}

func (h *UserHandler) UnlikeFood(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userID := userCtx.(models.User).ID

	foodIDHex := ctx.Param("foodID")
	foodID, err := primitive.ObjectIDFromHex(foodIDHex)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID format"})
		return
	}

	wasRemoved, err := h.userService.UnlikeFood(ctx, userID, foodID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if wasRemoved {
		go h.foodService.UpdateLikeStats(ctx, foodID, -1)

		ctx.JSON(http.StatusOK, gin.H{"message": "Food unliked successfully"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food was not liked"})
	}
}

func (h *UserHandler) GetLikedFoods(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userID := userCtx.(models.User).ID

	likedFoodIDs, err := h.userService.GetLikedFoodIDs(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get liked food ids"})
		return
	}

	foods, err := h.foodService.GetStandardFoodsByIDs(likedFoodIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch liked foods from ids"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likedFoods": foods})
}

func (h *UserHandler) SyncUserDay(ctx *gin.Context) {
	userID, err := GetUserID(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = h.userService.SyncUserDay(ctx, userID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user day synchronized successfully"})
}
