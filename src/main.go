package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	banner()

	if os.Getenv("APP_ENV") == "development" {
		go watcher()
		go startBrowserSync()
	}

	bundleAssets()
	server()
}

func loadEnv() {
	os.Setenv("APP_ENV", "production")
	err := godotenv.Load(".env") // load .env file
	if err == nil {
		os.Setenv("APP_ENV", "development")
	}

	// set default db host
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		os.Setenv("DB_HOST", "database")
	}

	// set default port
	port := os.Getenv("PORT")
	if port == "" {
		os.Setenv("PORT", "42069")
	}
}
