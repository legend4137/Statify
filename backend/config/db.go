package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "context"
)

var DB *mongo.Client

// ConnectDB connects to MongoDB using environment variables
func ConnectDB() *mongo.Client {
    err := godotenv.Load() // Load environment variables from .env file (optional)
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    mongoURI := os.Getenv("MONGO_URI")
    mongoDatabase := os.Getenv("MONGO_DATABASE")

    if mongoURI == "" || mongoDatabase == "" {
        log.Fatal("MongoDB connection string or database name not set in environment variables")
    }

    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatalf("MongoDB connection error: %v", err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatalf("MongoDB ping error: %v", err)
    }

    log.Println("Connected to MongoDB Atlas!")
    DB = client
    return client
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	databaseName := os.Getenv("MONGO_DATABASE")
    return client.Database(databaseName).Collection(collectionName)
}
