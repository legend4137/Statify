package controllers

import (
	"backend/services"
	"encoding/json"
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
		return
	}

	userID := requestData["user_id"]
	payload, _ := json.Marshal(map[string]string{"user_id": userID})

	tracks, err := services.GetPredictionFromFlask(payload)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch prediction from Flask"}`, http.StatusInternalServerError)
		return
	}

	trackDetails, err := services.GetSpotifyTracksDetails(tracks)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch Spotify track details"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackDetails)
}