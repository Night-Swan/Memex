CREATE TABLE documents (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    file_path TEXT,
    source_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chunks (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL
        REFERENCES documents(id)
        ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    page_number INTEGER,
    content TEXT NOT NULL,
    embedding VECTOR(768) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chunks_document_chunk_unique
        UNIQUE (document_id, chunk_index)
);

CREATE INDEX chunks_document_id_idx
ON chunks (document_id);

CREATE INDEX chunks_embedding_hnsw_idx
ON chunks
USING hnsw (embedding vector_cosine_ops);