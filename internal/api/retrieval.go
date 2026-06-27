package api

import (
	"context"

	"github.com/Night-Swan/memex/internal/db"
	"github.com/Night-Swan/memex/internal/embed"
	"github.com/pgvector/pgvector-go"
)

func RetrieveRelevantChunks(ctx context.Context, query string, limit int) ([]string, error) {
	vec, err := embed.GenerateEmbedding(query)
	if err != nil {
		return nil, err
	}

	rows, err := db.Pool.Query(ctx,
		`SELECT content
		 FROM chunks
		 ORDER BY embedding <=> $1
		 LIMIT $2;`, pgvector.NewVector(vec), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []string{}
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		results = append(results, content)
	}

	return results, nil
}