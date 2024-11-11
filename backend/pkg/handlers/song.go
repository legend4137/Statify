package handlers

import (
	"encoding/json"
	"net/http"
	"your_project/pkg/services"
)

const flaskURL = "http://localhost:5000/predict"

// SearchSongHandler handles song search requests.
func SearchSongHandler(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	songName := requestData["song_name"]
	artistName := requestData["artist_name"]

	song, err := services.GetSong(songName, artistName)
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

// RateSongHandler handles song rating requests.
func RateSongHandler(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := requestData["user_id"]
	trackID := requestData["track_id"]
	rating := 1

	if err := services.RateSong(userID, trackID, rating); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Rating saved successfully"})
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
