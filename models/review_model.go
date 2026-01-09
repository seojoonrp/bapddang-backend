// models/review.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewedFoodItem struct {
	FoodID   string `bson:"food_id" json:"foodID"`
	FoodType string `bson:"food_type" json:"foodType"`
}

type Review struct {
	ID     primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id" json:"userID"`

	Name     string             `bson:"name" json:"name"`
	Foods    []ReviewedFoodItem `bson:"foods" json:"foods"`
	Speed    string             `bson:"speed" json:"speed"`
	MealTime string             `bson:"meal_time" json:"mealTime"`

	Tags     []string `bson:"tags" json:"tags"`
	ImageURL string   `bson:"image_url" json:"imageUrl"`
	Comment  string   `bson:"comment" json:"comment"`
	Rating   int      `bson:"rating" json:"rating"`

	Day       int       `bson:"day" json:"day"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

type CreateReviewRequest struct {
	Name     string             `json:"name" binding:"required"`
	Foods    []ReviewedFoodItem `json:"foods" binding:"required"`
	Speed    string             `json:"speed" binding:"required"`
	MealTime string             `json:"mealTime" binding:"required"`
	Tags     []string           `json:"tags"`
	ImageURL string             `json:"imageURL"`
	Comment  string             `json:"comment"`
	Rating   int                `json:"rating"`
}

type UpdateReviewRequest struct {
	MealTime string    `json:"mealTime" binding:"required"`
	Tags     *[]string `json:"tags,omitempty"`
	ImageURL *string   `json:"imageURL,omitempty"`
	Comment  *string   `json:"comment,omitempty"`
	Rating   *int      `json:"rating,omitempty"`
}
