package main

import (
	"backend/pkg/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/search_song", handlers.SearchSongHandler)
	http.HandleFunc("/rate_song", handlers.RateSongHandler)
	http.HandleFunc("/predict", handlers.PredictHandler)

	log.Println("Server is running on port 4567")
	log.Fatal(http.ListenAndServe(":4567", nil))
}
