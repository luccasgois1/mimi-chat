package main

import (
	"fmt"
	"net/http"
)

type ServerConfiguration struct {
	URL  string
	PORT uint
}

var SERVER_CONFIGURATION ServerConfiguration = ServerConfiguration{"localhost", 8080}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to mimi-chat!")
}

func main() {
	http.HandleFunc("/", homeHandler)
	fmt.Printf("Server started at http://%s:%d\n", SERVER_CONFIGURATION.URL, SERVER_CONFIGURATION.PORT)
    portToListen := fmt.Sprintf(":%d", SERVER_CONFIGURATION.PORT)
	http.ListenAndServe(portToListen, nil)
}
