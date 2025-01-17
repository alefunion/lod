package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	copySpecialFiles()
	logSuccess(fmt.Sprintf("⚡️ SSG in %dms", time.Since(startTime).Milliseconds()))
}

// Copy potential special files to out directory that are generally not referenced in HTML but useful for the build
func copySpecialFiles() {
	for fn := range map[string]struct{}{
		".htaccess":  {},
		"robots.txt": {},
	} {
		from, err := os.Open(fn)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			panic(err)
		}
		defer from.Close()

		to, err := os.Create(filepath.Join(outDir, fn))
		if err != nil {
			panic(err)
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			panic(err)
		}
	}
}
