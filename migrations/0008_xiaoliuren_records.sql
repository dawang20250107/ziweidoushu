-- ============================================================
-- 0008_xiaoliuren_records.sql — 卦档 kind 扩展:纳入小六壬
-- ============================================================

BEGIN;

ALTER TABLE divination_records DROP CONSTRAINT IF EXISTS divination_records_kind_check;
ALTER TABLE divination_records
    ADD CONSTRAINT divination_records_kind_check
    CHECK (kind IN ('meihua', 'liuyao', 'xiaoliuren'));

COMMIT;
