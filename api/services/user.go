// api/services/user.go

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error)

	LoginWithGoogle(ctx context.Context, req models.GoogleLoginRequest) (models.LoginResponse, error)
	LoginWithKakao(ctx context.Context, req models.KakaoLoginRequest) (models.LoginResponse, error)
	LoginWithApple(ctx context.Context, req models.AppleLoginRequest) (models.LoginResponse, error)
	AgreeTerms(ctx context.Context, userID string) error

	Withdraw(ctx context.Context, userID string) error

	SyncUserDay(ctx context.Context, userID string) (models.SyncDayResponse, error)
}

type userService struct {
	userRepo        repositories.UserRepository
	foodRepo        repositories.FoodRepository
	reviewRepo      repositories.ReviewRepository
	likeRepo        repositories.LikeRepository
	marshmallowRepo repositories.MarshmallowRepository
	recHistoryRepo  repositories.RecHistoryRepository
}

func NewUserService(
	ur repositories.UserRepository,
	fr repositories.FoodRepository,
	rr repositories.ReviewRepository,
	lr repositories.LikeRepository,
	mr repositories.MarshmallowRepository,
	rhr repositories.RecHistoryRepository,
) UserService {
	return &userService{userRepo: ur, foodRepo: fr, reviewRepo: rr, likeRepo: lr, marshmallowRepo: mr, recHistoryRepo: rhr}
}

func (s *userService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	runes := []rune(username)
	if len(runes) < 3 || len(runes) > 15 {
		return false, apperr.BadRequest("username must be between 3 and 15 characters", nil)
	}

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
		IsAgreed:    true,
		AgreedAt:    time.Now(),
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
			IsAgreed:    false,
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

func (s *userService) LoginWithGoogle(ctx context.Context, req models.GoogleLoginRequest) (models.LoginResponse, error) {
	webClientID := config.AppConfig.GoogleWebClientID

	payload, err := idtoken.Validate(context.Background(), req.IDToken, webClientID)
	if err != nil {
		return models.LoginResponse{}, apperr.Unauthorized("invalid Google ID token", err)
	}

	socialID := payload.Subject
	email, _ := payload.Claims["email"].(string)

	return s.loginWithSocial(ctx, models.LoginMethodGoogle, socialID, email)
}

func (s *userService) LoginWithKakao(ctx context.Context, req models.KakaoLoginRequest) (models.LoginResponse, error) {
	client := &http.Client{}
	httpReq, _ := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)

	resp, err := client.Do(httpReq)
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

func (s *userService) LoginWithApple(ctx context.Context, req models.AppleLoginRequest) (models.LoginResponse, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := utils.VerifyAppleToken(req.IdentityToken, clientID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	socialID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	res, err := s.loginWithSocial(ctx, models.LoginMethodApple, socialID, email)
	if err != nil {
		return models.LoginResponse{}, err
	}

	if req.AuthorizationCode != "" {
		refreshToken, err := utils.GetAppleRefreshToken(req.AuthorizationCode)
		if err == nil {
			err = s.userRepo.UpdateAppleRefreshToken(ctx, res.User.ID, refreshToken)
			if err != nil {
				log.Println("[WARNING] Failed to update apple refresh token:", err)
			}
			res.User.AppleRefreshToken = refreshToken
		}
	}

	return res, nil
}

func (s *userService) AgreeTerms(ctx context.Context, userID string) error {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return apperr.InternalServerError("invalid user ID in token", err)
	}

	err = s.userRepo.UpdateAgreement(ctx, uID, true, time.Now())
	if err != nil {
		return apperr.InternalServerError("failed to update user agreement", err)
	}

	return nil
}

func (s *userService) Withdraw(ctx context.Context, userID string) error {
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

	likedFoodIDs, err := s.likeRepo.FindFoodIDsByUserID(ctx, uID)
	if err != nil {
		return apperr.InternalServerError("failed to fetch liked foods", err)
	}

	reviews, err := s.reviewRepo.FindAllByUserID(ctx, uID)
	if err != nil {
		return apperr.InternalServerError("failed to fetch user reviews", err)
	}

	for _, fID := range likedFoodIDs {
		err := s.foodRepo.DecrementLikeCount(ctx, fID)
		if err != nil {
			log.Println("[WARNING] Failed to decrement like count while withdrawing user:", err)
			continue
		}
	}

	for _, review := range reviews {
		if review.UserID != uID {
			log.Println("[WARNING] Review user ID does not match while withdrawing user")
			continue
		}

		var standardFoodIDs []primitive.ObjectID
		var customFoodIDs []primitive.ObjectID

		for _, foodItem := range review.Foods {
			foodID, err := primitive.ObjectIDFromHex(foodItem.FoodID)
			if err != nil {
				continue
			}
			if foodItem.Type == models.FoodTypeStandard {
				standardFoodIDs = append(standardFoodIDs, foodID)
			}
			if foodItem.Type == models.FoodTypeCustom {
				customFoodIDs = append(customFoodIDs, foodID)
			}
		}

		if len(standardFoodIDs) > 0 {
			err = s.foodRepo.UpdateStandardDeletedReviewStats(ctx, standardFoodIDs, review.Rating)
			if err != nil {
				continue
			}
		}
		if len(customFoodIDs) > 0 {
			err = s.foodRepo.UpdateCustomDeletedReviewStats(ctx, customFoodIDs)
			if err != nil {
				continue
			}
		}
	}

	if user.LoginMethod != models.LoginMethodLocal {
		err = s.handleSocialUnlink(user)
		if err != nil {
			log.Println("[WARNING] Failed to unlink social account while withdrawing user:", err)
		}
	}

	err = s.reviewRepo.DeleteByUserID(ctx, uID)
	err = s.likeRepo.DeleteByUserID(ctx, uID)
	err = s.marshmallowRepo.DeleteByUserID(ctx, uID)
	err = s.recHistoryRepo.DeleteByUserID(ctx, uID)
	err = s.userRepo.Delete(ctx, uID)
	if err != nil {
		return apperr.InternalServerError("failed to delete user and related data", err)
	}

	return nil
}

func (s *userService) handleSocialUnlink(user *models.User) error {
	switch user.LoginMethod {
	case models.LoginMethodKakao:
		return s.unlinkKakao(user.SocialID)
	case models.LoginMethodApple:
		return s.unlinkApple(user.AppleRefreshToken)
	case models.LoginMethodGoogle:
		return nil // 일단 구글은 안해도 될듯
	default:
		return nil
	}
}

func (s *userService) unlinkKakao(socialID string) error {
	client := &http.Client{}
	data := url.Values{}
	data.Set("target_id_type", "user_id")
	data.Set("target_id", socialID)

	req, _ := http.NewRequest("POST", "https://kapi.kakao.com/v1/user/unlink", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "KakaoAK "+config.AppConfig.KakaoAdminKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kakao unlink failed with status: %d", resp.StatusCode)
	}
	return nil
}

func (s *userService) unlinkApple(refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("refresh token is missing")
	}

	clientSecret, err := utils.GenerateAppleClientSecret()
	if err != nil {
		return err
	}

	revokeURL := "https://appleid.apple.com/auth/revoke"

	client := &http.Client{}
	data := url.Values{}
	data.Set("client_id", config.AppConfig.AppleBundleID)
	data.Set("client_secret", clientSecret)
	data.Set("token", refreshToken)
	data.Set("token_type_hint", "refresh_token")

	resp, err := client.PostForm(revokeURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("apple unlink failed with status: %d", resp.StatusCode)
	}
	return nil
}

