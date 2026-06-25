package api

import (
	"net/http"

	"github.com/Night-Swan/memex/internal/db"
	"github.com/Night-Swan/memex/internal/embed"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Note struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) CreateNote(c *gin.Context) {
	var note Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vec, err := embed.GenerateEmbedding(note.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	docID := uuid.New()
	_, err = db.Pool.Exec(c.Request.Context(),
		"INSERT INTO documents (id, title, file_path, source_url) VALUES ($1, $2, $3, $4)",
		docID, note.Title, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chunkID := uuid.New()
	_, err = db.Pool.Exec(c.Request.Context(),
		"INSERT INTO chunks (id, document_id, chunk_index, page_number, content, embedding) VALUES ($1, $2, $3, $4, $5, $6)",
		chunkID, docID, 0, nil, note.Content, pgvector.NewVector(vec))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": docID})
}

func (h *Handler) SearchNotes(c *gin.Context) {
	var query struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vec, err := embed.GenerateEmbedding(query.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(),
		`SELECT content, embedding <=> $1 AS distance
		FROM chunks
		ORDER BY embedding <=> $1
		LIMIT 5;`, pgvector.NewVector(vec))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		var content string
		var distance float64
		if err := rows.Scan(&content, &distance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, map[string]interface{}{
			"content":  content,
			"distance": distance,
		})
	}

	c.JSON(http.StatusOK, results)
}