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
	Page      Page
	Timestamp time.Time
}

var (
	pageCache = make(map[string]CacheEntry)
	cacheTTL  = 1 * time.Minute // TTL of 30 minutes
)

func getPageData(pageUrl string) (Page, error) {
	cacheMutex.RLock()
	entry, found := pageCache[pageUrl]
	cacheMutex.RUnlock()

	if found {
		// Check if the entry is still valid
		if time.Since(entry.Timestamp) < cacheTTL {
			return entry.Page, nil
		}
		// Invalidate the stale entry
		invalidateCache(pageUrl)
	}

	connectionString := "user=directus dbname=directus password=Y25GUFMNeaGpEd sslmode=disable"
	connectionString += " host=" + os.Getenv("DB_HOST")

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

	// Update the cache
	cacheMutex.Lock()
	pageCache[pageUrl] = CacheEntry{Page: page, Timestamp: time.Now()}
	cacheMutex.Unlock()

	return page, nil
}

// invalidateCache removes a specific page from the cache
func invalidateCache(pageUrl string) {
	cacheMutex.Lock()
	delete(pageCache, pageUrl)
	cacheMutex.Unlock()
}

// invalidateAllCache clears the entire page cache
func invalidateAllCache() {
	cacheMutex.Lock()
	pageCache = make(map[string]CacheEntry)
	cacheMutex.Unlock()
}
