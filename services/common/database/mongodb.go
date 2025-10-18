package database

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	UserDB      *mongo.Database
	WorkflowDB  *mongo.Database
)

func InitMongoDB(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	mongoClient = client
	UserDB = client.Database("seentics_users")
	WorkflowDB = client.Database("seentics_workflows")

	logrus.Info("Connected to MongoDB successfully")
	return nil
}

func GetMongoClient() *mongo.Client {
	return mongoClient
}

func GetUserDB() *mongo.Database {
	return UserDB
}

func GetWorkflowDB() *mongo.Database {
	return WorkflowDB
}

func CloseMongoDB() {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			logrus.Error("Error disconnecting from MongoDB: ", err)
		} else {
			logrus.Info("Disconnected from MongoDB")
		}
	}
}
