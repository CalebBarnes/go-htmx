package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Seo struct {
	Title       string
	Description string
}

type Page struct {
	ID     int            `db:"id"`
	Uri    sql.NullString `db:"uri"`
	Status string         `db:"status"`
	Title  string         `db:"title"`
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
	Data       map[string]interface{}
}

func getPageData(pageUrl string) (Page, error) {
	connectionString := "user=directus dbname=directus password=Y25GUFMNeaGpEd sslmode=disable"
	if os.Getenv("APP_VER") == "development" {
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
			Data:       make(map[string]interface{}),
		}
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", blockData.Collection)
		err = db.QueryRowx(query, blockData.Item).MapScan(block.Data)

		if err != nil {
			fmt.Println("Error querying for blocks:", err)
			fmt.Println("query:", query, blockData.ID)
			continue
		}

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

func blocksTemplateBuilder(blocks []Block) string {
	blockBuilderStr := `
	{{ define "blocks" }}
		{{ range .Data.Blocks }}
			{{ if eq .Collection "a" }}
				`
	for _, block := range blocks {
		// Check if template file exists
		if fileExists("src/components/blocks/" + block.Collection + ".go.html") {
			blockBuilderStr += `
					{{ else if eq .Collection "` + block.Collection + `" }}
						{{ template "` + block.Collection + `" .Data }}
					`
		}
	}
	blockBuilderStr += `
			{{ end }}
		{{ end }} 
	{{ end }}
	`
	return blockBuilderStr
}
