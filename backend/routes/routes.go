package routes

import (
	"backend/controllers"
	"net/http"
)

func SetupRoutes() {
	// Auth routes
	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/login", controllers.Login)

	http.HandleFunc("/check_user", controllers.CheckUser)
	http.HandleFunc("/add_new_user", controllers.AddNewUser)

	http.HandleFunc("/search_song", controllers.SearchSongHandler)
	http.HandleFunc("/predict", controllers.PredictHandler)
}
