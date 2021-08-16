package main

import (
	"fmt"
	"os"
	"time"
)

const outDir = "out"

func main() {
	startTime := time.Now()

	if err := os.RemoveAll(outDir); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	handleFile("index.html")

	fmt.Println("Done in", time.Now().Sub(startTime))
}
