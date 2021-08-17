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

	if !fileExists("index.html") {
		fmt.Println("An index.html entry file at to root of the project is required")
		os.Exit(0)
	}

	build()
	if len(os.Args) > 1 && (os.Args[1] == "watch" || os.Args[1] == "w") {
		watch()
	}
}

func build() {
	done = make(map[string]string) // Reset file URLs

	startTime := time.Now()
	handleFile("index.html")
	fmt.Printf("Done in %dms\n", time.Now().Sub(startTime).Milliseconds())
}
