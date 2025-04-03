package main

import (
	"expo-open-ota/config"
	"expo-open-ota/internal/metrics"
	"expo-open-ota/internal/router"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

func init() {
	config.LoadConfig()
	metrics.InitMetrics()
}

func main() {
	router := infrastructure.NewRouter()
	
	// Get port from environment variable for Heroku compatibility
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	log.Printf("Server is running on port %s", port)
	corsOptions := handlers.CORS(
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	err := http.ListenAndServe(":" + port, corsOptions(router))
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
