package main

import (
	"backend/config"
	"backend/routes"
	"log"
	"net/http"
)

func main() {
	// Connect to MongoDB Atlas
	config.ConnectDB()
	routes.SetupRoutes()

	log.Println("Server is running on port 4567")
	log.Fatal(http.ListenAndServe(":4567", nil))
}
