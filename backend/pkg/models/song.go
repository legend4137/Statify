package models

// Song represents the song details structure.
type Song struct {
	Track      string `json:"track"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	ImageURL   string `json:"image_url"`
	TrackURI   string `json:"track_uri"`
	SpotifyURL string `json:"spotify_url"`
}
