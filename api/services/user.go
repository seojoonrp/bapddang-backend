// api/services/user.go

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"github.com/seojoonrp/bapddang-server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type AppleKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type AppleKeys struct {
	Keys []AppleKey `json:"keys"`
}

type UserService interface {
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	SignUp(ctx context.Context, req models.SignUpRequest) error
	Login(ctx context.Context, input models.LoginRequest) (models.LoginResponse, error)

	LoginWithGoogle(ctx context.Context, idToken string) (models.LoginResponse, error)
	LoginWithKakao(ctx context.Context, accessToken string) (models.LoginResponse, error)
	LoginWithApple(ctx context.Context, identityToken string) (models.LoginResponse, error)

	SyncUserDay(ctx context.Context, userID string) (bool, error)
}

type userService struct {
	userRepo        repositories.UserRepository
	foodRepo        repositories.FoodRepository
	marshmallowRepo repositories.MarshmallowRepository
}

func NewUserService(ur repositories.UserRepository, fr repositories.FoodRepository, mr repositories.MarshmallowRepository) UserService {
	return &userService{userRepo: ur, foodRepo: fr, marshmallowRepo: mr}
}

func (s *userService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return false, apperr.InternalServerError("failed to fetch user", err)
	}
	if user != nil {
		return true, nil
	}
	return false, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	user, err := s.userRepo.FindByID(ctx, uID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch user", err)
	}
	if user == nil {
		return nil, apperr.NotFound("user not found", nil)
	}

	return user, nil
}

func (s *userService) SignUp(ctx context.Context, req models.SignUpRequest) error {
	runes := []rune(req.Username)
	if len(runes) < 3 || len(runes) > 15 {
		return apperr.BadRequest("username must be between 3 and 15 characters", nil)
	}

	exists, err := s.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		return apperr.InternalServerError("failed to fetch user", err)
	}
	if exists {
		return apperr.Conflict("user already exists", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apperr.InternalServerError("failed to hash password", err)
	}

	newUser := &models.User{
		ID:          primitive.NewObjectID(),
		Username:    req.Username,
		Password:    string(hashedPassword),
		LoginMethod: models.LoginMethodLocal,
		Day:         1,
		Week:        1,
		CreatedAt:   time.Now(),
	}

	err = s.userRepo.Create(ctx, newUser)
	if err != nil {
		return apperr.InternalServerError("failed to create user", err)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil || user == nil {
		return models.LoginResponse{}, apperr.Unauthorized("invalid username or password", nil)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return models.LoginResponse{}, apperr.Unauthorized("invalid username or password", nil)
	}

	token, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		return models.LoginResponse{}, apperr.InternalServerError("failed to generate token", err)
	}

	return models.LoginResponse{
		AccessToken: token,
		User:        user,
		IsNewUser:   false,
	}, nil
}

func (s *userService) loginWithSocial(ctx context.Context, provider string, socialID string, email string) (models.LoginResponse, error) {
	targetUsername := utils.GenerateHashUsername(provider, socialID)
	isNew := false

	user, err := s.userRepo.FindByUsername(ctx, targetUsername)
	if err != nil {
		return models.LoginResponse{}, apperr.InternalServerError("failed to fetch user", err)
	}

	if user == nil {
		isNew = true
		user = &models.User{
			ID:          primitive.NewObjectID(),
			Username:    targetUsername,
			SocialID:    socialID,
			LoginMethod: provider,
			Day:         1,
			Week:        1,
			CreatedAt:   time.Now(),
		}
		if email != "" {
			user.Email = email
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return models.LoginResponse{}, apperr.InternalServerError("failed to create user", err)
		}
	}

	signedToken, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		return models.LoginResponse{}, apperr.InternalServerError("failed to generate token", err)
	}

	return models.LoginResponse{
		AccessToken: signedToken,
		User:        user,
		IsNewUser:   isNew,
	}, nil
}

func (s *userService) LoginWithGoogle(ctx context.Context, idToken string) (models.LoginResponse, error) {
	webClientID := config.AppConfig.GoogleWebClientID

	payload, err := idtoken.Validate(context.Background(), idToken, webClientID)
	if err != nil {
		return models.LoginResponse{}, apperr.Unauthorized("invalid Google ID token", err)
	}

	socialID := payload.Subject
	email, _ := payload.Claims["email"].(string)

	return s.loginWithSocial(ctx, models.LoginMethodGoogle, socialID, email)
}

