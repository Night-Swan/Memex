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

type Profile struct {
	WeightKg      float64 `json:"weight_kg"`
	HeightCm      float64 `json:"height_cm"`
	Age           int     `json:"age"`
	Sex           string  `json:"sex"`
	ActivityLevel string  `json:"activity_level"`
	Goals         string  `json:"goals"`
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


func (h *Handler) SetProfile(c *gin.Context) {
	var profile Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Pool.Exec(c.Request.Context(),
		"DELETE FROM profile;")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	profileID := uuid.New()
	_, err = db.Pool.Exec(c.Request.Context(),
		`INSERT INTO profile (id, weight_kg, height_cm, age, sex, activity_level, goals)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		profileID, profile.WeightKg, profile.HeightCm, profile.Age, profile.Sex, profile.ActivityLevel, profile.Goals)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": profileID})
}