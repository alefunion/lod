package main

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func isFilepathRelative(path string) bool {
	if filepath.IsAbs(path) {
		return false
	}
	if _, err := url.ParseRequestURI(path); err == nil {
		return false
	}
	return true
}

// smallHash returns an 8-byte hash of the reader content.
func smallHash(r io.Reader) string {
	hasher := md5.New()
	if _, err := io.Copy(hasher, r); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))[24:]
}

func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return !os.IsNotExist(err)
}

// Logger

func logInfo(v ...interface{}) {
	log.Println(color.New(color.FgBlue).Sprintln(v...))
}

func logSuccess(v ...interface{}) {
	log.Println(color.New(color.FgGreen).Sprintln(v...))
}

func logWarning(v ...interface{}) {
	log.Println("⚠️  ", color.New(color.FgYellow).Sprintln(v...))
}

func logError(v ...interface{}) {
	log.Println("⭕️ ", color.New(color.FgRed).Sprintln(v...))
}

func logFatal(v ...interface{}) {
	log.Fatal("❌ ", color.New(color.FgRed).Sprintln(v...))
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"link": func(fp string) string {
			return processSource(fp)
		},

		"favicon": func(path string) template.HTML {
			return template.HTML(`<link rel="icon" href="` + path + `">`)
		},

		"script": func(path string) template.HTML {
			if isFilepathRelative(path) {
				return template.HTML(`<link rel="preload" href="` + path + `" as="script"><script defer src="` + path + `"></script>`)
			}
			url, err := url.Parse(path)
			if err != nil {
				return template.HTML(`<script defer src="` + path + `"></script>`)
			}
			return template.HTML(`<link rel="preconnect" href="` + url.Scheme + `://` + url.Hostname() + `"><script defer src="` + path + `"></script>`)
		},

		"style": func(path string) template.HTML {
			if isFilepathRelative(path) {
				return template.HTML(`<link rel="preload" href="` + path + `" as="style"><link rel="stylesheet" href="` + path + `">`)
			}
			url, err := url.Parse(path)
			if err != nil {
				return template.HTML(`<link rel="stylesheet" href="` + path + `">`)
			}
			return template.HTML(`<link rel="preconnect" href="` + url.Scheme + `://` + url.Hostname() + `"><link rel="stylesheet" href="` + path + `">`)
		},
	}
}
