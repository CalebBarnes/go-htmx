package main

import (
	"fmt"
	"os"

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
}

func logMsg(str string) {
	if watcherConfig.silenceLogs {
		return
	}
	prefixStr := color.MagentaString("[watcher] ")
	fmt.Println(prefixStr + str)
}
