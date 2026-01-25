// api/handlers/marshmallow.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/response"
)

type MarshmallowHandler struct {
	marshmallowService services.MarshmallowService
}

func NewMarshmallowHandler(ms services.MarshmallowService) *MarshmallowHandler {
	return &MarshmallowHandler{
		marshmallowService: ms,
	}
}

// @Summary 유저 마시멜로 조회
// @Description 특정 유저의 마시멜로 정보를 시간순으로 정렬해 반환한다.
// @Tags Marshmallow
// @Produce json
// @Success 200 {object} response.Response{data=[]models.Marshmallow} "마시멜로 조회 성공"
// @Failure 401 {object} response.Response "인증 실패"
// @Security BearerAuth
// @Router /marshmallows [get]
func (h *MarshmallowHandler) GetUserMarshmallows(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	marshmallows, err := h.marshmallowService.GetUserMarshmallows(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    marshmallows,
	})
}
