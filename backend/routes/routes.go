package routes

import (
    "net/http"
    "backend/controllers"
)

func SetupRoutes() {
	// Auth routes
	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/login", controllers.Login)
	
    http.HandleFunc("/check_user", controllers.CheckUser)
    http.HandleFunc("/add_new_user", controllers.AddNewUser)
}