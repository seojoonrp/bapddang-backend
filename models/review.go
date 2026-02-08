// models/review.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Image URL, Comment는 없으면 그냥 빈 문자열
type Review struct {
	ID        primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"userID"`
	Name      string             `bson:"name" json:"name"`
	Foods     []ReviewFoodItem   `bson:"foods" json:"foods"`
	MealTime  string             `bson:"meal_time" json:"mealTime"`
	ImageURL  string             `bson:"image_url" json:"imageURL"`
	Comment   string             `bson:"comment" json:"comment"`
	Rating    int                `bson:"rating" json:"rating"`
	Day       int                `bson:"day" json:"day"`
	Week      int                `bson:"week" json:"week"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

type CreateReviewRequest struct {
	Name     string           `json:"name" binding:"required"`
	Foods    []ReviewFoodItem `json:"foods" binding:"required"`
	MealTime string           `json:"mealTime" binding:"required"`
	ImageURL string           `json:"imageURL"`
	Comment  string           `json:"comment"`
	Rating   int              `json:"rating" binding:"required"`
	Day      int              `json:"day" binding:"required"`
}

type UpdateReviewRequest struct {
	MealTime string `json:"mealTime" binding:"required"`
	ImageURL string `json:"imageURL"`
	Comment  string `json:"comment"`
	Rating   int    `json:"rating" binding:"required"`
}

type RecentReviewResponse struct {
	Comment   string       `json:"comment"`
	Rating    int          `json:"rating"`
	CreatedAt time.Time    `json:"createdAt"`
	Food      StandardFood `json:"food"`
}
