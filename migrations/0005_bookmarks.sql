-- ============================================================
-- 0005_bookmarks.sql — 古籍阅读书签(丝滑体验:随手收藏 + 跨端同步)
--
--   阅读进度表(reading_progress)已在 0001 建立,本迁移只新增书签。
--   一段一书签:UNIQUE(user_id, book_slug, chapter_idx, paragraph_id)
--   使加书签天然幂等;列表按创建时间倒序,复合索引直接命中。
-- ============================================================

BEGIN;

CREATE TABLE reading_bookmarks (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_slug    TEXT        NOT NULL,
    chapter_idx  INT         NOT NULL,
    paragraph_id TEXT        NOT NULL DEFAULT '',
    excerpt      TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, book_slug, chapter_idx, paragraph_id)
);

-- 「我的书签」列表:用户维度按时间倒序,复合索引覆盖排序。
CREATE INDEX idx_reading_bookmarks_user ON reading_bookmarks(user_id, created_at DESC);

COMMIT;
