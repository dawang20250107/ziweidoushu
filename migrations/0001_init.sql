-- ============================================================
-- 0001_init.sql — 平台核心 Schema(PostgreSQL 16)
--
-- 设计原则:
--   1. users 只存身份;登录凭证独立成表(手机号/微信可并存、可增删);
--   2. 支付走「订单 → 权益发放」状态机,所有外部回调幂等;
--   3. 订阅为周期购买制(非自动扣款),到期由 entitlements 判定;
--   4. 命盘档案/对话存 JSONB 快照,排盘引擎升级不影响历史数据;
--   5. 所有表带 created_at/updated_at,软删除仅用于用户可见数据。
-- ============================================================

-- 零扩展依赖:gen_random_uuid() 为 PG13+ 内置;
-- RAG 向量表(需 pgvector)在 0002_rag.sql,P3 阶段启用。

BEGIN;

-- ── 用户与身份 ────────────────────────────────────────────────

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname      TEXT        NOT NULL DEFAULT '',
    avatar_url    TEXT        NOT NULL DEFAULT '',
    -- 会员层级:free / pro / master(由 entitlements 派生,此处为缓存字段)
    tier          TEXT        NOT NULL DEFAULT 'free'
                  CHECK (tier IN ('free', 'pro', 'master')),
    tier_expires_at TIMESTAMPTZ,          -- 当前层级到期时间(free 为 NULL)
    status        TEXT        NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active', 'banned', 'deleted')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 登录凭证:一个用户可绑定多种方式(手机号 / 微信),便于后续加渠道。
CREATE TABLE user_identities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- provider: phone / wechat(海外二期:apple / google)
    provider      TEXT        NOT NULL CHECK (provider IN ('phone', 'wechat', 'apple', 'google')),
    -- identifier: 手机号(E.164)/ 微信 unionid(打通 Web 与小程序)
    identifier    TEXT        NOT NULL,
    -- 渠道附加凭证:微信存 {openid_web, openid_mp, session_key?};手机号为空对象
    credentials   JSONB       NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, identifier)
);
CREATE INDEX idx_identities_user ON user_identities(user_id);

