package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func watch() {
	logInfo("⏳ Watching for changes...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watchDone := make(chan bool)
	go func() {
		for {
			select {

			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}

				if ev.Op&fsnotify.Write == fsnotify.Write {
					build()
				} else if ev.Op&fsnotify.Create == fsnotify.Create {
					if ev.Name != outDir {
						watcher.Add(ev.Name)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				panic(err)
			}
		}
	}()

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if root := filepath.SplitList(path)[0]; root == outDir || root == "node_modules" {
			return filepath.SkipDir
		}
		if err = watcher.Add(path); err != nil {
			logFatal("Cannot add " + path + " to watcher: " + err.Error())
		}
		return nil
	})

	<-watchDone
}
