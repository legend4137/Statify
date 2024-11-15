package controllers

import (
	"backend/services"
	"encoding/json"
	"log"
	"net/http"
)

// const flaskURL = "http://localhost:5000/predict"

// SearchSongHandler handles song search requests.
func SearchSongHandler(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	songName := requestData["song_name"]
	artistName := requestData["artist_name"]

	trackDetails, err := services.GetSpotifyTrackDetails(songName, artistName)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch Spotify track"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackDetails)
}

// PredictHandler handles prediction requests.
func PredictHandler(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error decoding request body:", err)
		return
	}

	userID := requestData["user_id"]
	log.Println("Received user_id:", userID) // Log user ID for debugging

	payload, _ := json.Marshal(map[string]string{"user_id": userID})
	log.Println("Payload sent to Flask:", string(payload)) // Log payload data

	// Fetch predictions from Flask
	spotify_songs, err := services.GetPredictionFromFlask(payload)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch prediction from Flask"}`, http.StatusInternalServerError)
		log.Println("Error fetching prediction from Flask:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spotify_songs)
}
