// api/handlers/user.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"github.com/seojoonrp/bapddang-server/response"
)

type UserHandler struct {
	userService services.UserService
	foodService services.FoodService
}

func NewUserHandler(userService services.UserService, foodService services.FoodService) *UserHandler {
	return &UserHandler{
		userService: userService,
		foodService: foodService,
	}
}

// @Summary 아이디 중복 확인
// @Description 입력한 아이디(Username)가 이미 존재하는지 확인한다.
// @Tags Auth
// @Accept json
// @Produce json
// @Param username query string true "확인할 아이디"
// @Success 200 {object} response.Response{data=bool} "중복 여부 (true: 존재함)"
// @Router /auth/check-username [get]
func (h *UserHandler) CheckUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.Error(apperr.BadRequest("username query parameter is required", nil))
		return
	}

	exists, err := h.userService.CheckUsernameExists(c, username)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    exists,
	})
}

// @Summary 일반 회원가입
// @Description 아이디와 비밀번호로 새로운 유저를 등록한다.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.SignUpRequest true "회원가입 정보"
// @Success 201 {object} response.Response{data=string} "성공 메시지"
// @Failure 409 {object} response.Response "이미 존재하는 아이디"
// @Router /auth/signup [post]
func (h *UserHandler) SignUp(c *gin.Context) {
	var req models.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	err := h.userService.SignUp(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    "user created successfully",
	})
}

// @Summary 일반 로그인
// @Description 아이디와 비밀번호로 로그인한다.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "로그인 정보"
// @Success 200 {object} response.Response{data=models.LoginResponse} "로그인 성공 정보"
// @Failure 401 {object} response.Response "아이디 또는 비밀번호 불일치"
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.userService.Login(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    result,
	})
}

// @Summary 구글 로그인
// @Description 구글 ID 토큰으로 로그인한다.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.GoogleLoginRequest true "구글 ID 토큰"
// @Success 200 {object} response.Response{data=models.LoginResponse} "로그인 성공 정보"
// @Router /auth/google [post]
func (h *UserHandler) GoogleLogin(c *gin.Context) {
	var req models.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	result, err := h.userService.LoginWithGoogle(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    result,
	})
}

// @Summary 카카오 로그인
// @Description 카카오 액세스 토큰으로 로그인한다.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.KakaoLoginRequest true "카카오 액세스 토큰"
// @Success 200 {object} response.Response{data=models.LoginResponse} "로그인 성공 정보"
// @Router /auth/kakao [post]
func (h *UserHandler) KakaoLogin(c *gin.Context) {
	var req models.KakaoLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	result, err := h.userService.LoginWithKakao(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    result,
	})
}

// @Summary 애플 로그인
// @Description 애플 아이덴티티 토큰으로 로그인한다.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.AppleLoginRequest true "애플 아이덴티티 토큰 및 전체 이름"
// @Success 200 {object} response.Response{data=models.LoginResponse} "로그인 성공 정보"
// @Router /auth/apple [post]
func (h *UserHandler) AppleLogin(c *gin.Context) {
	var req models.AppleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest("invalid request body", err))
		return
	}

	result, err := h.userService.LoginWithApple(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    result,
	})
}

// @Summary 내 정보 조회
// @Description 현재 로그인한 유저 정보를 가져온다.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=models.User} "현재 유저 정보"
// @Security BearerAuth
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	user, err := h.userService.GetUserByID(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    user,
	})
}

// @Summary 회원 탈퇴
// @Description 현재 로그인한 유저의 계정을 삭제한다.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=string} "성공 메시지"
// @Security BearerAuth
// @Router /users/me [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = h.userService.Withdraw(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    "user deleted successfully",
	})
}

// @Summary Day 동기화
// @Description 가입일로부터의 경과 일수를 바탕으로 유저의 Day와 Week을 업데이트하고, 마시멜로 상태를 동기화한다.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=bool} "Week 변경 여부 (true: 변경됨)"
// @Security BearerAuth
// @Router /users/me/sync [patch]
func (h *UserHandler) SyncUserDayAndWeek(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	weekUpdated, err := h.userService.SyncUserDay(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    weekUpdated,
	})
}
