package main
import (
	"html/template"
	"log"
	"net/http"
	"fmt"
	"time"
	"strconv"
)

type Film struct {
	Title string
	Director string
	Year int
	Actor string
}

var version string = "development"

func main() {
	fmt.Println("Starting server on http://localhost:42069")
	versionHash := version
		
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("URL Requested: ", r.URL.Path)
		if (version == "development") {
			 versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
		} 

		if (r.URL.Path != "/") {
			tmpl := template.Must(template.ParseFiles("templates/404.html"))
			tmpl.Execute(w, nil)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
	
		data := map[string]interface{
		}{
			"Version":  versionHash,
			"Films": []Film {
				{
					Title: "The Shawshank Redemption",
					Director: "Frank Darabont",
					Year: 1994,
					Actor: "Tim Robbins",
				},
				{
					Title: "The Godfather",
					Director: "Francis Ford Coppola",
					Year: 1972,
					Actor: "Marlon Brando",
				},
				{
					Title: "The Dark Knight",
					Director: "Christopher Nolan",
					Year: 2008,
					Actor: "Christian Bale",
				},
			},	
		}

		tmpl.Execute(w, data)
	}

	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":42069", nil))
}