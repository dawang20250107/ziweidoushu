-- ============================================================
-- 0004_billing_credits.sql — 双轨变现:订阅 + 按次付费(次卡)
--
--   products.kind = 'subscription' → 履约发放 tier 时长(entitlements)
--   products.kind = 'credits'      → 履约发放次数(credit_ledger + credit_balances)
--   次数消费为原子扣减(balance >= 0 约束兜底),账本全程可审计。
-- ============================================================

BEGIN;

-- ── 商品双轨化 ───────────────────────────────────────────────
ALTER TABLE products
    ADD COLUMN kind          TEXT NOT NULL DEFAULT 'subscription'
        CHECK (kind IN ('subscription', 'credits')),
    ADD COLUMN credit_type   TEXT,
    ADD COLUMN credit_amount INT;

ALTER TABLE products
    ALTER COLUMN tier DROP NOT NULL,
    ALTER COLUMN duration_days DROP NOT NULL;

ALTER TABLE products ADD CONSTRAINT products_kind_shape CHECK (
    (kind = 'subscription' AND tier IS NOT NULL AND duration_days IS NOT NULL)
 OR (kind = 'credits' AND credit_type IS NOT NULL AND credit_amount > 0)
);

-- ── 支付渠道:补充 dev 渠道(本地/E2E 模拟支付;生产由 PAY_DEV_ENABLED 闸门禁用)
ALTER TABLE orders DROP CONSTRAINT orders_channel_check;
ALTER TABLE orders ADD CONSTRAINT orders_channel_check
    CHECK (channel IN ('wechat', 'alipay', 'dev'));

-- ── 次数余额与账本 ───────────────────────────────────────────
CREATE TABLE credit_balances (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credit_type TEXT NOT NULL,
    balance     INT  NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, credit_type)
);

CREATE TABLE credit_ledger (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    credit_type TEXT NOT NULL,
    delta       INT  NOT NULL,             -- 正=发放/退还,负=消费
    reason      TEXT NOT NULL,             -- purchase / consume:xxx / refund:xxx
    order_id    UUID REFERENCES orders(id),
    ref         TEXT,                      -- 消费引用(如报告主题)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- 一单一发,天然幂等
CREATE UNIQUE INDEX idx_credit_ledger_order ON credit_ledger(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_credit_ledger_user ON credit_ledger(user_id, created_at DESC);

-- ── 种子商品(价格为占位,上线前在管理后台调整)────────────
INSERT INTO products (id, title, kind, tier, duration_days, credit_type, credit_amount, price_cents, original_price_cents, sort_order) VALUES
    ('pro_1m',    'Pro 月度会员',   'subscription', 'pro',    31,  NULL,          NULL, 2990,  NULL,  10),
    ('pro_12m',   'Pro 年度会员',   'subscription', 'pro',    366, NULL,          NULL, 19900, 35880, 11),
    ('master_1m', '大师版 月度',    'subscription', 'master', 31,  NULL,          NULL, 9900,  NULL,  20),
    ('report_1',  '深度报告 · 单次', 'credits',      NULL,     NULL, 'deep_report', 1,    1990,  NULL,  30),
    ('report_5',  '深度报告 · 5 次卡', 'credits',    NULL,     NULL, 'deep_report', 5,    7900,  9950,  31)
ON CONFLICT (id) DO NOTHING;

COMMIT;
