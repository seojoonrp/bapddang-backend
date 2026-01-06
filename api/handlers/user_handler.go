// api/handlers/user_handler.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
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

func (h *UserHandler) CheckUsernameExists(ctx *gin.Context) {
	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Username query parameter is required"})
		return
	}

	exists, err := h.userService.CheckUsernameExists(ctx, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	var input models.SignUpInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.SignUp(ctx, input)
	if err != nil {
		if err.Error() == "user already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var input models.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.userService.Login(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   false,
	})
}

func (h *UserHandler) GoogleLogin(ctx *gin.Context) {
	var input struct {
		IDToken string `json:"idToken" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isNew, token, user, err := h.userService.LoginWithGoogle(ctx, input.IDToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Google login failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   isNew,
	})
}

func (h *UserHandler) KakaoLogin(ctx *gin.Context) {
	var input struct {
		AccessToken string `json:"accessToken" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isNew, token, user, err := h.userService.LoginWithKakao(ctx, input.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Kakao login failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        user,
		"isNewUser":   isNew,
	})
}

func (h *UserHandler) AppleLogin(ctx *gin.Context) {
	var input struct {
		IdentityToken string `json:"identityToken" binding:"required"`
		FullName      struct {
			GivenName  string `json:"givenName"`
			FamilyName string `json:"familyName"`
		} `json:"fullName"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isNew, token, user, err := h.userService.LoginWithApple(ctx, input.IdentityToken)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Apple login failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
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
