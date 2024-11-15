package models

// Song represents the song details structure.
type Spotify_Song struct {
	Id         string `json:"id"`
	Track      string `json:"track"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	ImageURL   string `json:"image_url"`
	TrackURI   string `json:"track_uri"`
	SpotifyURL string `json:"spotify_url"`
	Valence    float64 `json:"valence"`
	Energy     float64 `json:"energy"`
	Genre      string  `json:"genre"`
	Language   string  `json:"language"`
}

type Song struct {
	ID      string  `bson:"_id,omitempty"`
	TrackID int     `bson:"track_id"`
	Track   string  `bson:"track"`
	Artist  string  `bson:"artist"`
	Genre   string  `bson:"genre"`
	Language string `bson:"language"`
	Valence float64 `bson:"valence"`
	Energy  float64 `bson:"energy"`
}