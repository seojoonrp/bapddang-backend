package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/api/middleware"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/api/routes"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/database"
)

// @title Bobttaeng API Server
// @version 1.0
// @description 밥땡의 백엔드 서버 API 명세서
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Bearer {token}" 형식으로 JWT 토큰을 전달
func main() {
	config.LoadConfig()

	client, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal("Failed to disconnect from DB: ", err)
		}
	}()

	db := client.Database(config.AppConfig.DBName)

	userRepository := repositories.NewUserRepository(db)
	foodRepository := repositories.NewFoodRepository(db)
	reviewRepository := repositories.NewReviewRepository(db)
	likeRepository := repositories.NewLikeRepository(db)
	recHistoryRepository := repositories.NewRecHistoryRepository(db)
	marshmallowRepository := repositories.NewMarshmallowRepository(db)

	userService := services.NewUserService(userRepository, foodRepository, marshmallowRepository)
	foodService := services.NewFoodService(context.Background(), foodRepository, likeRepository, recHistoryRepository)
	reviewService := services.NewReviewService(reviewRepository, foodRepository, userRepository, marshmallowRepository)
	likeService := services.NewLikeService(likeRepository, foodRepository)

	userHandler := handlers.NewUserHandler(userService, foodService)
	foodHandler := handlers.NewFoodHandler(foodService)
	reviewHandler := handlers.NewReviewHandler(reviewService, foodService)
	likeHandler := handlers.NewLikeHandler(likeService)

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(middleware.ErrorHandler())
	router.SetTrustedProxies(nil)

	routes.SetupRoutes(router, db, userHandler, foodHandler, reviewHandler, likeHandler)

	port := config.AppConfig.Port
	log.Println("Server started on port " + port + ".")
	router.Run(":" + port)
}
