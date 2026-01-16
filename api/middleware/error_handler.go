// api/middleware/error_handler.go

package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/response"
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
				c.JSON(appErr.StatusCode, response.Response{
					Success: false,
					Error: &response.ErrorDetail{
						Code:    appErr.Code,
						Message: appErr.Message,
					},
				})
			} else {
				fmt.Printf("[UNKNOWN ERROR] %v\n", err)
				c.JSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error: &response.ErrorDetail{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "An unexpected error occurred.",
					},
				})
			}
		}
	}
}
