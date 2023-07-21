package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func ConnectDB() *mongo.Client {
	uri := "mongodb://localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

var Client *mongo.Client = ConnectDB()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("restaurant_management").Collection(collectionName)
	return collection
}
