package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	banner()
	loadEnv()
	go bundleCss()

	println("Running APP_VER: ", os.Getenv("APP_VER"))
	if os.Getenv("APP_VER") == "development" {
		go watcher()
	}
	go startBrowserSync()
	server()
}

func loadEnv() {
	if os.Getenv("APP_ENV") == "development" {
		err := godotenv.Load() // load .env file
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		os.Setenv("PORT", "42069")
	}
}
