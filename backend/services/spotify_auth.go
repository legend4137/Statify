package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// getSpotifyAccessToken retrieves the Spotify API access token
func getSpotifyAccessToken() (string, error) {
	url := "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte("grant_type=client_credentials")))
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(spotifyClientID+":"+spotifyClientSecret)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result["access_token"].(string), nil
}
