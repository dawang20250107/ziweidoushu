-- ============================================================
-- 0006_divination.sql — 占卜线(梅花易数/小六壬)按次付费商品
--   起卦免费引流;AI 深度解卦消耗 divination 次数。

-- ============================================================

BEGIN;

INSERT INTO products (id, title, kind, tier, duration_days, credit_type, credit_amount, price_cents, original_price_cents, sort_order) VALUES
    ('divine_3',  'AI 解卦 · 3 次',  'credits', NULL, NULL, 'divination', 3,  990,  NULL, 40),
    ('divine_10', 'AI 解卦 · 10 次', 'credits', NULL, NULL, 'divination', 10, 2490, 3300, 41)
ON CONFLICT (id) DO NOTHING;

COMMIT;
