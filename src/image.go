package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"image"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fatih/color"
)

// todo, get the request headers here and check the Accept header for what image types are supported
// text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
func getImageProps(headers http.Header, id string, options ...string) template.HTMLAttr {
	color.Red("getImageProps called with id: " + id)

	attrs := make(map[string]string)
	order := []string{"src"} // Start with src as the first key
	imageUrl := os.Getenv("DIRECTUS_URL") + "/assets/" + id
	maxWidth := 1920 // Default maxWidth
	customSrcsetProvided := false
	customSizesProvided := false
	// customWidthProvided := false
	// customHeightProvided := false

	// get list of supported image types from the Accept header
	// if the Accept header is */*, assume all types are supported
	// if the Accept header is empty, assume all types are supported
	// if the Accept header is not empty, assume only the types in the header are supported

	acceptHeader := headers.Get("Accept")

	imageTypes := []string{"image/png", "image/jpeg", "image/webp", "image/avif"}
	acceptedImageTypes := []string{}

	if acceptHeader == "*/*" || acceptHeader == "" {
		acceptedImageTypes = imageTypes
	} else {
		for _, imageType := range imageTypes {
			if strings.Contains(acceptHeader, imageType) {
				acceptedImageTypes = append(acceptedImageTypes, imageType)
			}
		}
	}

	color.Green("acceptedImageTypes: " + strings.Join(acceptedImageTypes, ", "))

	// attrs["src"] = fmt.Sprintf("/.generated/images/image.png?url=%s&width=%d", url.QueryEscape(imageUrl), maxWidth)

	for _, option := range options {
		parts := strings.Split(option, "=")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Add key to order if it's a new key and not srcset or sizes
			if _, exists := attrs[key]; !exists && key != "srcset" && key != "sizes" {
				order = append(order, key)
			}

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

			// case "width":
			// 	customWidthProvided = true
			// 	attrs["width"] = value // Use the provided custom width
			// case "height":
			// 	customHeightProvided = true
			// 	attrs["height"] = value // Use the provided custom height

			default:
				attrs[key] = value // Handle other valid attributes
			}
		}
	}

	// // Generate default width if not provided
	// if !customWidthProvided {
	// 	attrs["width"] = "100%"
	// }

	// // Generate default height if not provided
	// if !customHeightProvided {
	// 	attrs["height"] = "auto"
	// }

	// Generate default sizes if not provided
	if !customSizesProvided {
		attrs["sizes"] = "(max-width: 600px) 100vw, (max-width: 1024px) 50vw, 25vw"
		order = append(order, "sizes") // Add sizes to the order
	}

	// Generate default srcset if not provided
	if !customSrcsetProvided {
		attrs["srcset"] = generateDefaultSrcset(imageUrl, maxWidth)
		order = append(order, "srcset") // Add srcset to the order
	}

	var b strings.Builder
	for _, key := range order {
		if value, ok := attrs[key]; ok {
			b.WriteString(fmt.Sprintf(`%s="%s" `, key, value))
		}
	}

	return template.HTMLAttr(b.String())
}

func generateDefaultSrcset(imageUrl string, maxWidth int) string {
	widths := generateWidths(maxWidth)
	srcsetValues := make([]string, len(widths))
	for i, width := range widths {
		optimizedImageUrl := fmt.Sprintf("/.generated/images/image.png?url=%s&width=%d", url.QueryEscape(imageUrl), width)
		srcsetValues[i] = fmt.Sprintf("%s %dw", optimizedImageUrl, width)
	}
	return strings.Join(srcsetValues, ", ")
}

func generateWidths(maxWidth int) []int {
	// Define a set of widths up to maxWidth
	// Example: [320, 480, 640, 800, 960, 1280, 1600, maxWidth]
	// Implement logic to generate this array based on maxWidth
	return []int{320, 480, 640, 768, 1024, 1280, 1600, maxWidth}
}

// simple in-memory cache
var optimizedImageCache = make(map[string]bool)

func optimizeImage(url string, width int) (string, error) {
	start := time.Now()
	cacheKey := fmt.Sprintf("%s-%d", url, width)

	// Check cache first
	if _, exists := optimizedImageCache[cacheKey]; exists {
		optimizedImagePath := getOptimizedImagePath(url, width)
		// fmt.Println("image already processed: ", optimizedImagePath)
		return optimizedImagePath, nil
	}
	// fmt.Println("image not processed yet: ", url)

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
