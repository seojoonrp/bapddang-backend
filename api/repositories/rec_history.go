// api/repositories/rec_history.go

package repositories

import (
	"context"
	"time"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecHistoryRepository interface {
	SaveHistory(ctx context.Context, history models.RecHistory) error
	GetRecentFoodIDsMap(ctx context.Context, userID primitive.ObjectID, days int) (map[primitive.ObjectID]time.Time, error)
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type recHistoryRepo struct {
	collection *mongo.Collection
}

func NewRecHistoryRepository(db *mongo.Database) RecHistoryRepository {
	return &recHistoryRepo{
		collection: db.Collection("recommendation_histories"),
	}
}

func (r *recHistoryRepo) SaveHistory(ctx context.Context, history models.RecHistory) error {
	_, err := r.collection.InsertOne(ctx, history)
	return err
}

func (r *recHistoryRepo) GetRecentFoodIDsMap(ctx context.Context, userID primitive.ObjectID, days int) (map[primitive.ObjectID]time.Time, error) {
	threshold := time.Now().AddDate(0, 0, -days)
	filter := primitive.M{
		"user_id":    userID,
		"created_at": primitive.M{"$gte": threshold},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	historyMap := make(map[primitive.ObjectID]time.Time)
	for cursor.Next(ctx) {
		var doc struct {
			FoodIDs   []primitive.ObjectID `bson:"food_ids"`
			CreatedAt time.Time            `bson:"created_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		for _, foodID := range doc.FoodIDs {
			if existing, ok := historyMap[foodID]; !ok || doc.CreatedAt.After(existing) {
				historyMap[foodID] = doc.CreatedAt
			}
		}
	}
	return historyMap, nil
}

func (r *recHistoryRepo) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, primitive.M{"user_id": userID})
	return err
}
