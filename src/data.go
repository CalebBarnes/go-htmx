package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

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

var cacheMutex sync.RWMutex

type CacheEntry struct {
	Page       Page
	Timestamp  time.Time
	Stale      bool
	Refreshing bool // New field to indicate a refresh is in progress
}

var (
	pageCache = make(map[string]CacheEntry)
	cacheTTL  = 10 * time.Second
)

func getPageData(pageUrl string) (Page, error) {
	cacheMutex.RLock()
	entry, found := pageCache[pageUrl]
	cacheMutex.RUnlock()

	if found {
		if time.Since(entry.Timestamp) < cacheTTL {
			// color.Green("Cache hit for %s", pageUrl)
			return entry.Page, nil
		} else {
			if !entry.Stale || (entry.Stale && !entry.Refreshing) {
				// color.Yellow("Cache stale for %s", pageUrl)
				cacheMutex.Lock()
				if !pageCache[pageUrl].Refreshing { // Check again to avoid race condition
					entry.Stale = true
					entry.Refreshing = true // Set the refreshing flag
					pageCache[pageUrl] = entry
					go refreshPageData(pageUrl) // Refresh the data in a separate goroutine
				}
				cacheMutex.Unlock()
				return entry.Page, nil
			}
		}
	}

	// color.Red("Cache miss for %s", pageUrl)
	return queryPageDataFromDB(pageUrl)
}

func refreshPageData(pageUrl string) {
	// fmt.Printf("Refreshing cache for %s\n", pageUrl)
	_, err := queryPageDataFromDB(pageUrl)
	if err != nil {
		fmt.Printf("Error refreshing page data for %s: %v\n", pageUrl, err)
		return
	}
}

var db *sqlx.DB

func initDB() error {
	var err error

	connectionString := "sslmode=disable"
	connectionString += " user=" + os.Getenv("DB_USER")
	connectionString += " dbname=" + os.Getenv("DB_NAME")
	connectionString += " password=" + os.Getenv("DB_PASSWORD")
	connectionString += " host=" + os.Getenv("DB_HOST")

	db, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}

	// Configure the connection pool here
	db.SetMaxOpenConns(25)                 // Set the maximum number of open connections
	db.SetMaxIdleConns(10)                 // Set the maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Set the maximum amount of time a connection may be reused

	return nil
}

func queryPageDataFromDB(pageUrl string) (Page, error) {
	// fmt.Printf("Querying DB for %s\n", pageUrl)

	var page Page
	var err error
	if pageUrl == "/" {
		err = db.Get(&page, "SELECT id, uri, title, status FROM page WHERE uri = '' OR uri IS NULL AND status = 'published'")
	} else {
		err = db.Get(&page, "SELECT id, uri, title, status FROM page WHERE uri = $1 AND status = 'published'", pageUrl)
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

	cacheMutex.Lock()
	pageCache[pageUrl] = CacheEntry{Page: page, Timestamp: time.Now(), Stale: false}
	cacheMutex.Unlock()
	// color.Yellow("Cache updated for %s\n", pageUrl)

	return page, nil
}
