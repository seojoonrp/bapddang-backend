// routes/routes.go

package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/api/middleware"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/api/services"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(router *gin.Engine, db *mongo.Database) {
	userCollection := db.Collection("users")
	userRepository := repositories.NewUserRepository(userCollection)

	standardFoodCollection := db.Collection("standard_foods")
	customFoodCollection := db.Collection("custom_foods")
	foodRepository := repositories.NewFoodRepository(standardFoodCollection, customFoodCollection)

	reviewCollection := db.Collection("reviews")
	reviewRepository := repositories.NewReviewRepository(reviewCollection)

	userService := services.NewUserService(userRepository, foodRepository)
	foodService := services.NewFoodService(context.Background(), foodRepository)
	reviewService := services.NewReviewService(reviewRepository, foodRepository)

	userHandler := handlers.NewUserHandler(userService, foodService)
	foodHandler := handlers.NewFoodHandler(foodService)
	reviewHandler := handlers.NewReviewHandler(reviewService, foodService)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Bobttaeng server is running!",
			})
		})

		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.GET("/check-username", userHandler.CheckUsernameExists)
			authRoutes.POST("/signup", userHandler.SignUp)
			authRoutes.POST("/login", userHandler.Login)
			authRoutes.POST("/google", userHandler.GoogleLogin)
			authRoutes.POST("/kakao", userHandler.KakaoLogin)
			authRoutes.POST("/apple", userHandler.AppleLogin)
		}

		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(userCollection))
		{
			protected.GET("/auth/me", userHandler.GetMe)

			protected.GET("/liked-foods", userHandler.GetLikedFoods)

			protected.POST("/foods/:foodID/like", userHandler.LikeFood)
			protected.DELETE("/foods/:foodID/like", userHandler.UnlikeFood)
			protected.POST("/custom-foods", foodHandler.FindOrCreateCustomFood)
			protected.POST("/foods/validate", foodHandler.ValidateFoods)

			protected.POST("/reviews", reviewHandler.CreateReview)
			protected.GET("/reviews/me", reviewHandler.GetMyReviewsByDay)
		}

		apiV1.GET("/foods/:foodID", foodHandler.GetStandardFoodByID)
		apiV1.GET("/foods/main-feed", foodHandler.GetMainFeedFoods)

		adminRoutes := apiV1.Group("/admin")
		{
			adminRoutes.POST("/new-food", foodHandler.CreateStandardFood)
		}
	}
}
