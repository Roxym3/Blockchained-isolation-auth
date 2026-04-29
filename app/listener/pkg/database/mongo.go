package database

import (
	"context"
	"fmt"
	"listener/pkg/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB(cfg *config.Database) (*mongo.Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client:%v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb server:%v", err)
	}

	return client, nil
}

func Disconnect(client *mongo.Client) error {
	if client == nil {
		return nil
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := client.Disconnect(closeCtx); err != nil {
		return fmt.Errorf("disconnect error: %v", err)
	}

	return nil
}