func (s *userService) LoginWithKakao(ctx context.Context, accessToken string) (models.LoginResponse, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return models.LoginResponse{}, apperr.ServiceUnavailable("kakao api server unreachable", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return models.LoginResponse{}, apperr.Unauthorized("expired or invalid kakao token", nil)
	} else if resp.StatusCode != http.StatusOK {
		return models.LoginResponse{}, apperr.InternalServerError("kakao api returned error status", fmt.Errorf("status: %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	var kakaoRes struct {
		ID           int64 `json:"id"`
		KakaoAccount struct {
			Email   string `json:"email"`
			Profile struct {
				Nickname string `json:"nickname"`
			} `json:"profile"`
		} `json:"kakao_account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&kakaoRes); err != nil {
		return models.LoginResponse{}, apperr.InternalServerError("failed to decode Kakao user info", err)
	}

	socialID := strconv.FormatInt(kakaoRes.ID, 10)
	email := kakaoRes.KakaoAccount.Email

	return s.loginWithSocial(ctx, models.LoginMethodKakao, socialID, email)
}

func (s *userService) verifyAppleToken(identityToken string, clientID string) (jwt.MapClaims, error) {
	appleJWKSURL := "https://appleid.apple.com/auth/keys"

	k, err := keyfunc.NewDefault([]string{appleJWKSURL})
	if err != nil {
		return nil, apperr.InternalServerError("failed to create keyfunc", err)
	}

	token, err := jwt.Parse(identityToken, k.Keyfunc)
	if err != nil {
		return nil, apperr.InternalServerError("invalid token", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] != "https://appleid.apple.com" {
			return nil, apperr.Unauthorized("invalid issuer", nil)
		}
		if claims["aud"] != clientID {
			return nil, apperr.Unauthorized("invalid audience", nil)
		}
		return claims, nil
	}

	return nil, apperr.Unauthorized("invalid token claims", nil)
}

func (s *userService) LoginWithApple(ctx context.Context, identityToken string) (models.LoginResponse, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := s.verifyAppleToken(identityToken, clientID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	socialID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	return s.loginWithSocial(ctx, models.LoginMethodApple, socialID, email)
}

func (s *userService) SyncUserDay(ctx context.Context, userID string) (bool, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, apperr.InternalServerError("invalid user ID in token", err)
	}

	user, err := s.userRepo.FindByID(ctx, uID)
	if err != nil {
		return false, apperr.InternalServerError("failed to fetch user", err)
	}
	if user == nil {
		return false, apperr.NotFound("user not found", nil)
	}

	seoulLoc, _ := time.LoadLocation("Asia/Seoul")
	nowKST := time.Now().In(seoulLoc)
	createdAtKST := user.CreatedAt.In(seoulLoc)

	todayZero := time.Date(nowKST.Year(), nowKST.Month(), nowKST.Day(), 0, 0, 0, 0, seoulLoc)
	startZero := time.Date(createdAtKST.Year(), createdAtKST.Month(), createdAtKST.Day(), 0, 0, 0, 0, seoulLoc)

	calculatedDay := int(todayZero.Sub(startZero).Hours()/24) + 1
	calculatedWeek := (calculatedDay-1)/7 + 1
	weekUpdated := false

	curM, err := s.marshmallowRepo.FindByUserIDAndWeek(ctx, uID, calculatedWeek)
	if err != nil {
		return false, apperr.InternalServerError("failed to fetch marshmallow", err)
	}
	if curM == nil {
		newM := models.Marshmallow{
			ID:          primitive.NewObjectID(),
			UserID:      user.ID,
			Week:        calculatedWeek,
			ReviewCount: 0,
			TotalRating: 0,
			Status:      -1,
			IsComplete:  false,
		}
		err := s.marshmallowRepo.Create(ctx, newM)
		if err != nil {
			return false, apperr.InternalServerError("failed to create marshmallow", err)
		}
	}

	if user.Day < calculatedDay {
		if user.Week < calculatedWeek {
			weekUpdated = true

			for week := user.Week; week < calculatedWeek; week++ {
				m, _ := s.marshmallowRepo.FindByUserIDAndWeek(ctx, uID, week)
				if m == nil { // 1주간 접속 기록이 아예 없어서 생성조차 안된 경우
					emptyMarshmallow := models.Marshmallow{
						ID:          primitive.NewObjectID(),
						UserID:      user.ID,
						Week:        week,
						ReviewCount: 0,
						TotalRating: 0,
						Status:      -1,
						IsComplete:  true,
					}
					err := s.marshmallowRepo.Create(ctx, emptyMarshmallow)
					if err != nil {
						log.Printf("[WARNING] Failed to create empty marshmallow for user %s week %d: %v", user.Username, week, err)
					}
				} else if !m.IsComplete { // 존재는 하는데 완료처리가 안된 경우
					finalStatus := utils.GetMarshmallowStatus(m.ReviewCount, m.TotalRating)
					err := s.marshmallowRepo.CompleteMarshmallow(ctx, m.ID, finalStatus)
					if err != nil {
						return false, apperr.InternalServerError("failed to complete marshmallow", err)
					}
				}
			}
		}

		err := s.userRepo.UpdateDayAndWeek(ctx, uID, calculatedDay, calculatedWeek)
		if err != nil {
			return false, apperr.InternalServerError("failed to update user day/week", err)
		}
		log.Printf("Successfully updated user %s's day to %d and week to %d", user.Username, calculatedDay, calculatedWeek)
	}

	return weekUpdated, nil
}
