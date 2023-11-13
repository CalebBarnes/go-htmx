package main

import (
	"os"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
	// "net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Seo struct {
	Title string
	Description string
}



var version string = "development"

func main() {
	versionHash := version

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("static"))
	genFileServer := http.FileServer(http.Dir(".generated"))
	
	requestHandler := func(w http.ResponseWriter, r *http.Request) {
		if (version == "development") {
			versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
		}

		pageData, err := getPageData(r.URL.Path)
		
		if err != nil {
			tmpl := template.Must(template.ParseFiles("src/templates/404.go.html"))
			template.Must(tmpl.ParseGlob("src/components/*.go.html"))
			data := map[string]interface{}{
				"Version":  versionHash,
				"Seo": Seo {
					Title:"404 - Page not found",
					Description: "You've hit a dead end...",
				},
			}

			if (version == "development") {
				data["Env"] = "development"
			} else {
				data["Env"] = "production"
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

			if (version == "development") {
				data["Env"] = "development"
			} else {
				data["Env"] = "production"
			}

			tmpl, err := template.ParseFiles("src/templates/index.go.html")
			if err != nil {
				log.Fatalf("Error parsing main template: %v", err)
			}
			tmpl.Funcs(template.FuncMap{
				// Render HTML in a template without escaping it (or any other strings)
				"noescape": func(str string) template.HTML {
					return template.HTML(str)
				},
				"getImageProps": getImageProps,
			})
			tmpl, err = tmpl.ParseGlob("src/components/*.go.html")
			if err != nil {
				log.Fatalf("Error parsing component templates: %v", err)
			}
			tmpl, err = tmpl.ParseGlob("src/components/blocks/*.go.html")
			if err != nil {
				log.Fatalf("Error parsing block templates: %v", err)
			}
			tmpl, err = tmpl.Parse(blocksTemplateBuilder(pageData.Blocks))
			if err != nil {
				log.Fatalf("Error parsing block templates: %v", err)
			}

			if (version == "production") {
			w.Header().Add("Cache-Control", fmt.Sprintf("private, max-age=%d", 60))
			}
			if err := tmpl.Execute(w, data); err != nil {
				log.Println("Error executing template:", err)
			}
		}
	}

	mux.Handle("/.generated/", maxAgeHandler(15552000, http.StripPrefix("/.generated/", genFileServer)))
	mux.Handle("/static/", maxAgeHandler(15552000, http.StripPrefix("/static/", fileServer)))
	mux.HandleFunc("/", requestHandler)

	log.Fatal(http.ListenAndServe(":42069", mux))
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	// fmt.Println("maxAgeHandler", seconds)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", seconds))
		h.ServeHTTP(w, r)
	})
}

type Page struct {
    ID    int    `db:"id"`
    Uri   sql.NullString `db:"uri"`
	Status string `db:"status"`
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
		err = db.Get(&page, "SELECT id, uri, title, status FROM page WHERE uri = '' OR uri IS NULL AND status = 'published'")
	} else {
		err = db.Get(&page, "SELECT id, uri, title, status FROM page WHERE uri = $1 AND status = 'published'", pageUrl)
	}
	if err != nil {
		fmt.Println("failed to get page: ", err)
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
        // for key, value := range block.Data {
        //     if byteValue, ok := value.([]byte); ok {
        //         var deserializedData interface{}
        //         if err := json.Unmarshal(byteValue, &deserializedData); err == nil {
        //             block.Data[key] = deserializedData
        //         }
		// 		// else {
        //         //     fmt.Println("Failed to unmarshal:", err)
        //         // }
        //     }
        // }

		for key, value := range block.Data {
			switch v := value.(type) {
			case []byte:
				// Try to unmarshal as JSON
				var jsonData interface{}
				if err := json.Unmarshal(v, &jsonData); err == nil {
					block.Data[key] = jsonData
				} else {
					// If not JSON, convert to string
					block.Data[key] = string(v)
				}
			default:
				// Handle other types as needed
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
				// Check if template file exists
				if fileExists("src/components/blocks/"+block.Collection+".go.html") {
					blockBuilderStr+=`
					{{ else if eq .Collection "`+block.Collection+`" }}
						{{ template "`+block.Collection+`" .Data }}
					`
				}
			}
				blockBuilderStr+=`
			{{ end }}
		{{ end }} 
	{{ end }}
	`
	return blockBuilderStr
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func getImageProps(id string, options ...string) template.HTMLAttr {
    validAttrs := map[string]bool{
        "src": true, "width": true, "height": true, "class": true,
        "alt": true, "title": true, // add other valid attributes here
    }
    attrs := make(map[string]string)
    attrs["src"] = "https://go-htmx-directus.cookieserver.gg/assets/" + id
    var order []string
    order = append(order, "src")

    var layoutOption string // Variable to store the value of "layout" option

    for _, option := range options {
        parts := strings.Split(option, "=")
        if len(parts) == 2 {
            key := parts[0]
            if validAttrs[key] {
                if _, exists := attrs[key]; !exists {
                    order = append(order, key) // Record the order of valid keys
                }
                attrs[key] = parts[1]
            } else {
                // Process non-valid attributes here
                if key == "layout" {
                    layoutOption = parts[1] // Store the "layout" option for later use
                    // Additional processing for "layout" can be done here
                }
            }
        }
    }

    // Handle class attribute
    if classValue, exists := attrs["class"]; exists {
        attrs["class"] = classValue + " image-wrapper"
        if attrs["class"][0] == ' ' {
            attrs["class"] = attrs["class"][1:]
        }
    } else {
        attrs["class"] = "image-wrapper"
        order = append(order, "class")
    }

    var b strings.Builder
    for _, key := range order {
        b.WriteString(fmt.Sprintf(`%s="%s" `, key, attrs[key]))
    }

    // Additional logic using layoutOption here, if needed
	println(id)
	println(layoutOption)

    return template.HTMLAttr(b.String())
}


