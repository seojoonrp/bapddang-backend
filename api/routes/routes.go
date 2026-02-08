// routes/routes.go

package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/api/middleware"
	_ "github.com/seojoonrp/bapddang-server/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(
	router *gin.Engine,
	db *mongo.Database,
	userHandler *handlers.UserHandler,
	foodHandler *handlers.FoodHandler,
	reviewHandler *handlers.ReviewHandler,
	likeHandler *handlers.LikeHandler,
	marshmallowHandler *handlers.MarshmallowHandler,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

		users := apiV1.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("/me", userHandler.GetMe)
			users.PATCH("/me/agreement", userHandler.AgreeTerms)
			users.GET("/me/liked-foods", likeHandler.GetLikedFoods)
			users.GET("/me/reviews", reviewHandler.GetMyReviewsByDay)
			users.PATCH("/me/sync", userHandler.SyncUserDayAndWeek)
			users.DELETE("/me", userHandler.DeleteUser)
		}

		foods := apiV1.Group("/foods")
		{
			foods.GET("/:foodID", foodHandler.GetStandardFoodByID)

			protectedFoods := foods.Group("")
			protectedFoods.Use(middleware.AuthMiddleware())
			{
				protectedFoods.POST("/:foodID/likes", likeHandler.LikeFood)
				protectedFoods.DELETE("/:foodID/likes", likeHandler.UnlikeFood)

				protectedFoods.GET("/main-feed", foodHandler.GetMainFeedFoods)
				protectedFoods.GET("/category", foodHandler.GetFoodsByCategories)

				protectedFoods.POST("/resolve", foodHandler.ResolveFoodItems)
			}
		}

		reviews := apiV1.Group("/reviews")
		reviews.Use(middleware.AuthMiddleware())
		{
			reviews.POST("", reviewHandler.Create)
			reviews.PATCH("/:reviewID", reviewHandler.Update)
			reviews.DELETE("/:reviewID", reviewHandler.Delete)

			reviews.GET("/recent", reviewHandler.GetRecentWithStandardFood)
		}

		marshmallows := apiV1.Group("/marshmallows")
		marshmallows.Use(middleware.AuthMiddleware())
		{
			marshmallows.GET("", marshmallowHandler.GetUserMarshmallows)
		}

		adminRoutes := apiV1.Group("/admin") // no protection for now
		{
			adminRoutes.POST("/standard-foods", foodHandler.CreateStandardFoods)
		}
	}
}
