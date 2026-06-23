package main

import (
	"log"

	"github.com/Night-Swan/memex/internal/api"
	"github.com/Night-Swan/memex/internal/db"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to database successfully")

	handler := api.NewHandler()

	router := gin.Default()
	router.POST("/notes", handler.CreateNote)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}