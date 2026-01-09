// api/middleware/error_handler.go

package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/apperr"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := err.(*apperr.AppError); ok {
				if appErr.Raw != nil {
					fmt.Printf("[ERROR] %v\n", appErr.Raw)
				}
				c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
			} else {
				fmt.Printf("[UNKNOWN ERROR] %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			}
		}
	}
}
