package main
import (
	"html/template"
	"log"
	"net/http"
	"fmt"
	"time"
	"strconv"
)

type Seo struct {
	Title string
	Description string
}

var version string = "development"

func main() {
	fmt.Println("Starting server on http://localhost:42069")
	versionHash := version
		
	requestHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("URL Requested: ", r.URL.Path)
		if (version == "development") {
			versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
		}

		if (r.URL.Path != "/") {
			http.NotFound(w, r)
			return
		}

		data := map[string]interface{
		}{
			"Version":  versionHash,
			"Name": "Hello from my nested template!",
			"Films": getFilms(),
			"Seo": Seo{
				Title: "This is the SEO title",
				Description: "This is the SEO description",
			},
		}

		tmpl := template.Must(template.ParseFiles("templates/index.go.html"))
		template.Must(tmpl.ParseGlob("components/*.go.html"))

		tmpl.Execute(w, data)
	}

	http.HandleFunc("/", requestHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":42069", nil))
}


type Film struct {
	Title string
	Director string
	Year int
	Actor string
}

func getFilms() []Film {
	return []Film {
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
	}
}
