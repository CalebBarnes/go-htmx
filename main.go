package main
import (
	"html/template"
	"log"
	"net/http"
	"fmt"
	"time"
	"strconv"
	// "database/sql"
	"os/exec"
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
		
	requestHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("URL Requested: ", r.URL.Path)
		if (version == "development") {
			versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		
		if (r.URL.Path == "/_webhook/git-pull") {
			// secret: wqaeKr65XJZiSL
			// create a new *Cmd instance
			// here we pass the command as the first argument and the arguments to pass to the command as the
			// remaining arguments in the function
			cmd := exec.Command("git", "pull")

			// The `Output` method executes the command and
			// collects the output, returning its value
			out, err := cmd.Output()
			if err != nil {
			// if there was any error, print it here
			fmt.Println("could not run command 'git pull': ", err)
			}
			// otherwise, print the output from running the command
			fmt.Println("Output: ", string(out))
	
			cmd2 := exec.Command("npm", "run", "build")
			out2, err2 := cmd2.Output()
			if err2 != nil {
			// if there was any error, print it here
			fmt.Println("could not run command 'npm run build': ", err)
			}
			// otherwise, print the output from running the command
			fmt.Println("Output: ", string(out2))

			return
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
			"Films": getFilms(),	
		}

		tmpl.Execute(w, data)
	}

	http.HandleFunc("/", requestHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":42069", nil))
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