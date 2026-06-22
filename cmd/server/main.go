package main

import (
	"log"

	"github.com/Night-Swan/memex/internal/db"
	"github.com/Night-Swan/memex/internal/embed"
)

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to database successfully")

	vec, err := embed.GenerateEmbedding("test")
	if err != nil {
		log.Fatalf("embedding failed: %v", err)
	}
	log.Printf("got embedding with %d dimensions", len(vec))
}