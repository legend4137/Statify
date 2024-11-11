package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    // Initialize server settings or configurations
    fmt.Println("Starting server...")

    // Set up routes and handlers (using your services, etc.)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, Statify!")
    })

    // Start the server
    log.Fatal(http.ListenAndServe(":8080", nil))
}
