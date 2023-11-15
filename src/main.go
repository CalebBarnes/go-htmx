package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	banner()
	loadEnv()
	go postCSS()

	if os.Getenv("APP_ENV") == "development" {
		go watcher()
		go startBrowserSync()
	}

	server()
}

func loadEnv() {
	if os.Getenv("APP_ENV") == "development" {
		err := godotenv.Load(".env") // load .env file
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	appEnv := os.Getenv("APP_ENV")
	println("APP_ENV: ", appEnv)

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		os.Setenv("DB_HOST", "database")
	}
	println("DB_HOST: ", dbHost)

	port := os.Getenv("PORT")
	if port == "" {
		os.Setenv("PORT", "42069")
	}
}
