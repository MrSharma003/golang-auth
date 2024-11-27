package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var(
	clientInstance *mongo.Client
	clientOnce sync.Once
)


func DBinstance() *mongo.Client {
	clientOnce.Do(func() { // Ensures this block runs only once
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		MongoDb := os.Getenv("MONGODB_URL")
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(MongoDb).SetServerAPIOptions(serverAPI)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connected to MongoDB")
		clientInstance = client
	})
	return clientInstance
}

// OpenCollection opens a collection with the singleton client
func OpenCollection(collectionName string) *mongo.Collection {
	client := DBinstance() // Ensures singleton client
	return client.Database("cluster0").Collection(collectionName)
}