func (s *userService) SyncUserDay(ctx context.Context, userID string) (models.SyncDayResponse, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return models.SyncDayResponse{}, apperr.InternalServerError("invalid user ID in token", err)
	}

	user, err := s.userRepo.FindByID(ctx, uID)
	if err != nil {
		return models.SyncDayResponse{}, apperr.InternalServerError("failed to fetch user", err)
	}
	if user == nil {
		return models.SyncDayResponse{}, apperr.NotFound("user not found", nil)
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
		return models.SyncDayResponse{}, apperr.InternalServerError("failed to fetch current week's marshmallow", err)
	}
	if curM == nil {
		log.Println("No current week marshmallow found, creating new one")
		newM := models.Marshmallow{
			ID:          primitive.NewObjectID(),
			UserID:      user.ID,
			Week:        calculatedWeek,
			ReviewCount: 0,
			TotalRating: 0,
			Status:      -2, // 진행중
			IsComplete:  false,
		}
		err := s.marshmallowRepo.Create(ctx, newM)
		if err != nil {
			return models.SyncDayResponse{}, apperr.InternalServerError("failed to create marshmallow", err)
		}
	}

	var lastM *models.Marshmallow
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
						log.Printf("[WARNING] Failed to create empty marshmallow: %v\n", err)
					}
				} else if !m.IsComplete { // 존재는 하는데 완료처리가 안된 경우
					finalStatus := utils.GetMarshmallowStatus(m.ReviewCount, m.TotalRating)
					err := s.marshmallowRepo.CompleteMarshmallow(ctx, m.ID, finalStatus)
					if err != nil {
						return models.SyncDayResponse{}, apperr.InternalServerError("failed to complete marshmallow", err)
					}
					log.Println("Successfully completed marshmallow for week", week)
				}
			}

			lastM, err = s.marshmallowRepo.FindByUserIDAndWeek(ctx, uID, calculatedWeek-1)
			if err != nil {
				return models.SyncDayResponse{}, apperr.InternalServerError("failed to fetch last week's marshmallow", err)
			}
		}

		err := s.userRepo.UpdateDayAndWeek(ctx, uID, calculatedDay, calculatedWeek)
		if err != nil {
			return models.SyncDayResponse{}, apperr.InternalServerError("failed to update user day/week", err)
		}
	}

	user.Day = calculatedDay
	user.Week = calculatedWeek

	return models.SyncDayResponse{
		UpdatedUser:     user,
		IsNewWeek:       weekUpdated,
		LastMarshmallow: lastM,
	}, nil
}
