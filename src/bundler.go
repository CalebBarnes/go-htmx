package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fatih/color"
)

func bundleAssets() {
	start := time.Now()

	postCSS("src/styles/main.css", "tmp/postcss/bundled.css")

	// esbuild -> bundles js and any css that is imported in ts
	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"src/ts/app.ts"},
		Bundle:            true,
		Outfile:           "tmp/esbuild/ts/app.js",
		Write:             true,
		Target:            api.ESNext,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			bundlerLogger(color.RedString("Error bundleAssets:"))
			os.Stderr.WriteString(err.Text)
		}
		return
	}

	// combine postcss and esbuild css with cat command
	cmd := exec.Command("bash", "-c", "cat ./tmp/postcss/bundled.css ./tmp/esbuild/ts/app.css > ./tmp/bundled.css")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}

	// minify the bundled css
	result = api.Build(api.BuildOptions{
		EntryPoints:       []string{"tmp/bundled.css"},
		Bundle:            true,
		Outfile:           ".generated/css/main.css",
		Write:             true,
		Target:            api.ES2015,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
	})

	if len(result.Errors) > 0 {
		bundlerLogger(color.RedString("Error esbuildCSS:"))
		os.Stderr.WriteString(result.Errors[0].Text)
	}

	if _, err := os.Stat(".generated/js"); os.IsNotExist(err) {
		os.Mkdir(".generated/js", 0755)
	}

	err := os.Rename("tmp/esbuild/ts/app.js", ".generated/js/app.js")
	if err != nil {
		bundlerLogger(color.RedString("Error moving app.js:", err))
		return
	}

	bundlerLogger(fmt.Sprintf("assets bundled in %0.2fms", time.Since(start).Seconds()*1000))
}

func postCSS(inputPath string, outputPath string) {
	cmd := exec.Command(
		"./node_modules/.bin/postcss",
		inputPath,
		"-o", outputPath,
		"--config", "postcss.config.js")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		color.Red(fmt.Sprint(cmdErr) + ": " + stderr.String())
		return
	}
}
