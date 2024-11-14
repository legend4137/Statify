package main

import (
	"backend/config"
	"backend/pkg/handlers"
	"backend/routes"
	"log"
	"net/http"
)

func main() {
	// Connect to MongoDB Atlas
	config.ConnectDB()
	routes.SetupRoutes()

	http.HandleFunc("/search_song", handlers.SearchSongHandler)
	// http.HandleFunc("/rate_song", handlers.RateSongHandler)
	http.HandleFunc("/predict", handlers.PredictHandler)
	// http.HandleFunc("/check_user", handlers.CheckUser)
	// http.HandleFunc("/add_user_profile", handlers.AddUserProfile)

	log.Println("Server is running on port 4567")
	log.Fatal(http.ListenAndServe(":4567", nil))
}
