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

	_, err := services.GetSong(songName, artistName)
	if err != nil {
		http.Error(w, `{"error": "Song not found in database"}`, http.StatusNotFound)
		return
	}

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
	tracks, err := services.GetPredictionFromFlask(payload)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch prediction from Flask"}`, http.StatusInternalServerError)
		log.Println("Error fetching prediction from Flask:", err)
		return
	}

	// Fetch Spotify track details
	trackDetails, err := services.GetSpotifyTracksDetails(tracks)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch Spotify track details"}`, http.StatusInternalServerError)
		log.Println("Error fetching Spotify track details:", err)
		return
	}

	log.Println("Track details:", trackDetails) // Log final track details

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackDetails)
}
