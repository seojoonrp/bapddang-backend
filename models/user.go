// models/user.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	LoginMethodLocal  = "local"
	LoginMethodGoogle = "google"
	LoginMethodKakao  = "kakao"
	LoginMethodApple  = "apple"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username          string             `bson:"username" json:"username"`
	SocialID          string             `bson:"social_id,omitempty" json:"-"`
	Password          string             `bson:"password,omitempty" json:"-"`
	Email             string             `bson:"email,omitempty" json:"email"`
	LoginMethod       string             `bson:"login_method" json:"loginMethod"`
	Day               int                `bson:"day" json:"day"`
	Week              int                `bson:"week" json:"week"`
	IsAgreed          bool               `bson:"is_agreed" json:"isAgreed"`
	AgreedAt          time.Time          `bson:"agreed_at,omitempty" json:"agreedAt,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"createdAt"`
	AppleRefreshToken string             `bson:"apple_refresh_token,omitempty" json:"-"`
}

type SignUpRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type GoogleLoginRequest struct {
	IDToken string `json:"idToken" binding:"required"`
}

type KakaoLoginRequest struct {
	AccessToken string `json:"accessToken" binding:"required"`
}

type AppleLoginRequest struct {
	IdentityToken     string `json:"identityToken" binding:"required"`
	AuthorizationCode string `json:"authorizationCode"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	User        *User  `json:"user"`
	IsNewUser   bool   `json:"isNewUser"`
}

type SyncDayResponse struct {
	UpdatedUser     *User        `json:"updatedUser"`
	IsNewWeek       bool         `json:"isNewWeek"`
	LastMarshmallow *Marshmallow `json:"lastMarshmallow"`
}
