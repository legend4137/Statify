package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"context"
	"fmt"
	"net/http"
	"backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

const spotifyClientID = "e113dc056db54c8d8f39dd670061eb0c"
const spotifyClientSecret = "7b873f8a97244f9eb937c5fa3b3c0e93"
const flaskURL = "http://localhost:5000/predict"

// GetSong retrieves a song from the database.
func GetSong(songName, artistName string) (models.Song, error) {
	var song models.Song
	// Fetch song from MongoDB (mocked)
	// Ideally, interact with the database here to fetch song details
	return song, nil
}

// GetSpotifyTrackDetails fetches song details from Spotify API.
func GetSpotifyTrackDetails(songName, artistName string) (models.Song, error) {
	accessToken, err := getSpotifyAccessToken()
	if err != nil {
		return models.Song{}, err
	}

	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/search?q=track:%s artist:%s&type=track", songName, artistName)
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.Song{}, err
	}
	defer resp.Body.Close()

	var spotifyData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&spotifyData)

	items := spotifyData["tracks"].(map[string]interface{})["items"].([]interface{})
	if len(items) > 0 {
		track := items[0].(map[string]interface{})
		trackDetails := models.Song{
			Track:      songName,
			Artist:     artistName,
			Album:      track["album"].(map[string]interface{})["name"].(string),
			ImageURL:   track["album"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string),
			TrackURI:   track["uri"].(string),
			SpotifyURL: track["external_urls"].(map[string]interface{})["spotify"].(string),
		}

		return trackDetails, nil
	}

	return models.Song{}, errors.New("song not found on Spotify")
}

// GetPredictionFromFlask calls Flask service to get song predictions.
func GetPredictionFromFlask(payload []byte) ([]models.Song, error) {
	req, _ := http.NewRequest("POST", flaskURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch prediction from Flask")
	}
	defer resp.Body.Close()

	var tracks []models.Song
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		return nil, err
	}

	return tracks, nil
}

// RateSong saves the rating for a song in the database.
func RateSong(userID, trackID string, rating int) error {
	// Check if the rating is valid (you can modify this validation based on your needs).
	if rating < 1 || rating > 5 {
		return errors.New("invalid rating value, must be between 1 and 5")
	}

	// Insert the rating into the MongoDB collection.
	_, err := ratingsCollection.InsertOne(context.Background(), bson.M{
		"user_id": userID,
		"track_id": trackID,
		"rating":   rating,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetSpotifyTracksDetails fetches Spotify details for multiple tracks.
func GetSpotifyTracksDetails(tracks []models.Song) ([]models.Song, error) {
	var trackDetails []models.Song
	for _, track := range tracks {
		spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/search?q=track:%s artist:%s&type=track", track.Track, track.Artist)
		req, _ := http.NewRequest("GET", spotifyURL, nil)
		accessToken, err := getSpotifyAccessToken()
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "Bearer "+ accessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var spotifyData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&spotifyData)

		items := spotifyData["tracks"].(map[string]interface{})["items"].([]interface{})
		if len(items) > 0 {
			spTrack := items[0].(map[string]interface{})
			trackDetails = append(trackDetails, models.Song{
				Track:      track.Track,
				Artist:     track.Artist,
				Album:      spTrack["album"].(map[string]interface{})["name"].(string),
				ImageURL:   spTrack["album"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string),
				TrackURI:   spTrack["uri"].(string),
				SpotifyURL: spTrack["external_urls"].(map[string]interface{})["spotify"].(string),
			})
		}
	}

	return trackDetails, nil
}
