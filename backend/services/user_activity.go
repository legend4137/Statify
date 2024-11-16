package services

import (
	"backend/config"
	"backend/models"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userActivityCollection = config.GetCollection(config.DB, "user_activity")
var userCollection = config.GetCollection(config.DB, "users")

func AddTrackToUserActivity(userID string, trackID int) (string, error) {
	convertedUserID, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Error converting user_id to int: %v", err)
		return "", err
	}

	// Fetching user age

	filter := bson.M{"user_id": convertedUserID}
	var result struct {
		Age int `bson:"age"`
	}
	// Execute the query
	err = userCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return "", fmt.Errorf("user with user_id %d not found", convertedUserID)
	} else if err != nil {
		return "", err
	}
	age := result.Age

	// Fetching Track info

	filter = bson.M{"track_id": trackID}
	var audio_features models.Song
	// Execute the query
	err = songsCollection.FindOne(context.TODO(), filter).Decode(&audio_features)
	if err == mongo.ErrNoDocuments {
		return "", fmt.Errorf("song with track_id %d not found", trackID)
	} else if err != nil {
		return "", err
	}

	// Check if the user exists
	filter = bson.M{"user_id": convertedUserID}
	var userActivity models.UserActivity
	err = userActivityCollection.FindOne(context.TODO(), filter).Decode(&userActivity)

	now := time.Now()

	if err == mongo.ErrNoDocuments {
		// User doesn't exist, create a new document
		newUser := models.UserActivity{
			UserID:           convertedUserID,
			Tracks:           []int{trackID}, // Add the track to the array
			Mood_Energy:       float64(audio_features.Energy),            // Default values for new user
			Mood_Valence:      float64(audio_features.Valence),
			Preferred_Genre:   "Unknown",
			Preferred_Language: "Unknown",
			Age:              age,
			Language:         "Unknown",
			CreatedAt: 	      now,
			UpdatedAt:		  now,
		}

		_, err := userActivityCollection.InsertOne(context.TODO(), newUser)
		if err != nil {
			log.Printf("Error creating new user: %v", err)
			return "", err
		}

		return fmt.Sprintf("New user created with user_id %d and track_id %d added to their activity", convertedUserID, trackID), nil
	} else if err != nil {
		// Unexpected error
		log.Printf("Error finding user: %v", err)
		return "", err
	}

	// User exists, check if track is already in their activity
	for _, existingTrackID := range userActivity.Tracks {
		if existingTrackID == trackID {
			return fmt.Sprintf("Track %d already exists in the user's activity", trackID), nil
		}
	}

	filter = bson.M{"user_id": convertedUserID}
	var user_activity models.UserActivity
	// Execute the query
	err = userActivityCollection.FindOne(context.TODO(), filter).Decode(&user_activity)
	if err == mongo.ErrNoDocuments {
		return "", fmt.Errorf("user_activity with user_id %d not found", convertedUserID)
	} else if err != nil {
		return "", err
	}
	mod_energy := (user_activity.Mood_Energy * float64(len(user_activity.Tracks)) + float64(audio_features.Energy))/float64((len(userActivity.Tracks)+1))
	mod_valence := (user_activity.Mood_Valence * float64(len(user_activity.Tracks)) + float64(audio_features.Valence))/float64((len(userActivity.Tracks)+1))

	// Add the track to the user's activity
	update := bson.M{
		"$push": bson.M{"tracks": trackID},
		"$set": bson.M{
			"mood_energy":  mod_energy,
			"mood_valence": mod_valence,
			"updated_at": 	now,
		},
	}

	_, err = userActivityCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf("Error updating user activity: %v", err)
		return "", err
	}

	return fmt.Sprintf("Track %d added to user_id %d's activity", trackID, convertedUserID), nil
}
