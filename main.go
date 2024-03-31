package main

import (
	"context"
	"log"
	"net/http"
)

var port = ":8000"

func main() {
	manager := NewManager(context.Background())

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serveWS)
	http.HandleFunc("/login", manager.loginHandler)

	log.Println("Running server on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
