package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func New(url string) *mongo.Database {
	clientOptions := options.Client().ApplyURI(url)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Error while checking connection: %v", err)
	}

	log.Println("Successfully MongoDB connection")

	return client.Database("tasks")
}
