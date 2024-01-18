package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func pageFound(pageData Page, w http.ResponseWriter, r *http.Request) {
	versionHash := getVersionHash()

	data := map[string]interface {
	}{
		"Env":     os.Getenv("APP_ENV"),
		"Version": versionHash,
		"Seo": Seo{
			Title:       pageData.Title,
			Description: "This is the SEO description",
		},
		"Data": pageData,
	}

	templateName := getTemplateName(pageData.Template)

	tmpl, err := bootstrapTemplate(r, pageData)
	if err != nil {
		log.Fatalf("Error bootstrapping template: %v", err)
	}

	tmpl = template.Must(tmpl.ParseFiles("src/templates/" + templateName + ".go.html"))

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

	tmpl, err := bootstrapTemplate(r, Page{})
	if err != nil {
		log.Fatalf("Error bootstrapping template: %v", err)
	}
	tmpl = template.Must(tmpl.ParseFiles("src/templates/404.go.html"))

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

// Bootstrap the template with some global functions and components
func bootstrapTemplate(r *http.Request, pageData Page) (*template.Template, error) {
	getImagePropsWithContext := func(imageUrl string, otherParams ...string) template.HTMLAttr {
		return getImageProps(r.Header, imageUrl, otherParams...)
	}

	getImageByIdWithContext := func(id string, otherParams ...string) template.HTMLAttr {
		imageUrl := os.Getenv("DIRECTUS_URL") + "/assets/" + id
		return getImageProps(r.Header, imageUrl, otherParams...)
	}

	tmpl := template.Must(template.ParseFiles("src/templates/layout.go.html"))

	tmpl.Funcs(template.FuncMap{
		// Render HTML in a template without escaping it (or any other strings)
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},

		"imageProps":         getImagePropsWithContext,
		"directusImageProps": getImageByIdWithContext,
	})

	templateName := getTemplateName(pageData.Template)
	var autoLoadBodyStr string
	var autoLoadHeadStr string

	if fileExists(".generated/esbuild/templates/" + templateName + ".js") {
		autoLoadBodyStr += `
		<script defer src="/bundle/` + templateName + `.js?v={{.Version}}"></script>
		`
	}
	if fileExists(".generated/esbuild/templates/" + templateName + ".css") {
		autoLoadBodyStr += `
		<link rel="stylesheet" href="/bundle/` + templateName + `.css?v={{.Version}}" />
		`
	}
	if fileExists(".generated/esbuild/templates/app.js") {
		autoLoadHeadStr += `
		<script defer src="/bundle/app.js?v={{.Version}}"></script>
		`
	}

	tmpl.Parse(`
	{{ define "autoload_head" }}
		` + autoLoadHeadStr + `
	{{ end }}
	`)
	tmpl.Parse(`
	{{ define "autoload_body_scripts" }}
		` + autoLoadBodyStr + `
	{{ end }}
	`)

	// Add all templates in the components folder
	err := appendTemplates(tmpl, "src/components", ".go.html")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// Build the blocks template from the blocks in the page data
func blocksTemplateBuilder(blocks []Block) string {
	blockBuilderStr := `
	{{ define "blocks" }}
		{{ range .Data.Blocks }}
			{{ if eq .Collection "a" }}
				`

	for _, block := range blocks {
		// Check if template file exists
		blockFileName := strings.Replace(block.Collection, "block_", "", 1)
		if fileExists("src/components/blocks/" + blockFileName + ".go.html") {
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

func getTemplateName(templateName string) string {
	if templateName == "" || templateName == "default" {
		templateName = "index"
	}
	return templateName
}
