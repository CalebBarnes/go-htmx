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

const ImageBaseRoute = "/_image"

type ImageFormat string

const (
	FormatWebP ImageFormat = "webp"
	FormatPNG  ImageFormat = "png"
)

func getImageProps(headers http.Header, imageUrl string, options ...string) template.HTMLAttr {
	attrs := make(map[string]string)
	order := []string{"src"} // Start with src as the first key
	maxWidth := 1920         // Default maxWidth
	// customSrcsetProvided := false
	customSizesProvided := false

	imageFormat := getSupportedImageFormat(headers)

	attrs["src"] = fmt.Sprintf(ImageBaseRoute+"/image.%s?url=%s&width=%d", imageFormat, url.QueryEscape(imageUrl), maxWidth)

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
			case "layout":
				if value == "fill" {
					attrs["style"] = "position: absolute; height: 100%; width: 100%; inset: 0px; color: transparent;"
				}
				// switch value {
				// case "fill":
				// attrs["style"] = "object-fit: cover; object-position: center center;"
				// }

			case "style":
				attrs["style"] += " " + value
			// case "maxWidth":
			// var err error
			// maxWidth, err = strconv.Atoi(value)
			// if err != nil {
			// 	maxWidth = 1920 // Reset to default if conversion fails
			// }

			case "sizes":
				customSizesProvided = true
				attrs["sizes"] = value // Use the provided custom sizes
			default:
				attrs[key] = value // Handle other valid attributes
			}
		}
	}

	// Generate default sizes if not provided
	if !customSizesProvided {
		attrs["sizes"] = "(max-width: 768px) 100vw, 56rem"
		order = append(order, "sizes") // Add sizes to the order
	}

	attrs["srcset"] = generateSrcset(imageUrl, maxWidth, imageFormat)
	order = append(order, "srcset") // Add srcset to the order
	order = append(order, "style")
	attrs["decoding"] = "async"
	order = append(order, "decoding")

	var b strings.Builder
	for _, key := range order {
		if value, ok := attrs[key]; ok {
			b.WriteString(fmt.Sprintf(`%s="%s" `, key, value))
		}
	}

	return template.HTMLAttr(b.String())
}

func generateSrcset(imageUrl string, maxWidth int, format ImageFormat) string {
	widths := generateWidths(maxWidth)
	srcsetValues := make([]string, len(widths))
	for i, width := range widths {
		optimizedImageUrl := fmt.Sprintf(ImageBaseRoute+"/image.%s?url=%s&width=%d", format, url.QueryEscape(imageUrl), width)
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

func getSupportedImageFormat(headers http.Header) ImageFormat {
	return FormatPNG
	// acceptHeader := headers.Get("Accept")
	// if strings.Contains(acceptHeader, "image/webp") {
	// 	return FormatWebP
	// } else {
	// 	return FormatPNG // Default to PNG if no specific format is requested
	// }
}

func optimizeImage(url string, width int, format ImageFormat) (string, error) {
	start := time.Now()
	cacheKey := fmt.Sprintf("%s-%d", url, width)
	imgPath := getOptimizedImagePath(url, width, format)

	// Check cache first
	if _, exists := optimizedImageCache[cacheKey]; exists {
		// fmt.Println("image already processed and in cache: ", imgPath)
		return imgPath, nil
	}

	// not in cache yet, check if its in the file system before downloading it
	if _, err := os.Stat(imgPath); err == nil {
		// fmt.Println("image already processed: ", imgPath)
		optimizedImageCache[cacheKey] = true // Update the cache
		return imgPath, nil
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

	// make sure the directory exists (without the filename)
	err = os.MkdirAll(imgPath[:strings.LastIndex(imgPath, "/")], 0755)
	if err != nil {
		fmt.Println("failed to create directory: ", err)
		return "", err
	}

	switch format {
	case FormatWebP:
		color.Red("FormatWebP not supported yet")

		// buffer, err := bimg.Read("image.jpg")
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// }

		// newImage, err := bimg.NewImage(srcImage).Resize(800, 600)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// }

		// size, err := bimg.NewImage(newImage).Size()
		// if size.Width == 800 && size.Height == 600 {
		// 	fmt.Println("The image size is valid")
		// }

		// bimg.Write("new.jpg", newImage)

	case FormatPNG:
		dstImageFill := imaging.Resize(srcImage, width, 0, imaging.Lanczos)
		err = imaging.Save(dstImageFill, imgPath)
		if err != nil {
			fmt.Println("failed to save image: ", err)
			return "", err
		}

	}

	// Update the cache
	optimizedImageCache[cacheKey] = true
	// Log time duration
	duration := time.Since(start)
	fmt.Println("image processed: ", duration)
	return getOptimizedImagePath(url, width, format), nil
}

func getOptimizedImagePath(url string, width int, format ImageFormat) string {
	safeImageName := convertURLToFilePath(url)
	return fmt.Sprintf(".generated/images/%s/%dw.%s", safeImageName, width, format)
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

	format := getSupportedImageFormat(r.Header)

	optimizedImagePath, err := optimizeImage(url, width, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve the optimized image
	http.ServeFile(w, r, optimizedImagePath)
}
