// database/database.go

package database

import (
	"context"
	"log"
	"time"

	"github.com/seojoonrp/bapddang-server/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() (*mongo.Client, error) {
	mongoURI := config.AppConfig.MongoURI

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	InitIndexes(client)

	log.Println("Successfully connected to MongoDB.")
	return client, nil
}
