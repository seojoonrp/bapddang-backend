// models/like.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Like struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	FoodID    primitive.ObjectID `bson:"food_id"`
	CreatedAt time.Time          `bson:"created_at"`
}
