package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

// @securityDefinitions.http BearerAuth
// @scheme bearer
// @bearerFormat JWT
// @description "Bearer {token}" 형식으로 JWT 토큰을 전달
func main() {
	config.LoadConfig()

	if config.AppConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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

	userService := services.NewUserService(userRepository, foodRepository, reviewRepository, likeRepository, marshmallowRepository, recHistoryRepository)
	foodService := services.NewFoodService(context.Background(), foodRepository, likeRepository, recHistoryRepository)
	reviewService := services.NewReviewService(reviewRepository, foodRepository, userRepository, marshmallowRepository)
	likeService := services.NewLikeService(likeRepository, foodRepository)
	marshmallowService := services.NewMarshmallowService(marshmallowRepository)

	userHandler := handlers.NewUserHandler(userService, foodService)
	foodHandler := handlers.NewFoodHandler(foodService)
	reviewHandler := handlers.NewReviewHandler(reviewService, foodService)
	likeHandler := handlers.NewLikeHandler(likeService)
	marshmallowHandler := handlers.NewMarshmallowHandler(marshmallowService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://api.bapddang.com", "https://bapddang.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(middleware.ErrorHandler())
	router.SetTrustedProxies(nil)

	routes.SetupRoutes(
		router,
		db,
		userHandler,
		foodHandler,
		reviewHandler,
		likeHandler,
		marshmallowHandler,
	)

	port := config.AppConfig.Port
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Println("Server started on port " + port + ".")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server error: ", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("Server forced to shutdown: ", err)
	}
	log.Println("Server exited.")
}
