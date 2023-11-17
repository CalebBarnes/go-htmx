package main

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var globs = []string{
	"src/**/*",
	"src/**/**/*",
}

var exts = []string{
	".html",
	".css",
	".ts",
}

func watcher() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	paths := []string{}
	// get all file paths from globs
	for _, glob := range globs {
		filePaths, _ := filepath.Glob(glob)
		paths = append(paths, filePaths...)
	}
	// add all paths to watcher
	for _, path := range paths {
		watcher.Add(path)
	}
	// Start listening for events.
	go watchForChanges(watcher)

	// Block main goroutine forever.
	<-make(chan struct{})
}

// watchForChanges handles file system events
func watchForChanges(watcher *fsnotify.Watcher) {
	watcherLogger("watching for file changes...")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Create {
				// watcherLogger("created file: " + event.Name)
				handleEvent(event, watcher)
			}

			if event.Op == fsnotify.Write {
				// watcherLogger("updated file: " + event.Name)
				handleEvent(event, watcher)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func handleEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
	fileExt := filepath.Ext(event.Name)
	for _, ext := range exts {
		if fileExt == ext {
			watcher.Add(event.Name)
			bundleAssets()
			return
		}
	}
}
