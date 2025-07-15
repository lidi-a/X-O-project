package main

import (
	"log"
	"net/http"
)

func main() {

	cache := NewInMemoryCache()

	handler, err := NewHandler(cache)
	if err != nil {
		log.Fatalf("Failed to create handler: %s", err)

	}

	http.HandleFunc("/message", handler.HandleMessage)
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
