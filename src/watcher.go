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
}

var watcherConfig = WatcherConfig{
	directories: []string{
		"src/ts",
		"src/custom-elements",
		"src/templates",
		"src/components",
		"src/styles",
	},
	includedExt: []string{".html", ".css", ".ts"},
}

// * src/templates/*.html, src/components/**/*.html, src/styles/*.css

func logMsg(str string) {
	prefixStr := color.MagentaString("[watcher] ")
	fmt.Println(prefixStr + str)
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

				go postCSS()
				if ext == ".ts" {
					go bundleJs()
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
				go postCSS()
				if ext == ".ts" {
					go bundleJs()
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

	if _, err := os.Stat(".generated/css/main.css"); err == nil {
		// file exists
		println("file exists at .generated/css/main.css")
		// read contents
		contents, err := os.ReadFile(".generated/css/main.css")
		if err == nil {

			f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				println("error opening file")
				log.Println(err)
			}
			println("opened file " + outputPath)
			defer f.Close()

			_, err = f.Write(contents)
			if err != nil {
				log.Println(err)
			} else {
				println("wrote contents to " + outputPath)
			}

		}

	}

	esbuildCSS()
	logMsg(color.GreenString("CSS generated %0.2fms", time.Since(start).Seconds()*1000))
}

func esbuildCSS() {
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
	}
}

func startBrowserSync() {
	filesToWatch := "src"

	cmdArgs := []string{"start",
		"--proxy", "localhost:" + os.Getenv("PORT"),
		"--files",
		filesToWatch,
		"--plugins", "bs-html-injector?files[]=*.html",
		"--no-notify",
		"--no-open",
		"--port", "3000",
		"--ui-port", "3001",
	}

	cmd := exec.Command(
		"./node_modules/.bin/browser-sync", cmdArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()

	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}
}

func bundleJs() {
	start := time.Now()

	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"src/ts/app.ts"},
		Bundle:            true,
		Outfile:           ".generated/esbuild/bundle.js",
		Write:             true,
		Target:            api.ESNext,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			logMsg(color.RedString("Error bundleJs:"))
			os.Stderr.WriteString(err.Text)
		}

	}

	logMsg(color.GreenString("JS generated %0.2fms", time.Since(start).Seconds()*1000))
}
