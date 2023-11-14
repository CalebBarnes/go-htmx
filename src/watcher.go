package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

type WatcherConfig struct {
	directories []string
	includedExt []string
	silenceLogs bool
}

var watcherConfig = WatcherConfig{
	// directories: []string{"src/templates", "src/components", "src/styles"},
	directories: []string{"src/templates", "src/components", "src/styles"},
	includedExt: []string{".html", ".css"},
	silenceLogs: false,
}

func watcher() {
	logMsg(color.GreenString("Watching directories..."))

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add directories to watcher and walk through them
	for _, dir := range watcherConfig.directories {
		// logMsg(color.BlueString("Watching directory: " + dir))
		addDirToWatcher(watcher, dir)
	}

	// Start listening for events.
	go watchForChanges(watcher)

	// Block main goroutine forever.
	<-make(chan struct{})
}

// watchForChanges handles file system events
func watchForChanges(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleEvent(event, watcher)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func handleEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
	// if event is CREATE, add to watcher so we can keep track of new subdirs
	if event.Op == fsnotify.Create {
		// check if its a dir
		fi, err := os.Stat(event.Name)
		if err != nil {
			log.Fatal(err)
		}
		mode := fi.Mode()
		if mode.IsDir() {
			logMsg(color.GreenString("added directory: " + event.Name))
			watcher.Add(event.Name)
		}

		// check if watcherConfig.includedExt
		for _, ext := range watcherConfig.includedExt {
			if filepath.Ext(event.Name) == ext {
				logMsg(color.MagentaString("created: " + event.Name))
				if ext == ".css" {
					go postCSS()
				} else {
					go postCSS()
				}
				return
			}
		}
	}

	if event.Op == fsnotify.Write {
		// check if file is a css file
		fi, err := os.Stat(event.Name)
		if err != nil {
			log.Fatal(err)
		}
		mode := fi.Mode()
		// skip directories
		if mode.IsDir() {
			return
		}

		for _, ext := range watcherConfig.includedExt {
			if filepath.Ext(event.Name) == ext {
				logMsg(color.MagentaString("update: " + event.Name))
				if ext == ".css" {
					go postCSS()
				} else {
					go postCSS()
				}
				return
			}
		}

		// if RENAME, log what it was renamed to
		if event.Op == fsnotify.Rename {
			logMsg(color.RedString("removed: " + event.Name))
		}

		// if RENAME, log what it was renamed to
		if event.Op == fsnotify.Rename {
			logMsg(color.RedString("removed: " + event.Name))
		}
	}
}

func addDirToWatcher(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	logMsg(color.BlueString(dir))
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// skip if same as top-level dir
			if path == dir {
				return nil
			}
			logMsg(color.BlueString("|-- " + path[len(dir)+1:]))
			watcher.Add(path)
		}
		return nil
	})
}

func postCSS() {
	start := time.Now()

	outputPath := "tmp/css/bundled.css"

	cmd := exec.Command(
		"./node_modules/.bin/postcss",
		"src/styles/main.css",
		"-o", outputPath,
		"--config", "postcss.config.js")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}

	logMsg(color.GreenString("PostCSS finished in %0.2fms", time.Since(start).Seconds()*1000))
	esbuildCSS()
}

func startBrowserSync() {
	cmd := exec.Command(
		"./node_modules/.bin/browser-sync", "start",
		"--proxy", "localhost:"+os.Getenv("PORT"),
		"--files",
		"'.generated/css, .generated/css/main.css, src/templates/*.html, src/components/**/*.html, src/styles/*.css'",
		"--plugins", "bs-html-injector?files[]=*.html",
		"--no-notify",
		"--no-open")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()

	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}
}

func esbuildCSS() {
	start := time.Now()

	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"tmp/css/bundled.css"},
		Bundle:            true,
		Outfile:           ".generated/css/main.css",
		Write:             true,
		Target:            api.ES2015,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
	})

	if len(result.Errors) > 0 {
		logMsg(color.RedString("Error esbuildCSS:"))
		os.Stderr.WriteString(result.Errors[0].Text)
		os.Exit(1)
	}

	logMsg(color.GreenString("Esbuild finished minifying CSS in %0.2fms", time.Since(start).Seconds()*1000))
}
