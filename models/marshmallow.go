// models/marshmallow.go

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Marshmallow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userID"`
	Week        int                `bson:"week" json:"week"`
	ReviewCount int                `bson:"review_count" json:"reviewCount"`
	TotalRating int                `bson:"total_rating" json:"totalRating"`
	Status      int                `bson:"status" json:"status"`
	IsComplete  bool               `bson:"is_complete" json:"isComplete"`
}
