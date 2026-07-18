-- ============================================================
-- 0003_session_version.sql — 会话版本(JWT 秒级全端吊销)
--
-- P1 以 PostgreSQL 承载会话版本(带进程内 TTL 缓存);
-- 横向扩容后如需更低延迟,可平移到 Redis(接口已抽象)。
-- ============================================================

BEGIN;

ALTER TABLE users ADD COLUMN session_ver INT NOT NULL DEFAULT 1;

COMMIT;
