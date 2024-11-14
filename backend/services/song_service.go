package services

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const (
	spotifyClientID = "e113dc056db54c8d8f39dd670061eb0c"
	spotifyClientSecret = "7b873f8a97244f9eb937c5fa3b3c0e93"
	spotifyAPIBase = "https://api.spotify.com/v1/search"
	flaskURL = "http://localhost:5000/predict"
)

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
		return models.Song{}, err
	}
	defer resp.Body.Close()
	// Parse the response to get song information

	var response struct {
		Tracks struct {
			Items []struct {
				Id string `json:"id"`
				Name   string `json:"name"`
				Artist []struct {
					Name string `json:"name"`
				} `json:"artists"`
				URI   string `json:"url"`
				Album struct {
					Name string `json:"name"`
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
		return models.Song{}, err
	}

	// Check if any song was found
	if len(response.Tracks.Items) == 0 {
		log.Panicf("No Song Found")
		return models.Song{}, nil 
	}

	// Display song information
	song := response.Tracks.Items[0]
	fmt.Println("Song Id:", song.Id)
	fmt.Println("Song Name:", song.Name)
	fmt.Println("Artist Name:", song.Artist[0].Name)
	fmt.Println("Album Name:", song.Album.Name)
	fmt.Println("Album Image:", song.Album.Images[0].URL)
	fmt.Println("Spotify URL:", song.ExternalURLs.Spotify)

	trackDetails := models.Song{
		Id: 		song.Id,
		Track:      song.Name,
		Artist:     song.Artist[0].Name,
		Album:      song.Album.Name,
		ImageURL:   song.Album.Images[0].URL,
		TrackURI:   song.URI,
		SpotifyURL: song.ExternalURLs.Spotify,
	}

	return trackDetails, nil
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
