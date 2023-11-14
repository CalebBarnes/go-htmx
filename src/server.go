package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func server() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pageData, err := getPageData(r.URL.Path)
		if err != nil {
			notFound(w, r)
		} else {
			pageTemplate(pageData, w, r)
		}
	})

	// Handler for generated CSS files
	cssFileServer := http.FileServer(http.Dir(".generated/css"))
	mux.Handle("/css/", maxAgeHandler(15552000, http.StripPrefix("/css/", cssFileServer)))

	// Handler for static files like favicon, robots etc
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", maxAgeHandler(15552000, http.StripPrefix("/static/", fileServer)))

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	// Handler for image optimization
	mux.Handle(ImageBaseRoute+"/", http.HandlerFunc(imageHandler))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}

func pageTemplate(pageData Page, w http.ResponseWriter, r *http.Request) {
	versionHash := getVersionHash()

	data := map[string]interface {
	}{
		"Version": versionHash,
		"Seo": Seo{
			Title:       pageData.Title,
			Description: "This is the SEO description",
		},
		"Data": pageData,
	}

	getImagePropsWithContext := func(imageUrl string, otherParams ...string) template.HTMLAttr {
		return getImageProps(r.Header, imageUrl, otherParams...)
	}

	getImageByIdWithContext := func(id string, otherParams ...string) template.HTMLAttr {
		imageUrl := os.Getenv("DIRECTUS_URL") + "/assets/" + id
		return getImageProps(r.Header, imageUrl, otherParams...)
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
		"imageProps":         getImagePropsWithContext,
		"directusImageProps": getImageByIdWithContext,
	})

	err = appendTemplates(tmpl, "src/components", ".go.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	tmpl, err = tmpl.Parse(blocksTemplateBuilder(pageData.Blocks))
	if err != nil {
		log.Fatalf("Error parsing block templates: %v", err)
	}

	if version == "production" {
		w.Header().Add("Cache-Control", fmt.Sprintf("private, max-age=%d stale-while-revalidate=%d", 60, 86400))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Error executing template:", err)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	versionHash := getVersionHash()

	tmpl := template.Must(template.ParseFiles("src/templates/404.go.html"))
	template.Must(tmpl.ParseGlob("src/components/*.go.html"))
	data := map[string]interface{}{
		"Version": versionHash,
		"Seo": Seo{
			Title:       "404 - Page not found",
			Description: "You've hit a dead end...",
		},
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Set the HTTP status code to 404
	w.WriteHeader(http.StatusNotFound)
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing 404 template: %v", err)
		return
	}
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

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", seconds))
		h.ServeHTTP(w, r)
	})
}

var version string

func getVersionHash() string {
	var versionHash string

	if os.Getenv("APP_ENV") == "development" {
		versionHash = strconv.FormatInt(time.Now().UnixNano(), 10)
	} else if version == "" {
		versionHash = version
	}

	return versionHash
}

func appendTemplates(tmpl *template.Template, rootDir, suffix string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, suffix) {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
