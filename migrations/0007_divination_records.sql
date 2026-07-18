-- ============================================================
-- 0007_divination_records.sql — 卦档:占卜记录存档
--   登录用户起卦自动存档;AI 解卦成功后回填 reading。
--   每用户保留最近 200 条,超出由应用层截断。
-- ============================================================

BEGIN;

CREATE TABLE IF NOT EXISTS divination_records (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind             TEXT        NOT NULL CHECK (kind IN ('meihua', 'liuyao')),
    question         TEXT        NOT NULL DEFAULT '',
    summary          TEXT        NOT NULL DEFAULT '',
    payload          JSONB       NOT NULL,
    reading          TEXT,
    reading_provider TEXT,
    cast_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_divination_records_user
    ON divination_records (user_id, created_at DESC);

COMMIT;
