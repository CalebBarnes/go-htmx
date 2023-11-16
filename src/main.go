package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	banner()
	postCSS()

	if os.Getenv("APP_ENV") == "development" {
		go watcher()
		go startBrowserSync()
	}

	server()
}

func loadEnv() {
	os.Setenv("APP_ENV", "production")
	err := godotenv.Load(".env") // load .env file
	if err == nil {
		os.Setenv("APP_ENV", "development")
	}

	appEnv := os.Getenv("APP_ENV")
	println("APP_ENV: ", appEnv)

	// set default db host
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		os.Setenv("DB_HOST", "database")
	}
	println("DB_HOST: ", dbHost)

	// set default port
	port := os.Getenv("PORT")
	if port == "" {
		os.Setenv("PORT", "42069")
	}
}
