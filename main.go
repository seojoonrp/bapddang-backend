package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/api/routes"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/database"
)

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

	userService := services.NewUserService(userRepository, foodRepository)
	foodService := services.NewFoodService(context.Background(), foodRepository)
	reviewService := services.NewReviewService(reviewRepository, foodRepository)

	userHandler := handlers.NewUserHandler(userService, foodService)
	foodHandler := handlers.NewFoodHandler(foodService)
	reviewHandler := handlers.NewReviewHandler(reviewService, foodService)

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(cors.Default())

	routes.SetupRoutes(router, db, userHandler, foodHandler, reviewHandler)

	port := config.AppConfig.Port
	log.Println("Server started on port " + port + ".")
	router.Run(":" + port)
}
