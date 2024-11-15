package services

import (
	"backend/models"
	"backend/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	// "io"
	"log"
	"net/http"
	"net/url"
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	spotifyAPIBase = "https://api.spotify.com/v1/search"
	flaskURL = "http://localhost:5000/predict"
)

var userCollection = config.GetCollection(config.ConnectDB(), "songs")

// Fetch songs from MongoDB by track_id
func FetchSongsFromMongo(trackIDs []int) ([]models.Song, error) {

	// Query MongoDB for matching track_ids
	filter := bson.M{"track_id": bson.M{"$in": trackIDs}}
	cursor, err := userCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query MongoDB: %w", err)
	}

	var songs []models.Song
	if err := cursor.All(context.TODO(), &songs); err != nil {
		return nil, fmt.Errorf("failed to decode MongoDB response: %w", err)
	}

	return songs, nil
}

// GetSpotifyTrackDetails fetches song details from Spotify API.
func GetSpotifyTrackDetails(songName, artistName string) (models.Spotify_Song, error) {
	accessToken, err := getSpotifyAccessToken()

	if err != nil {
		return models.Spotify_Song{}, err
	}

	baseURL := "https://api.spotify.com/v1/search"
	params := url.Values{}
	query := fmt.Sprintf("track:%s artist:%s", songName, artistName)
	params.Add("q", query)
	params.Add("type", "track")
	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create a new GET request with the access token in the Authorization header
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Panicf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.Spotify_Song{}, err
	}
	defer resp.Body.Close()
	// Parse the response to get song information

	var response struct {
		Tracks struct {
			Items []struct {
				Id     string `json:"id"`
				Name   string `json:"name"`
				Artist []struct {
					Name string `json:"name"`
				} `json:"artists"`
				URI   string `json:"url"`
				Album struct {
					Name   string `json:"name"`
					Images []struct {
						URL string `json:"url"`
					} `json:"images"`
				} `json:"album"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.Spotify_Song{}, err
	}

	// Check if any song was found
	if len(response.Tracks.Items) == 0 {
		log.Panicf("No Song Found")
		return models.Spotify_Song{}, nil
	}

	// Display song information
	song := response.Tracks.Items[0]
	fmt.Println("Song Id:", song.Id)
	fmt.Println("Song Name:", song.Name)
	fmt.Println("Artist Name:", song.Artist[0].Name)
	fmt.Println("Album Name:", song.Album.Name)
	fmt.Println("Album Image:", song.Album.Images[0].URL)
	fmt.Println("Spotify URL:", song.ExternalURLs.Spotify)

	trackDetails := models.Spotify_Song{
		Id:         song.Id,
		Track:      song.Name,
		Artist:     song.Artist[0].Name,
		Album:      song.Album.Name,
		ImageURL:   song.Album.Images[0].URL,
		TrackURI:   song.URI,
		SpotifyURL: song.ExternalURLs.Spotify,
	}

	return trackDetails, nil
}

func FetchSpotifyDetailsForTracks(trackIDs []int) ([]models.Spotify_Song, error) {
	// Fetch songs from MongoDB
	songs, err := FetchSongsFromMongo(trackIDs)
	if err != nil {
		log.Println("Error fetching songs from MongoDB:", err)
		return nil, err
	}

	var spotifySongs []models.Spotify_Song
	for _, song := range songs {
		// Call Spotify API for each song
		spotifySong, err := GetSpotifyTrackDetails(song.Track, song.Artist)
		if err != nil {
			log.Printf("Error fetching Spotify details for track '%s' by artist '%s': %v", song.Track, song.Artist, err)
			continue
		}

		log.Printf("Fetched Spotify song details: %+v", spotifySong)
		spotifySongs = append(spotifySongs, spotifySong)
	}

	return spotifySongs, nil
}

// GetPredictionFromFlask calls Flask service to get song predictions.
func GetPredictionFromFlask(payload []byte) ([]models.Spotify_Song, error) {
	req, _ := http.NewRequest("POST", flaskURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("Error during Flask POST request:", err)
		return nil, errors.New("failed to fetch prediction from Flask")
	}
	defer resp.Body.Close()

	var response struct {
		Songs []int `json:"songs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println("Error decoding Flask response:", err)
		return nil, err
	}

	// log.Println("Track IDs from Flask:", response.Songs)

	// Fetch Spotify details
	spotifySongs, err := FetchSpotifyDetailsForTracks(response.Songs)
	if err != nil {
		log.Fatalf("Failed to fetch Spotify song details: %v", err)
	}
	
	// log.Println("Return Value:", spotifySongs)
	return spotifySongs, nil
}