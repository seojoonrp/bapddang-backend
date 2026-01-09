// middleware/auth.go
// JWT 인증 미들웨어

package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/config"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Error(apperr.Unauthorized("authorization header is required", nil))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})
		if err != nil {
			c.Error(apperr.Unauthorized("invalid or expired token", err))
			c.Abort()
			return
		}
		if !token.Valid {
			c.Error(apperr.Unauthorized("invalid token", nil))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Error(apperr.Unauthorized("invalid token claims", nil))
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.Error(apperr.Unauthorized("invalid user ID in token", nil))
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
