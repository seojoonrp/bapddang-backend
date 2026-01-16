// models/rec_history_model.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecHistory struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID   `bson:"user_id" json:"userID"`
	FoodIDs   []primitive.ObjectID `bson:"food_ids" json:"foodIDs"`
	CreatedAt time.Time            `bson:"created_at" json:"createdAt"`
}
