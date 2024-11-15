package controllers

import (
	"backend/config"
	"backend/models"
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	counterCollection = config.GetCollection(config.ConnectDB(), "counters")
)

// getNextSequence gets the next auto-incrementing ID
func getNextSequence(sequenceName string) (int, error) {
	var counter models.Counter

	// Find and update the counter atomically
	result := counterCollection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": sequenceName},
		bson.M{"$inc": bson.M{"sequence": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)

	err := result.Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.Sequence, nil
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	var existingUser models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Get next user ID
	userID, err := getNextSequence("user_id")
	if err != nil {
		http.Error(w, "Error generating user ID", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Create new user
	newUser := models.User{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		Email:            req.Email,
		Password:         string(hashedPassword),
		UserSongLanguage: req.UserSongLanguage,
		UserName:         req.UserName,
		UserAge:          req.UserAge,
		UserGender:       req.UserGender,
	}

	newUserActivity := models.UserActivity{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		Tracks:       []string{},
		Mood_Energy:  0,
		Mood_Valence: 0,
		// Preferred_Genre:    req.Preferred_Genre,
		Preferred_Language: req.UserName,
		Age:                req.UserAge,
		Language:           req.UserSongLanguage,
	}

	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	_, err = userCollection.InsertOne(context.TODO(), newUserActivity)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Don't send password back in response
	newUser.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
}

// Login handles user login
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Don't send password back in response
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
