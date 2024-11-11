package services

import (
    "context"
    "log"
    "time"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var ratingsCollection *mongo.Collection
var songsCollection *mongo.Collection

func init() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://b22ai015:uwhNGGuRUOGYSpij@musicrmdvic.5p7x3.mongodb.net/?retryWrites=true&w=majority&appName=MusicRmdVic"))
    if err != nil {
        log.Fatalf("MongoDB connection error: %v", err)
    }

    ratingsCollection = client.Database("MusicRmdVic").Collection("ratings")
    songsCollection = client.Database("MusicRmdVic").Collection("songs")
}
