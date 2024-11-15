package services

import (
	"backend/config"
	"backend/models"
	"context"
	"fmt"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ratingsCollection = config.GetCollection(config.ConnectDB(), "ratings")

// AddOrCheckRating checks if a rating exists or adds a new one
func AddOrCheckRating(userID string, trackID int) (string, error) {
	// Convert user_id to int
	convertedUserID, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Error converting user_id to int: %v", err)
		return "", err
	}

	// Check if the rating already exists
	filter := bson.M{"user_id": convertedUserID, "track_id": trackID}
	var existingRating models.Rating
	err = ratingsCollection.FindOne(context.TODO(), filter).Decode(&existingRating)
	if err == nil {
		// Rating exists, return a message
		return fmt.Sprintf("Rating already exists for user_id %d and track_id %d", convertedUserID, trackID), nil
	}

	if err != mongo.ErrNoDocuments {
		// Unexpected error
		log.Printf("Error checking for existing rating: %v", err)
		return "", err
	}

	// Rating doesn't exist, create a new document
	newRating := models.Rating{
		UserID:  convertedUserID,
		TrackID: trackID,
		Rating:  1, // Default rating
	}

	_, err = ratingsCollection.InsertOne(context.TODO(), newRating)
	if err != nil {
		log.Printf("Error inserting new rating: %v", err)
		return "", err
	}

	return fmt.Sprintf("New rating added for user_id %d and track_id %d", convertedUserID, trackID), nil
}