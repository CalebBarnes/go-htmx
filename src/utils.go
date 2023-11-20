package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func banner() {
	str := `
  _____          __    _       _______             __
 / ___/__  ___  / /__ (_)__   / ___/ /__  __ _____/ /
/ /__/ _ \/ _ \/  '_// / -_) / /__/ / _ \/ // / _  / 
\___/\___/\___/_/\_\/_/\__/  \___/_/\___/\_,_/\_,_/  
`
	c := color.New(color.FgCyan)
	c.Println(str)
	println(
		color.HiCyanString(`	⚡️Cookie Go 1.0.0`) + "\n" +
			`	- Server started at http://localhost:` + os.Getenv("PORT") + "\n" +
			`	- BrowserSync proxy started at http://localhost:3000` + "\n" +
			`	- Environment: ` + os.Getenv("APP_ENV") + "\n")
}

type LoggerConfig struct {
	prefix string
	str    string
}

func logger(config LoggerConfig) {
	println(config.prefix + config.str)
}

func watcherLogger(str string) {
	logger(LoggerConfig{
		prefix: color.GreenString("⚡️[watcher] "),
		str:    color.HiBlueString(str),
	})
}

func bundlerLogger(str string) {
	logger(LoggerConfig{
		prefix: color.MagentaString("⚡️[bundler] "),
		str:    color.YellowString(str),
	})
}

func serverLogger(str string) {
	logger(LoggerConfig{
		prefix: color.HiGreenString("⚡️[server] "),
		str:    color.HiCyanString(str),
	})
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
