// api/handlers/base.go

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/apperr"
)

func GetUserID(c *gin.Context) (string, error) {
	uIDVal, exists := c.Get("user_id")
	if !exists {
		return "", apperr.Unauthorized("user session expired", nil)
	}

	uIDStr, ok := uIDVal.(string)
	if !ok {
		return "", apperr.InternalServerError("invalid context user ID", nil)
	}

	return uIDStr, nil
}
