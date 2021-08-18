package main

import (
	"net/http"
)

var serverAddr = ":8080"

// Serve starts a file server in a new goroutine and waits until it's ready
func serve() {
	ready := make(chan bool)
	go startServer(ready)
	<-ready
}

func startServer(ready chan bool) {
	http.Handle("/", http.FileServer(http.Dir(outDir)))

	host := serverAddr
	if host[0] == ':' {
		host = "localhost" + serverAddr
	}
	logInfo("🌐 Served at http://" + host)

	ready <- true
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		logFatal("Server error: ", err)
	}
}
