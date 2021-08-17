package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func watch() {
	fmt.Println("Watching for changes...")

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
						fmt.Println("add to watcher:", ev.Name)
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
		if filepath.SplitList(path)[0] == outDir {
			return filepath.SkipDir
		}
		if err = watcher.Add(path); err != nil {
			panic(err)
		}
		return nil
	})

	<-watchDone
}