-- 短信验证码(防刷主要靠 Redis 计数,这里存审计与状态)
CREATE TABLE sms_codes (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    phone         TEXT        NOT NULL,
    code_hash     TEXT        NOT NULL,           -- 哈希存储,不存明文
    purpose       TEXT        NOT NULL CHECK (purpose IN ('login', 'bind')),
    expires_at    TIMESTAMPTZ NOT NULL,
    consumed_at   TIMESTAMPTZ,
    client_ip     INET,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sms_phone_created ON sms_codes(phone, created_at DESC);

-- 刷新令牌(旋转失效;访问令牌为无状态 JWT,吊销靠 Redis 会话版本)
CREATE TABLE refresh_tokens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash    TEXT        NOT NULL UNIQUE,    -- SHA-256,不存明文
    device_info   TEXT        NOT NULL DEFAULT '',
    expires_at    TIMESTAMPTZ NOT NULL,
    rotated_from  UUID,                           -- 旋转链,复用检测
    revoked_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_user ON refresh_tokens(user_id) WHERE revoked_at IS NULL;

-- ── 命盘档案(数据锁定核心)──────────────────────────────────

CREATE TABLE chart_profiles (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label         TEXT        NOT NULL DEFAULT '',   -- 「本人」「母亲」「客户A」…
    relation      TEXT        NOT NULL DEFAULT 'self',
    -- 排盘输入(最小事实):公历生辰 + 性别 + 真太阳时参数
    birth_input   JSONB       NOT NULL,
    -- 命盘快照:引擎输出全量 JSON + engine_version,历史盘不随引擎升级漂移
    chart_snapshot JSONB      NOT NULL,
    engine_version TEXT       NOT NULL,
    is_default    BOOLEAN     NOT NULL DEFAULT false,
    deleted_at    TIMESTAMPTZ,                       -- 软删除
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_profiles_user ON chart_profiles(user_id) WHERE deleted_at IS NULL;
-- 每用户仅一个默认档案
CREATE UNIQUE INDEX idx_profiles_default ON chart_profiles(user_id) WHERE is_default AND deleted_at IS NULL;

-- ── AI 对话(粘性核心:带命盘上下文的持续咨询)────────────────

CREATE TABLE conversations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id    UUID        REFERENCES chart_profiles(id) ON DELETE SET NULL,
    title         TEXT        NOT NULL DEFAULT '',
    deleted_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_conversations_user ON conversations(user_id, updated_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE messages (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id UUID      NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role          TEXT        NOT NULL CHECK (role IN ('user', 'assistant')),
    content       TEXT        NOT NULL,
    -- assistant 消息记录:模型、token 用量、引用的古籍段落 id 等
    meta          JSONB       NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_messages_conv ON messages(conversation_id, id);

-- AI 用量配额(按自然日;层级限额配置在应用层)
CREATE TABLE ai_usage_daily (
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date    DATE        NOT NULL,
    request_count INT         NOT NULL DEFAULT 0,
    token_input   BIGINT      NOT NULL DEFAULT 0,
    token_output  BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date)
);

-- ── 商品 / 订单 / 权益(周期购买制)──────────────────────────

CREATE TABLE products (
    id            TEXT PRIMARY KEY,                 -- 'pro_1m' / 'pro_12m' / 'master_1m' …
    title         TEXT        NOT NULL,
    tier          TEXT        NOT NULL CHECK (tier IN ('pro', 'master')),
    duration_days INT         NOT NULL,
    price_cents   INT         NOT NULL,             -- 人民币分
    original_price_cents INT,                       -- 划线价
    active        BOOLEAN     NOT NULL DEFAULT true,
    sort_order    INT         NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no      TEXT        NOT NULL UNIQUE,      -- 对外单号(商户侧,含日期前缀)
    user_id       UUID        NOT NULL REFERENCES users(id),
    product_id    TEXT        NOT NULL REFERENCES products(id),
    amount_cents  INT         NOT NULL,
    -- 状态机:created → paying → paid → fulfilled;超时/取消 → closed;退款 → refunded
    status        TEXT        NOT NULL DEFAULT 'created'
                  CHECK (status IN ('created', 'paying', 'paid', 'fulfilled', 'closed', 'refunded')),
    channel       TEXT        CHECK (channel IN ('wechat', 'alipay')),
    -- 渠道交易号与回调原文(审计/对账)
    channel_txn_id TEXT,
    channel_payload JSONB     NOT NULL DEFAULT '{}',
    paid_at       TIMESTAMPTZ,
    fulfilled_at  TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ NOT NULL,             -- 未支付订单超时关闭
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_status ON orders(status) WHERE status IN ('created', 'paying', 'paid');
CREATE UNIQUE INDEX idx_orders_channel_txn ON orders(channel, channel_txn_id) WHERE channel_txn_id IS NOT NULL;

-- 权益发放流水:订单履约的结果,tier 有效期 = 现有效期顺延 duration
CREATE TABLE entitlements (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id       UUID        NOT NULL REFERENCES users(id),
    order_id      UUID        NOT NULL REFERENCES orders(id) UNIQUE,  -- 一单一发,天然幂等
    tier          TEXT        NOT NULL CHECK (tier IN ('pro', 'master')),
    starts_at     TIMESTAMPTZ NOT NULL,
    ends_at       TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_entitlements_user ON entitlements(user_id, ends_at DESC);

-- 支付回调事件流水(webhook 幂等:同一事件只处理一次)
CREATE TABLE payment_events (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    channel       TEXT        NOT NULL,
    event_id      TEXT        NOT NULL,             -- 渠道事件唯一 id
    order_no      TEXT,
    payload       JSONB       NOT NULL,
    processed_at  TIMESTAMPTZ,
    error         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel, event_id)
);

-- ── 古籍阅读进度(丝滑体验:跨端续读)────────────────────────

CREATE TABLE reading_progress (
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_slug     TEXT        NOT NULL,
    chapter_idx   INT         NOT NULL DEFAULT 0,
    paragraph_id  TEXT        NOT NULL DEFAULT '',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, book_slug)
);

-- ── 通用触发器:updated_at 自动更新 ──────────────────────────

CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['users', 'chart_profiles', 'conversations', 'orders']
    LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%s_touch BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION touch_updated_at()',
            t, t);
    END LOOP;
END $$;

COMMIT;
