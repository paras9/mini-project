package main

import (
	"log"
	"net/http"
	"os"

	"mini2/config"
	"mini2/database"
	"mini2/handlers"
	"mini2/models"
)

func main() {
	// Load environment variables
	config.LoadConfig()

	// Initialize logger
	config.InitLogger()
	logger := config.Logger

	// Initialize database
	database.InitDB(os.Getenv("DB_DSN"))
	database.Migrate(&models.Device{})

	// Setup routes
	http.HandleFunc("/upload", handlers.UploadHandler(logger, handlers.ProcessFile))
	http.HandleFunc("/entries", handlers.GetFilteredEntriesHandler(logger))
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	// Start server
	port := os.Getenv("SERVER_PORT")
	logger.Infof("Server started at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
