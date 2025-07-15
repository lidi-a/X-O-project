package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/message", handleMessage)
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
