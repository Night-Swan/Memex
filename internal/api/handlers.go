package api

import (
	"net/http"

	"github.com/Night-Swan/memex/internal/db"
	"github.com/Night-Swan/memex/internal/embed"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"strconv"
	"github.com/jackc/pgx/v5"
	"errors"
)

type Note struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type Handler struct {
}

type Profile struct {
	WeightKg float64 `json:"weight_kg"`
	HeightCm float64 `json:"height_cm"`
	Age      int     `json:"age"`
	Sex      string  `json:"sex"`
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
		`INSERT INTO profile (id, weight_kg, height_cm, age, sex)
		 VALUES ($1, $2, $3, $4, $5)`,
		profileID, profile.WeightKg, profile.HeightCm, profile.Age, profile.Sex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": profileID})
}

func (h *Handler) AskQuestion(c *gin.Context) {
	var question struct {
		Question string `json:"question" binding:"required"`
	}
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chunks, err := RetrieveRelevantChunks(c.Request.Context(), question.Question, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var profile Profile

	err = db.Pool.QueryRow(c.Request.Context(),
		"SELECT weight_kg, height_cm, age, sex FROM profile LIMIT 1").Scan(
		&profile.WeightKg, &profile.HeightCm, &profile.Age, &profile.Sex)

	hasProfile := true

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			hasProfile = false
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	prompt := buildPrompt(question.Question, chunks, hasProfile, profile)

	answer, err := embed.GenerateText(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}

func buildPrompt(question string, chunks []string, hasProfile bool, profile Profile) string {
	prompt := "You are a personal assistant that answers questions using ONLY the information provided below. "
	prompt += "Do not use any outside knowledge, even if you know more about the topic. "
	prompt += "If the provided information is insufficient to answer the question, say so explicitly instead of adding information from elsewhere.\n\n"

	if hasProfile {
		prompt += "User Profile:\n"
		prompt += "Weight: " + strconv.FormatFloat(profile.WeightKg, 'f', -1, 64) + " kg\n"
		prompt += "Height: " + strconv.FormatFloat(profile.HeightCm, 'f', -1, 64) + " cm\n"
		prompt += "Age: " + strconv.Itoa(profile.Age) + "\n"
		prompt += "Sex: " + profile.Sex + "\n\n"
	}

	if len(chunks) == 0 {
		prompt += "Relevant Information: None found.\n\n"
	} else {
		prompt += "Relevant Information from the user's own notes:\n"
		for i, chunk := range chunks {
			prompt += "Chunk " + strconv.Itoa(i+1) + ": " + chunk + "\n"
		}
		prompt += "\n"
	}

	prompt += "Question: " + question + "\n"
	prompt += "Answer using ONLY the chunks above. If they don't contain the answer, say: \"I don't have enough information in your notes to answer this.\""

	return prompt
}

