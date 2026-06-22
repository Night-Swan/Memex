package main

import (
	"log"

	"github.com/Night-Swan/memex/internal/db"
)

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to database successfully")
}