// middleware/auth.go
// JWT 인증 미들웨어

package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var seoulLoc *time.Location

func init() {
	var err error
	seoulLoc, err = time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("WARNING: Failed to load Asia/Seoul locatio, using UTC: %v", err)
		seoulLoc = time.UTC
	}
}

func AuthMiddleware(userCollection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format not matched."})
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userIDHex, ok := claims["sub"].(string)
			if !ok {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				return
			}

			userID, err := primitive.ObjectIDFromHex(userIDHex)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				return
			}

			var user models.User
			err = userCollection.FindOne(ctx, primitive.M{"_id": userID}).Decode(&user)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}

			calculatedDay := calculateCurrentDay(user.CreatedAt)

			if user.Day < calculatedDay {
				user.Day = calculatedDay
				go updateUserDay(user.ID, calculatedDay, userCollection)
			}

			ctx.Set("currentUser", user)

			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}

func updateUserDay(userID primitive.ObjectID, newDay int, userCollection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"day": newDay}}

	_, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update user day: %v", err)
	}
}

func calculateCurrentDay(createdAt time.Time) int {
	nowKST := time.Now().In(seoulLoc)
	createdAtKST := createdAt.In(seoulLoc)

	endDate := time.Date(nowKST.Year(), nowKST.Month(), nowKST.Day(), 0, 0, 0, 0, seoulLoc)
	startDate := time.Date(createdAtKST.Year(), createdAtKST.Month(), createdAtKST.Day(), 0, 0, 0, 0, seoulLoc)

	daysPassed := int(endDate.Sub(startDate).Hours() / 24)
	return daysPassed + 1
}
