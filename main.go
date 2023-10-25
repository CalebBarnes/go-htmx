package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
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
		
		
		pageDataTest := getPageFromDBByUrl(r.URL.Path)
		log.Println("pageDataTest: ", pageDataTest)
		
		pageData := getPageData(r.URL.Path)

		data := map[string]interface{
		}{
			"Version":  versionHash,
			"Seo": Seo{
				Title: pageData.Title,
				Description: "This is the SEO description",
			},
			"Data": pageData,
		}

		tmpl := template.Must(template.ParseFiles("templates/index.go.html"))
		template.Must(tmpl.ParseGlob("components/*.go.html"))

		tmpl.Execute(w, data)
	}

	http.HandleFunc("/", requestHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":42069", nil))
}


type Blocker interface{}

type Page struct {
    Url    string    `json:"Url"`
    Title  string    `json:"Title"`
    Blocks []Blocker `json:"Blocks"`
}

func getPageData(Url string) Page {
    fmt.Println("getPageData: ", Url)

    return Page{
        Url:   "/",
        Title: "Cookie's go-htmx - Home",
    }
}


func getPageFromDBByUrl(pageUrl string) (Page, error) {
	db, err := sqlx.Connect("postgres", "postgresql://directus@localhost:5432/directus")

	if err != nil {
		return Page{}, err
	}
	defer db.Close()

	var page Page
	err = db.Get(&page, "SELECT id, url, title FROM pages WHERE url = $1", pageUrl)
	if err != nil {
		return Page{}, err
	}

	// blocks, err := getBlocksFromDB(db, page.ID)
	// if err != nil {
	// 	return Page{}, err
	// }

	// page.Blocks = blocks
	return page, nil
}




// func getBlocksFromDB(pageID int) ([]map[string]interface{}, error) {
// 	db, err := sqlx.Connect("postgres", "your_connection_string")
// 	if err != nil {
// 		return nil, err
// 	}

// 	rows, err := db.Queryx("SELECT * FROM blocks WHERE page_id = $1", pageID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var blocks []map[string]interface{}

// 	for rows.Next() {
// 		rowMap := make(map[string]interface{})
// 		err := rows.MapScan(rowMap)
// 		if err != nil {
// 			return nil, err
// 		}
// 		blocks = append(blocks, rowMap)
// 	}

// 	return blocks, nil
// }