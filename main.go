package main

import (
	"fmt"
	"os"
	"time"
)

const outDir = "out"

func main() {
	if err := os.RemoveAll(outDir); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	build()
	if len(os.Args) > 1 && (os.Args[1] == "watch" || os.Args[1] == "w") {
		if len(os.Args) > 2 {
			serverAddr = os.Args[2]
		}
		serve()
		watch()
	}
}

func build() {
	if !fileExists("index.html") {
		logError("An index.html entry file is required at to root of the project")
		return
	}

	done = make(map[string]string) // Reset file URLs

	startTime := time.Now()
	handleFile("index.html")
	logSuccess(fmt.Sprintf("⚡️ SSG in %dms", time.Now().Sub(startTime).Milliseconds()))
}
