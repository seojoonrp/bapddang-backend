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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username    string             `bson:"username" json:"username"`
	SocialID    string             `bson:"social_id,omitempty" json:"-"`
	Password    string             `bson:"password,omitempty" json:"-"`
	Email       string             `bson:"email,omitempty" json:"email"`
	LoginMethod string             `bson:"login_method" json:"loginMethod"`
	Day         int                `bson:"day" json:"day"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}

type SignUpRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
