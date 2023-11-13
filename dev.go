package main

import (
    "log"
	"os"
	"os/exec"
    "github.com/fsnotify/fsnotify"
	"path/filepath"
	"fmt"
	"io/ioutil"

	"github.com/fatih/color"
)

func main() {
	banner()
	color.Green("Starting file watcher")
	bundleCss() // bundle on startup
    // Create new watcher.
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    // Start listening for events.
    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                // log.Println("event:", event)
				
				// if event is CREATE, add to watcher so we can keep track of new subdirs
				if event.Op == fsnotify.Create {
					// check if its a dir
					fi, err := os.Stat(event.Name)
					if err != nil {
						log.Fatal(err)
					}
					mode := fi.Mode()
					if mode.IsDir() {
						log.Println("watching new dir:", event.Name)
						watcher.Add(event.Name)
					}
				}

                if event.Has(fsnotify.Write) {
                    fmt.Println("modified file:", event.Name)

					// check if file is a css file
					fi, err := os.Stat(event.Name)
					if err != nil {
						log.Fatal(err)
					}
					mode := fi.Mode()
					if !mode.IsDir() && filepath.Ext(event.Name) == ".css" {
						bundleCss()
					}
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("error:", err)
            }
        }
    }()

    // Add a path.
    err = watcher.Add("src/components")
    err = watcher.Add("src/styles")
	color.Blue("Watching directories:")
	// Recursively add all subdirs to watcher.Add
	filepath.Walk("src/components", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			color.Green(path)
			watcher.Add(path)
		}
		return nil
	})
	filepath.Walk("src/styles", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			color.Green(path)
			watcher.Add(path)
		}
		return nil
	})

    if err != nil {
        log.Fatal(err)
    }

    // Block main goroutine forever.
    <-make(chan struct{})
}

func bundleCss() {
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
		log.Fatal(err)
	}

	// write all other css files to bundle.css
	filepath.Walk("src/styles", func(path string, info os.FileInfo, err error) error {
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

	// run postcss-cli command
	cmd := exec.Command("./node_modules/.bin/postcss", "-o", ".generated/css/main.css", "tmp/bundle.css")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}


func banner() {
	c := color.New(color.FgCyan)
    b, err := ioutil.ReadFile("banner.txt")
    if err != nil {
        panic(err)
    }
    c.Println(string(b))
}