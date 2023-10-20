package main
import (
	"html/template"
	"log"
	"net/http"
	"fmt"
)

type Film struct {
	Title string
	Director string
	Year int
	Actor string
}

func main() {
	fmt.Println("Starting server on http://localhost:42069")

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("URL Requested: ", r.URL.Path)

		if (r.URL.Path != "/") {
			tmpl := template.Must(template.ParseFiles("templates/404.html"))
			// http.Error(w, "404 Not Found", http.StatusNotFound)
			tmpl.Execute(w, nil)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		
		ctx := map[string][]Film{
			"Films": {
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

		tmpl.Execute(w, ctx)
	}

	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":42069", nil))
}