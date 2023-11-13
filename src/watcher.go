package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

func watcher() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	// Define top-level directories to watch relative to current working directory (includes subdirectories recursively)
	directories := []string{"src/components", "src/styles"}

	// Add directories to watcher and walk through them
	for _, dir := range directories {
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
			color.Green("added directory: " + event.Name)
			watcher.Add(event.Name)
		}

		// check if file is a css file
		if !mode.IsDir() && filepath.Ext(event.Name) == ".css" {
			color.Green("CSS file created, bundling...")
			go bundleCss()
		}
	}

	if event.Op == fsnotify.Write {
		// check if file is a css file
		fi, err := os.Stat(event.Name)
		if err != nil {
			log.Fatal(err)
		}
		mode := fi.Mode()
		if !mode.IsDir() && filepath.Ext(event.Name) == ".html" {
			go bundleCss()
		} else if !mode.IsDir() && filepath.Ext(event.Name) == ".css" {
			go bundleCss()
		} else {
			color.Green("file changed: " + event.Name)
		}
	}

	// if RENAME, log what it was renamed to
	if event.Op == fsnotify.Rename {
		color.Red("removed: " + event.Name)
	}

}

func addDirToWatcher(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	color.Blue(dir)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// skip if same as top-level dir
			if path == dir {
				return nil
			}
			color.Blue("|-- " + path[len(dir)+1:])
			watcher.Add(path)
		}
		return nil
	})
}

func bundleCss() {
	color.Green("Bundling CSS...")
	// open file for writing
	f, err := os.Create("./tmp/bundle.css")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// write global.css to bundle.css
	globalCss, err := os.ReadFile("src/styles/global.css")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write(globalCss)
	if err != nil {
		// log.Fatal(err)
		color.Red("Error writing global.css to bundle.css")
		println(err.Error())
	}

	// write all other css files to bundle.css
	filepath.Walk("src/styles", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			color.Red("Error walking src/styles directory")
			log.Fatal(err)
		}

		// skip directories
		if info.IsDir() {
			return nil
		}
		// skip global.css, as this file is added at the top of the file
		if info.Name() == "global.css" {
			return nil
		}
		// read file
		css, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		// write file to bundle.css
		_, err = f.Write(css)
		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	// measure how long this command takes to execute
	start := time.Now()

	cmd := exec.Command("./node_modules/.bin/postcss", "-o", ".generated/css/main.css", "tmp/bundle.css")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}

	duration := time.Since(start)

	color.Green("Bundled CSS in " + duration.String())
}

func startBrowserSync() {
	cmd := exec.Command("./node_modules/.bin/browser-sync", "start", "--proxy", "localhost:"+os.Getenv("PORT"), "--files", "'static/css/*.css, src/templates/*.html, src/components/**/*.html, *.go'", "--no-notify", "--plugins", "bs-html-injector?files[]=*.html", "--no-open")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}
}
