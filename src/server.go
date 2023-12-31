package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fatih/color"
)

var assetMaxAge = 15552000

func server() {
	// if os.Getenv("APP_ENV") == "development" {
	// 	assetMaxAge = 0
	// }

	err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database connection pool: %v", err)
	}

	mux := http.NewServeMux()

	// HTTP Route Handler for all pages
	mux.HandleFunc("/", routeHandler)

	// HTTP Route Handler for generated CSS files
	cssFileServer := http.FileServer(http.Dir(".generated/css"))
	mux.Handle("/css/", maxAgeHandler(assetMaxAge, http.StripPrefix("/css/", cssFileServer)))

	// HTTP Route Handler for generated CSS files
	esbuildFileServer := http.FileServer(http.Dir(".generated/esbuild/templates"))
	mux.Handle("/bundle/", maxAgeHandler(assetMaxAge, http.StripPrefix("/bundle/", esbuildFileServer)))

	// HTTP Route Handler for static files like favicon, robots etc
	fileServer := http.FileServer(http.Dir("static")) // serves any file in /static directory
	mux.Handle("/static/", maxAgeHandler(assetMaxAge, http.StripPrefix("/static/", fileServer)))
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	// Handler for image optimization
	mux.Handle(ImageBaseRoute+"/", maxAgeHandler(15552000, http.HandlerFunc(imageRouteHandler)))

	httpServer := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: mux,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			color.Red("Error starting server: %s\n", err)
			log.Fatal(err)
		}
	}()

	// block until a signal is received.
	<-stopChan
	log.Println("Shutting down server...")
	//create deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	httpServer.Shutdown(ctx)

	log.Println("Server gracefully stopped")
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", seconds))
		h.ServeHTTP(w, r)
	})
}

func routeHandler(w http.ResponseWriter, r *http.Request) {
	pageData, err := getPageData(r.URL.Path)
	if err != nil {
		notFound(w, r)
	}
	pageFound(pageData, w, r)
}

func imageRouteHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	widthStr := r.URL.Query().Get("width")

	// get all url query
	query := r.URL.Query()
	// println each query key and value
	// serverLogger("Image URL: " + url)
	for key, value := range query {
		serverLogger(key + ": " + value[0])
	}

	width, err := strconv.Atoi(widthStr)
	if err != nil {
		http.Error(w, "Invalid width", http.StatusBadRequest)
		return
	}

	format := getSupportedImageFormat(r.Header)

	optimizedImagePath, err := optimizeImage(url, width, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve the optimized image
	http.ServeFile(w, r, optimizedImagePath)
}

var version string

func getVersionHash() string {
	var versionHash string

	if os.Getenv("APP_ENV") == "development" {
		versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
	} else if os.Getenv("APP_ENV") == "production" {
		versionHash = version
	}

	return versionHash
}
