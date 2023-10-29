package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
		if (version == "development") {
			versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
		}

		pageData, err := getPageData(r.URL.Path)
		
		if err != nil {
			tmpl := template.Must(template.ParseFiles("templates/404.go.html"))
			template.Must(tmpl.ParseGlob("components/*.go.html"))
			data := map[string]interface{}{
				"Version":  versionHash,
				"Seo": Seo {
					Title:"404 - Page not found",
					Description: "You've hit a dead end...",
				},
			}
			tmpl.Execute(w, data)
			
		} else {
			data := map[string]interface{
				}{
					"Version":  versionHash,
					"Seo": Seo{
						Title: pageData.Title,
						Description: "This is the SEO description",
					},
					"Data": pageData,
				}
			tmpl, err := template.ParseFiles("templates/index.go.html")
			if err != nil {
				log.Fatalf("Error parsing main template: %v", err)
			}
			tmpl.Funcs(template.FuncMap{
				// Render HTML in a template without escaping it (or any other strings)
				"noescape": func(str string) template.HTML {
					return template.HTML(str)
				},
			})
			tmpl, err = tmpl.ParseGlob("components/*.go.html")
			if err != nil {
				log.Fatalf("Error parsing component templates: %v", err)
			}
			tmpl, err = tmpl.ParseGlob("components/blocks/*.go.html")
			if err != nil {
				log.Fatalf("Error parsing block templates: %v", err)
			}
			tmpl, err = tmpl.Parse(blocksTemplateBuilder(pageData.Blocks))
			if err != nil {
				log.Fatalf("Error parsing block templates: %v", err)
			}
			if err := tmpl.Execute(w, data); err != nil {
				log.Println("Error executing template:", err)
			}
		}
	}

	http.HandleFunc("/", requestHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":42069", nil))
}


type Page struct {
    ID    int    `db:"id"`
    Uri   sql.NullString `db:"uri"`
    Title string `db:"title"`
	Blocks []Block
}

type BlockData struct {
    Collection string `db:"collection"`
    ID         int    `db:"id"`
    Item       string `db:"item"`
    PageID     int    `db:"page_id"`
    Sort       int    `db:"sort"`
}

type Block struct {
	Collection string `db:"collection"`
	Data map[string]interface{}
}

func getPageData(pageUrl string) (Page, error) {
	connectionString := "user=directus dbname=directus password=Y25GUFMNeaGpEd sslmode=disable"
	if (version == "development") {
		connectionString += " host=cookie-go-htmx"
	} else {
		connectionString += " host=database"
	}
	db, err := sqlx.Connect("postgres", connectionString)

	if err != nil {
		fmt.Println("failed to connect to db: ", err)
		return Page{}, err
	}
	defer db.Close()

	var page Page
	if pageUrl == "/" {
		err = db.Get(&page, "SELECT id, uri, title FROM page WHERE uri = '' OR uri IS NULL")
	} else {
		err = db.Get(&page, "SELECT id, uri, title FROM page WHERE uri = $1", pageUrl)
	}
	if err != nil {
		return Page{}, err
	}

	var blocksDatas []BlockData
	err = db.Select(&blocksDatas, "SELECT collection, id, item, page_id, sort FROM page_blocks WHERE page_id = $1 ORDER BY sort ASC", page.ID)
	if err != nil {
		fmt.Println(err)
	}

	blocks := make([]Block, 0) // Initialize the Blocks map

    for _, blockData := range blocksDatas {
		block := Block{
			Collection: blockData.Collection,
			Data: make(map[string]interface{}),
		} 
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", blockData.Collection)
		err = db.QueryRowx(query, blockData.Item).MapScan(block.Data)

		if err != nil {
            fmt.Println("Error querying for blocks:", err)
            fmt.Println("query:", query, blockData.ID)
            continue
        }

		// dynamically unmarshalling bytecode when a blocks field contains it (serialized json data usually)
        for key, value := range block.Data {
            if byteValue, ok := value.([]byte); ok {
                var deserializedData interface{}
                if err := json.Unmarshal(byteValue, &deserializedData); err == nil {
                    block.Data[key] = deserializedData
                }
				// else {
                //     fmt.Println("Failed to unmarshal:", err)
                // }
            }
        }
        // Append to the overall list of blocks.
        blocks = append(blocks, block)
	}

	page.Blocks = blocks
	return page, nil
}

func blocksTemplateBuilder(blocks []Block)(string){
	blockBuilderStr := `
	{{ define "blocks" }}    
		{{ range .Data.Blocks }}
			{{ if eq .Collection "a" }}
				`
			for _, block := range blocks {
				blockBuilderStr+=`
					{{ else if eq .Collection "`+block.Collection+`" }}
						{{ template "`+block.Collection+`" .Data }}
				`
			}
				blockBuilderStr+=`
			{{ end }}
		{{ end }} 
	{{ end }}
	`
	return blockBuilderStr
}