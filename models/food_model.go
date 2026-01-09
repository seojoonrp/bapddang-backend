// models/food.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StandardFood struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name" binding:"required"`
	ImageURL string             `bson:"image_url" json:"imageURL" binding:"required"`

	Speed      string   `bson:"speed" json:"speed" binding:"required"`
	Type       string   `bson:"type" json:"type" binding:"required"`
	Categories []string `bson:"categories" json:"categories"`

	LikeCount   int `bson:"like_count" json:"likeCount"`
	ReviewCount int `bson:"review_count" json:"reviewCount"`
	TotalRating int `bson:"total_rating" json:"totalRating"`
}

type CreateStandardFoodRequest struct {
	Name       string   `json:"name"`
	ImageURL   string   `json:"imageURL"`
	Speed      string   `json:"speed"`
	Type       string   `json:"type"`
	Categories []string `json:"categories"`
}

type CustomFood struct {
	ID           primitive.ObjectID   `bson:"_id, omitempty" json:"id"`
	Name         string               `bson:"name" json:"name" binding:"required"`
	UsingUserIDs []primitive.ObjectID `bson:"using_user_ids" json:"usingUserIDs" binding:"required"`
	CreatedAt    time.Time            `bson:"created_at" json:"createdAt"`
}

type CreateCustomFoodRequest struct {
	Name string `json:"name"`
}
