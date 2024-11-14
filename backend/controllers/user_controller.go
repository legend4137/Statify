package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"backend/config"
	"backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = config.GetCollection(config.ConnectDB(), "users")

// Check if a user exists
func CheckUser(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		UserID string `json:"user_id"`
	}

	json.NewDecoder(r.Body).Decode(&requestData)
	var user models.User

	filter := bson.M{"user_id": requestData.UserID}
	err := userCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// Add a new user
func AddNewUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request body: %v", err), http.StatusBadRequest)
		return
	}

	user.ID = primitive.NewObjectID()

	_, err = userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Printf("Error inserting new user: %v", err)
		http.Error(w, "Error adding user", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
