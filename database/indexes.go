// database/indexes.go

package database

import (
	"context"
	"log"
	"time"

	"github.com/seojoonrp/bapddang-server/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(client *mongo.Client) {
	db := client.Database(config.AppConfig.DBName)

	initUserIndexes(db.Collection("users"))
	initStandardFoodIndexes(db.Collection("standard_foods"))
	initCustomFoodIndexes(db.Collection("custom_foods"))
	initReviewIndexes(db.Collection("reviews"))
	initLikeIndexes(db.Collection("likes"))
	initRecHistoryIndexes(db.Collection("recommendation_histories"))
}

func initUserIndexes(coll *mongo.Collection) {
	createIndex(coll, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_unique_username"),
	})
}

func initStandardFoodIndexes(coll *mongo.Collection) {
	createIndex(coll, mongo.IndexModel{
		Keys:    bson.D{{Key: "speed", Value: 1}},
		Options: options.Index().SetName("idx_food_speed"),
	})

	createIndex(coll, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_unique_food_name"),
	})
}

func initCustomFoodIndexes(coll *mongo.Collection) {
	createIndex(coll, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetName("idx_custom_food_name"),
	})
}

func initReviewIndexes(coll *mongo.Collection) {
	createIndex(coll, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "day", Value: -1},
		},
		Options: options.Index().SetName("idx_user_day_review"),
	})
}

func initLikeIndexes(coll *mongo.Collection) {
	createIndex(coll, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "food_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("idx_unique_user_food_like"),
	})

	createIndex(coll, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetName("idx_like_user_id"),
	})
}

func initRecHistoryIndexes(coll *mongo.Collection) {
	createIndex(coll, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
		Options: options.Index().SetName("idx_user_createdat_rec_history"),
	})

	createIndex(coll, mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(60 * 60 * 24 * 7).SetName("idx_rec_history_ttl"),
	})
}

func createIndex(coll *mongo.Collection, model mongo.IndexModel) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name, err := coll.Indexes().CreateOne(ctx, model)
	if err != nil {
		log.Printf("Error while creating index on %s: %v", coll.Name(), err)
		return
	}
	log.Printf("Successfully applied index %s on collection %s", name, coll.Name())
}
