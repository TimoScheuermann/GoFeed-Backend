package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Database *mongo.Database
var client *mongo.Client

func Connect() {

	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err != nil {
		fmt.Printf("Error occured while connecting to MongoDB: %v\n", err.Error())
		os.Exit(1)
		return
	}

	err = c.Ping(context.Background(), nil)

	if err != nil {
		fmt.Printf("Couldn't connect to MongoDB: %v\n", err.Error())
		os.Exit(1)
		return
	}

	fmt.Println("Successfully connected to MongoDB")

	client = c
	Database = c.Database("gofeed-go")
}

func Disconnect() {
	if client != nil {
		client.Disconnect(context.Background())
	}
}
