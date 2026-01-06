// api/services/user_service.go

package services

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seojoonrp/bapddang-server/api/repositories"
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
	SignUp(ctx context.Context, input models.SignUpInput) (*models.User, error)
	Login(ctx context.Context, input models.LoginInput) (string, *models.User, error)
	LoginWithGoogle(ctx context.Context, idToken string) (bool, string, *models.User, error)
	LoginWithKakao(ctx context.Context, accessToken string) (bool, string, *models.User, error)
	LoginWithApple(ctx context.Context, identityToken string) (bool, string, *models.User, error)

	LikeFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error)
	UnlikeFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error)
	GetLikedFoodIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
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
		return false, err
	}
	if user != nil {
		return true, nil
	}
	return false, nil
}

func (s *userService) SignUp(ctx context.Context, input models.SignUpInput) (*models.User, error) {
	runes := []rune(input.Username)
	if len(runes) < 3 || len(runes) > 15 {
		return nil, errors.New("username must be between 3 and 15 characters")
	}

	exists, err := s.CheckUsernameExists(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:           primitive.NewObjectID(),
		Username:     input.Username,
		Password:     string(hashedPassword),
		LoginMethod:  models.LoginMethodEmail,
		Day:          1,
		LikedFoodIDs: make([]primitive.ObjectID, 0),
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.Save(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) Login(ctx context.Context, input models.LoginInput) (string, *models.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", nil, err // Invalid username
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", nil, err // Invalid password
	}

	token, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *userService) loginWithSocial(ctx context.Context, provider string, socialID string, email string) (bool, string, *models.User, error) {
	targetUsername := utils.GenerateHashUsername(provider, socialID)
	isNew := false

	user, err := s.userRepo.FindByUsername(ctx, targetUsername)
	if err != nil {
		return false, "", nil, err
	}

	if user == nil {
		isNew = true
		user = &models.User{
			ID:           primitive.NewObjectID(),
			Username:     targetUsername,
			SocialID:     socialID,
			LoginMethod:  provider,
			Day:          1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt:    time.Now(),
		}
		if email != "" {
			user.Email = email
		}

		if err := s.userRepo.Save(ctx, user); err != nil {
			return false, "", nil, err
		}
	}

	signedToken, err := utils.GenerateToken(user.ID.Hex())
	return isNew, signedToken, user, err
}

func (s *userService) LoginWithGoogle(ctx context.Context, idToken string) (bool, string, *models.User, error) {
	webClientID := config.AppConfig.GoogleWebClientID

	payload, err := idtoken.Validate(ctx, idToken, webClientID)
	if err != nil {
		return false, "", nil, errors.New("invalid Google ID token")
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
	if err != nil || resp.StatusCode != http.StatusOK {
		return false, "", nil, errors.New("invalid Kakao access token")
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
		return false, "", nil, errors.New("failed to decode Kakao response")
	}

	socialID := strconv.FormatInt(kakaoRes.ID, 10)
	email := kakaoRes.KakaoAccount.Email

	return s.loginWithSocial(ctx, models.LoginMethodKakao, socialID, email)
}

func (s *userService) verifyAppleToken(identityToken string, clientID string) (jwt.MapClaims, error) {
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var appleKeys AppleKeys
	json.NewDecoder(resp.Body).Decode(&appleKeys)

	token, err := jwt.Parse(identityToken, func(token *jwt.Token) (interface{}, error) {
		kid := token.Header["kid"].(string)
		for _, key := range appleKeys.Keys {
			if key.Kid == kid {
				nBytes, _ := base64.RawURLEncoding.DecodeString(key.N)
				eBytes, _ := base64.RawURLEncoding.DecodeString(key.E)
				var e int
				for _, b := range eBytes {
					e = e<<8 + int(b)
				}
				pubKey := &rsa.PublicKey{
					N: new(big.Int).SetBytes(nBytes),
					E: e,
				}
				return pubKey, nil
			}
		}
		return nil, errors.New("public key not found")
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid Apple identity token")
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer")
	}
	if claims["aud"] != clientID {
		return nil, errors.New("invalid audience")
	}

	fmt.Println("Successfully parsed apple token.")
	return claims, nil
}

func (s *userService) LoginWithApple(ctx context.Context, identityToken string) (bool, string, *models.User, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := s.verifyAppleToken(identityToken, clientID)
	if err != nil {
		fmt.Println("Error while verifying id token:", err)
		return false, "", nil, err
	}

	socialID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	return s.loginWithSocial(ctx, models.LoginMethodApple, socialID, email)
}

func (s *userService) LikeFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error) {
	wasAdded, err := s.userRepo.AddLikedFood(ctx, userID, foodID)
	if err != nil {
		return false, err
	}

	return wasAdded, nil
}

func (s *userService) UnlikeFood(ctx context.Context, userID, foodID primitive.ObjectID) (bool, error) {
	wasRemoved, err := s.userRepo.RemoveLikedFood(ctx, userID, foodID)
	if err != nil {
		return false, err
	}

	return wasRemoved, nil
}

func (s *userService) GetLikedFoodIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	return s.userRepo.GetLikedFoodIDs(ctx, userID)
}
