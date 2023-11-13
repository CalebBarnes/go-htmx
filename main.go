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
	"net/url"

	"strings"
	"image"

	"crypto/sha1"
    "io"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/disintegration/imaging"
)

type Seo struct {
	Title string
	Description string
}



var version string = "development"

func main() {
	versionHash := version

	mux := http.NewServeMux()
	
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

	// Handler for generated CSS files
	cssFileServer := http.FileServer(http.Dir(".generated/css"))
	mux.Handle("/css/", http.StripPrefix("/css/", cssFileServer))
	
	// Handler for image optimization
	mux.Handle("/.generated/images/", http.HandlerFunc(imageHandler))

	// Handler for static files like favicon, robots etc
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", maxAgeHandler(15552000, http.StripPrefix("/static/", fileServer)))

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

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
    attrs := make(map[string]string)
    imageUrl := "https://go-htmx-directus.cookieserver.gg/assets/" + id
    maxWidth := 1920 // Default maxWidth
    customSrcsetProvided := false
    customSizesProvided := false
    lazyLoading := false // Default lazy loading

    for _, option := range options {
        parts := strings.Split(option, "=")
        if len(parts) == 2 {
            key := parts[0]
            value := parts[1]

            switch key {
            case "maxWidth":
                var err error
                maxWidth, err = strconv.Atoi(value)
                if err != nil {
                    maxWidth = 1920 // Reset to default if conversion fails
                }
            case "srcset":
                customSrcsetProvided = true
                attrs["srcset"] = value // Use the provided custom srcset
            case "sizes":
                customSizesProvided = true
                attrs["sizes"] = value // Use the provided custom sizes
            case "loading":
                if value == "lazy" {
                    lazyLoading = true
                }
            default:
                attrs[key] = value // Handle other valid attributes
            }
        }
    }

    // Generate default srcset if not provided
    if !customSrcsetProvided {
        widths := generateWidths(maxWidth)
        srcsetValues := make([]string, len(widths))
        for i, width := range widths {
            optimizedImageUrl := fmt.Sprintf("/.generated/images/image.png?url=%s&width=%d", url.QueryEscape(imageUrl), width)
            srcsetValues[i] = fmt.Sprintf("%s %dw", optimizedImageUrl, width)
        }
        attrs["srcset"] = strings.Join(srcsetValues, ", ")
    }

    // Generate default sizes if not provided
    if !customSizesProvided {
        attrs["sizes"] = "(max-width: 600px) 100vw, (max-width: 1024px) 50vw, 25vw"
    }

    attrs["src"] = fmt.Sprintf("/.generated/images/image.png?url=%s&width=%d", url.QueryEscape(imageUrl), maxWidth)
    
    if lazyLoading {
        attrs["loading"] = "lazy"
    }

    var b strings.Builder
    for key, value := range attrs {
        b.WriteString(fmt.Sprintf(`%s="%s" `, key, value))
    }

    return template.HTMLAttr(b.String())
}



// simple in-memory cache
var optimizedImageCache = make(map[string]bool)
func optimizeImage(url string, width int) (string, error) {
    start := time.Now()
    cacheKey := fmt.Sprintf("%s-%d", url, width)
    
    // Check cache first
    if _, exists := optimizedImageCache[cacheKey]; exists {
        optimizedImagePath := getOptimizedImagePath(url, width)
		fmt.Println("image already processed: ", optimizedImagePath)
        return optimizedImagePath, nil
    }
	fmt.Println("image not processed yet: ", url)

    // Download the image from the URL
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("failed to download image: ", err)
        return "", err
    }
    defer resp.Body.Close()

    // Decode the image
    srcImage, _, err := image.Decode(resp.Body)
    if err != nil {
        fmt.Println("failed to decode image: ", err)
        return "", err
    }

    // Image processing logic
    dstImageFill := imaging.Fill(srcImage, width, width, imaging.Center, imaging.Lanczos)

	imgPath := getOptimizedImagePath(url, width) // example .generated/images/3f6efd64ad786567e3ad5f558e01238bdb079ee3/500w.jpg
	println(imgPath)

	// make sure the directory exists (without the filename)
	err = os.MkdirAll(imgPath[:strings.LastIndex(imgPath, "/")], 0755)
	if err != nil {
		fmt.Println("failed to create directory: ", err)
		return "", err
	}

    // Save the file in .generated/images/<imageID>/<width>.jpg
    err = imaging.Save(dstImageFill, imgPath)
	

    if err != nil {
        fmt.Println("failed to save image: ", err)
        return "", err
    }

    // Update the cache
    optimizedImageCache[cacheKey] = true

    // Log time duration
    duration := time.Since(start)
    fmt.Println("image processed: ", duration)

    return getOptimizedImagePath(url, width), nil
}


func getOptimizedImagePath(url string, width int) string {
	// encode url to be safe for urls and file systems
	safeImageName := convertURLToFilePath(url)
    return fmt.Sprintf(".generated/images/%s/%dw.jpg", safeImageName, width)
}
func convertURLToFilePath(url string) string {
    h := sha1.New()
    io.WriteString(h, url)
    hashed := fmt.Sprintf("%x", h.Sum(nil))
    return hashed
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
    // Extract the necessary parameters from the URL
    // For example, the image ID and the requested width
    // This depends on how your URLs are structured
    url := r.URL.Query().Get("url")
    widthStr := r.URL.Query().Get("width")
    width, err := strconv.Atoi(widthStr)
    if err != nil {
        http.Error(w, "Invalid width", http.StatusBadRequest)
        return
    }

    optimizedImagePath, err := optimizeImage(url, width)
    if err != nil {
        // Handle errors (e.g., image not found, processing error)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Set cache control headers
    w.Header().Set("Cache-Control", "public, max-age=15552000")

    // Serve the optimized image
    http.ServeFile(w, r, optimizedImagePath)
}

func generateWidths(maxWidth int) []int {
    // Define a set of widths up to maxWidth
    // Example: [320, 480, 640, 800, 960, 1280, 1600, maxWidth]
    // Implement logic to generate this array based on maxWidth
	return []int{320, 480, 640, 768, 1024, 1280, 1600, maxWidth}
}