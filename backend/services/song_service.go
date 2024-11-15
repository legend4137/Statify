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

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	spotifyAPIBase = "https://api.spotify.com/v1/search"
	flaskURL = "http://localhost:5000/predict"
)

var songsCollection = config.GetCollection(config.ConnectDB(), "songs")

// Fetch songs from MongoDB by track_id
func FetchSongsFromMongo(trackIDs []int) ([]models.Song, error) {

	// Query MongoDB for matching track_ids
	filter := bson.M{"track_id": bson.M{"$in": trackIDs}}
	cursor, err := songsCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query MongoDB: %w", err)
	}

	var songs []models.Song
	if err := cursor.All(context.TODO(), &songs); err != nil {
		return nil, fmt.Errorf("failed to decode MongoDB response: %w", err)
	}

	return songs, nil
}

func AddSongToCollection(spotifySong models.Spotify_Song) (int, error) {

	// Check if the song already exists by track name
	filter := bson.M{"track": spotifySong.Track}
	var existingSong models.Song
	err := songsCollection.FindOne(context.TODO(), filter).Decode(&existingSong)
	if err == nil {
		// Song already exists, return existing track_id
		return existingSong.TrackID, nil
	} else if err != mongo.ErrNoDocuments {
		// Error during query
		log.Println("Error checking song existence:", err)
		return 0, err
	}

	// Song does not exist, calculate the new track_id
	count, err := songsCollection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error counting documents:", err)
		return 0, err
	}
	newTrackID := int(count) + 1

	// Create a new Song object for insertion
	newSong := models.Song{
		TrackID:  newTrackID,
		Track:    spotifySong.Track,
		Artist:   spotifySong.Artist,
		Valence:  spotifySong.Valence,
		Energy:   spotifySong.Energy,
	}

	// Insert the new song into MongoDB
	_, err = songsCollection.InsertOne(context.TODO(), newSong)
	if err != nil {
		log.Println("Error inserting song into collection:", err)
		return 0, err
	}

	// Return the new track_id
	return newTrackID, nil
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
	track := response.Tracks.Items[0]
	trackID := track.Id

	// fmt.Println("Song Id:", track.Id)
	// fmt.Println("Song Name:", track.Name)
	// fmt.Println("Artist Name:", track.Artist[0].Name)
	// fmt.Println("Album Name:", track.Album.Name)
	// fmt.Println("Album Image:", track.Album.Images[0].URL)
	// fmt.Println("Spotify URL:", track.ExternalURLs.Spotify)

	// Fecth Genre
	baseURL = "https://api.spotify.com/v1/search"
	params = url.Values{}
	query = fmt.Sprintf("track:%s artist:%s", songName, artistName)
	params.Add("q", query)
	params.Add("type", "album")
	searchURL = fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create a new GET request with the access token in the Authorization header
	req, err = http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Panicf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return models.Spotify_Song{}, err
	}
	defer resp.Body.Close()

	var resp_genres struct {
		Artists struct {
			Items []struct {
				Genres []string `json:"genres"`
			} `json:"items"`
		} `json:"artists"`
	}

	// if len(resp_genres.Artists.Items) == 0 {
	// 	log.Panicln("Genere Not Found")
	// }

	// fmt.Println("Genere:", resp_genres.Artists.Items[0].Genres[0])
	fmt.Println(resp_genres)

	// Fetch audio features
	audioFeaturesURL := fmt.Sprintf("https://api.spotify.com/v1/audio-features/%s", trackID)
	audioReq, _ := http.NewRequest("GET", audioFeaturesURL, nil)
	audioReq.Header.Set("Authorization", "Bearer "+accessToken)

	audioResp, err := client.Do(audioReq)
	if err != nil {
		return models.Spotify_Song{}, err
	}
	defer audioResp.Body.Close()

	var audioFeatures struct {
		Valence float64 `json:"valence"`
		Energy  float64 `json:"energy"`
	}
	if err := json.NewDecoder(audioResp.Body).Decode(&audioFeatures); err != nil {
		return models.Spotify_Song{}, err
	}

	trackDetails := models.Spotify_Song{
		Id:         track.Id,
		Track:      track.Name,
		Artist:     track.Artist[0].Name,
		Album:      track.Album.Name,
		ImageURL:   track.Album.Images[0].URL,
		TrackURI:   track.URI,
		SpotifyURL: track.ExternalURLs.Spotify,
		Energy: 	audioFeatures.Energy,
		Valence: 	audioFeatures.Valence,
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

	log.Println(spotifySongs)
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