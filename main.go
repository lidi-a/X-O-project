package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {

	cache := NewInMemoryCache()
	ctx := context.Background()

	go cache.cleanupLoop(ctx, time.Minute, 30*time.Minute) // Каждую минуту удаляем неактивные более 30 минут игры

	handler, err := NewHandler(cache)
	if err != nil {
		log.Fatalf("Failed to create handler: %s", err)

	}

	http.HandleFunc("/message", handler.HandleMessage)
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
