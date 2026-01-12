// models/food.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SpeedFast = "fast"
	SpeedSlow = "slow"
)

type StandardFood struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" binding:"required"`
	ImageURL    string             `bson:"image_url" json:"imageURL" binding:"required"`
	Speed       string             `bson:"speed" json:"speed" binding:"required"`
	Parents     []string           `bson:"parents" json:"parents"`
	Categories  []string           `bson:"categories" json:"categories"`
	LikeCount   int                `bson:"like_count" json:"likeCount"`
	ReviewCount int                `bson:"review_count" json:"reviewCount"`
	TotalRating int                `bson:"total_rating" json:"totalRating"`
}

type CreateStandardFoodRequest struct {
	Name       string   `json:"name" binding:"required"`
	ImageURL   string   `json:"imageURL" binding:"required"`
	Speed      string   `json:"speed" binding:"required"`
	Parents    []string `json:"parents" binding:"required"`
	Categories []string `json:"categories" binding:"required"`
}

type CustomFood struct {
	ID          primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" binding:"required"`
	ReviewCount int                `bson:"review_count" json:"reviewCount"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}

type ResolveFoodItemsRequest struct {
	Names []string `json:"names" binding:"required"`
}
