-- ============================================================
-- 0002_rag.sql — 古籍/知识库向量检索(P3 阶段应用)
-- 依赖:pgvector 扩展(https://github.com/pgvector/pgvector)
-- 注意:embedding 维度以最终选定的 embedding 模型为准,此处按 1024 预设。
-- ============================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE corpus_embeddings (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    book_slug     TEXT        NOT NULL,
    paragraph_id  TEXT        NOT NULL UNIQUE,
    content       TEXT        NOT NULL,
    embedding     vector(1024),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_corpus_embeddings_hnsw
    ON corpus_embeddings USING hnsw (embedding vector_cosine_ops);

COMMIT;
