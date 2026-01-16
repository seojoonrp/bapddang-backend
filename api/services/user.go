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
	Login(ctx context.Context, input models.LoginRequest) (string, *models.User, error)

	LoginWithGoogle(ctx context.Context, idToken string) (bool, string, *models.User, error)
	LoginWithKakao(ctx context.Context, accessToken string) (bool, string, *models.User, error)
	LoginWithApple(ctx context.Context, identityToken string) (bool, string, *models.User, error)

	SyncUserDay(ctx context.Context, userID string) error
}

type userService struct {
	userRepo repositories.UserRepository
	foodRepo repositories.FoodRepository
}

func NewUserService(userRepo repositories.UserRepository, foodRepo repositories.FoodRepository) UserService {
	return &userService{userRepo: userRepo, foodRepo: foodRepo}
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
		CreatedAt:   time.Now(),
	}

	err = s.userRepo.Create(ctx, newUser)
	if err != nil {
		return apperr.InternalServerError("failed to create user", err)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, req models.LoginRequest) (string, *models.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil || user == nil {
		return "", nil, apperr.Unauthorized("invalid username or password", nil)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, apperr.Unauthorized("invalid username or password", nil)
	}

	token, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		return "", nil, apperr.InternalServerError("failed to generate token", err)
	}

	return token, user, nil
}

func (s *userService) loginWithSocial(ctx context.Context, provider string, socialID string, email string) (bool, string, *models.User, error) {
	targetUsername := utils.GenerateHashUsername(provider, socialID)
	isNew := false

	user, err := s.userRepo.FindByUsername(ctx, targetUsername)
	if err != nil {
		return false, "", nil, apperr.InternalServerError("failed to fetch user", err)
	}

	if user == nil {
		isNew = true
		user = &models.User{
			ID:          primitive.NewObjectID(),
			Username:    targetUsername,
			SocialID:    socialID,
			LoginMethod: provider,
			Day:         1,
			CreatedAt:   time.Now(),
		}
		if email != "" {
			user.Email = email
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return false, "", nil, apperr.InternalServerError("failed to create user", err)
		}
	}

	signedToken, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		return false, "", nil, apperr.InternalServerError("failed to generate token", err)
	}

	return isNew, signedToken, user, nil
}

func (s *userService) LoginWithGoogle(ctx context.Context, idToken string) (bool, string, *models.User, error) {
	webClientID := config.AppConfig.GoogleWebClientID

	payload, err := idtoken.Validate(context.Background(), idToken, webClientID)
	if err != nil {
		return false, "", nil, apperr.Unauthorized("invalid Google ID token", err)
	}

	socialID := payload.Subject
	email, _ := payload.Claims["email"].(string)

	return s.loginWithSocial(ctx, models.LoginMethodGoogle, socialID, email)
}

func (s *userService) LoginWithKakao(ctx context.Context, accessToken string) (bool, string, *models.User, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return false, "", nil, apperr.ServiceUnavailable("kakao api server unreachable", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return false, "", nil, apperr.Unauthorized("expired or invalid kakao token", nil)
	} else if resp.StatusCode != http.StatusOK {
		return false, "", nil, apperr.InternalServerError("kakao api returned error status", fmt.Errorf("status: %d", resp.StatusCode))
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
		return false, "", nil, apperr.InternalServerError("failed to decode Kakao user info", err)
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

func (s *userService) LoginWithApple(ctx context.Context, identityToken string) (bool, string, *models.User, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := s.verifyAppleToken(identityToken, clientID)
	if err != nil {
		return false, "", nil, err
	}

	socialID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	return s.loginWithSocial(ctx, models.LoginMethodApple, socialID, email)
}

func (s *userService) SyncUserDay(ctx context.Context, userID string) error {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return apperr.InternalServerError("invalid user ID in token", err)
	}

	user, err := s.userRepo.FindByID(ctx, uID)
	if err != nil {
		return apperr.InternalServerError("failed to fetch user", err)
	}
	if user == nil {
		return apperr.NotFound("user not found", nil)
	}

	seoulLoc, _ := time.LoadLocation("Asia/Seoul")
	nowKST := time.Now().In(seoulLoc)
	createdAtKST := user.CreatedAt.In(seoulLoc)

	todayZero := time.Date(nowKST.Year(), nowKST.Month(), nowKST.Day(), 0, 0, 0, 0, seoulLoc)
	startZero := time.Date(createdAtKST.Year(), createdAtKST.Month(), createdAtKST.Day(), 0, 0, 0, 0, seoulLoc)

	calculatedDay := int(todayZero.Sub(startZero).Hours()/24) + 1

	if user.Day < calculatedDay {
		err := s.userRepo.UpdateDay(ctx, uID, calculatedDay)
		if err != nil {
			return apperr.InternalServerError("failed to update user day", err)
		}
		user.Day = calculatedDay
	}

	log.Printf("Successfully update user %s's day to %d", user.Username, calculatedDay)
	return nil
}
