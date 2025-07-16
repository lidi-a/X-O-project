package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lidi-a/X-O-project/internal"
)

func main() {

	//cache := NewInMemoryCache()
	//ctx := context.Background()
	//go cache.cleanupLoop(ctx, time.Minute, 30*time.Minute) // Каждую минуту удаляем неактивные более 30 минут игры

	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, relying on environment variables")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	cache := internal.NewRedisCache(redisAddr, os.Getenv("REDIS_PASS"), 0, 30*time.Minute, 3*time.Second)

	handler, err := internal.NewHandler(cache)
	if err != nil {
		log.Fatalf("Failed to create handler: %s", err)

	}

	http.HandleFunc("/message", handler.HandleMessage)
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
