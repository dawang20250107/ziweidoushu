"use client";

import { useMemo, useState } from "react";
import type { Reading, ReadingSection } from "@/lib/types";
import { RichReading, LEVEL_ACCENT } from "@/components/chart/RichReading";

/**
 * 多维断语面板:命身格局 / 十二宫位 / 大限流年 / 流年择时 四组,
 * 每组内逐维切换。断语由命盘「主星 × 庙旺 × 四化 × 煞吉 × 双星组合 × 应验」
 * 组合而成,随盘而异——确定性「准头」骨架,与 AI 深度解读互补。
 */

const LEVEL: Record<string, { label: string; cls: string }> = {
  good: { label: "吉", cls: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]" },
  caution: { label: "慎", cls: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]" },
  neutral: { label: "平", cls: "text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]" },
};

// 维度分组(按 section.key),组内保持后端顺序。
const CATS: { id: string; label: string; keys: string[] }[] = [
  { id: "ming", label: "命身格局", keys: ["sihua", "ming"] },
  {
    id: "gong",
    label: "十二宫位",
    keys: ["caibo", "guanlu", "fuqi", "qianyi", "fude", "jie", "tianzhai", "zinv", "xiongdi", "jiaoyou", "fumu"],
  },
  { id: "yun", label: "大限流年", keys: ["daxian", "liunian"] },
  { id: "shi", label: "流年择时", keys: ["healthtiming", "fortunetiming", "monthtiming"] },
];

export function ReadingPanel({ reading }: { reading: Reading }) {
  // 分组:每类保留有内容的维度,未归类的落入「其他」。
  const groups = useMemo(() => {
    const byKey = new Map(reading?.sections.map((s) => [s.key, s]) ?? []);
    const used = new Set<string>();
    const out: { id: string; label: string; items: ReadingSection[] }[] = [];
    for (const cat of CATS) {
      const items = cat.keys.map((k) => byKey.get(k)).filter((s): s is ReadingSection => !!s);
      items.forEach((s) => used.add(s.key));
      if (items.length) out.push({ id: cat.id, label: cat.label, items });
    }
    const rest = (reading?.sections ?? []).filter((s) => !used.has(s.key));
    if (rest.length) out.push({ id: "misc", label: "其他", items: rest });
    return out;
  }, [reading]);

  const [catId, setCatId] = useState(groups[0]?.id ?? "");
  const [key, setKey] = useState(groups[0]?.items[0]?.key ?? "");

  if (!reading || groups.length === 0) return null;

  const group = groups.find((g) => g.id === catId) ?? groups[0];
  const cur = group.items.find((s) => s.key === key) ?? group.items[0];

  function pickCat(id: string) {
    setCatId(id);
    const g = groups.find((x) => x.id === id);
    if (g) setKey(g.items[0].key);
  }

  return (
    <section className="rounded-[10px] bg-bg-raised shadow-[0_0_0_1px_var(--line)]">
      <div className="flex items-baseline justify-between px-5 pt-4">
        <span className="text-[12px] tracking-[0.24em] text-gold">多维断语</span>
        <span className="text-[11px] text-ink-faint">主星 × 庙旺 × 四化 × 煞吉 × 应验 · 随盘而异</span>
      </div>

      {/* 命格总论 */}
      <div className="mx-5 mt-3 rounded-[8px] bg-bg px-4 py-3 shadow-[inset_0_0_0_1px_var(--line)]">
        <p className="mb-1.5 text-[11px] tracking-[0.16em] text-ink-faint">命格总论</p>
        <RichReading text={reading.overview} />
      </div>

      {/* 一级:分组 */}
      <div className="mt-4 flex gap-1 overflow-x-auto px-5">
        {groups.map((g) => (
          <button
            key={g.id}
            type="button"
            onClick={() => pickCat(g.id)}
            aria-pressed={g.id === group.id}
            className={[
              "shrink-0 border-b-2 px-3 pb-2 text-[13px] transition-colors",
              g.id === group.id
                ? "border-gold font-medium text-gold"
                : "border-transparent text-ink-faint hover:text-ink-secondary",
            ].join(" ")}
          >
            {g.label}
          </button>
        ))}
      </div>

      {/* 二级:组内维度 */}
      <div className="flex gap-1.5 overflow-x-auto border-t border-line px-5 pt-3 pb-1">
        {group.items.map((s) => (
          <button
            key={s.key}
            type="button"
            onClick={() => setKey(s.key)}
            aria-pressed={s.key === cur.key}
            className={[
              "flex shrink-0 items-center gap-1 rounded-full px-3 py-1.5 text-[12px] transition-colors",
              s.key === cur.key
                ? "bg-[var(--gold-glow)] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                : "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] hover:text-ink",
            ].join(" ")}
          >
            {s.title}
            <span className={`rounded-[3px] px-1 text-[10px] leading-none ${LEVEL[s.level]?.cls ?? ""}`}>
              {LEVEL[s.level]?.label ?? "平"}
            </span>
          </button>
        ))}
      </div>

      {/* 当前维度详情 */}
      <div className="px-5 pb-6 pt-4">
        <div
          key={cur.key}
          className="palace-enter rounded-[8px] bg-bg px-4 py-3.5 shadow-[inset_0_0_0_1px_var(--line)]"
          style={{ borderLeft: `2px solid ${LEVEL_ACCENT[cur.level] ?? "var(--line-strong)"}` }}
        >
          <div className="mb-2.5 flex flex-wrap items-center gap-x-3 gap-y-1">
            <span className="font-display text-[16px] font-semibold text-ink">{cur.title}</span>
            <span
              className={`rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none ${LEVEL[cur.level]?.cls ?? ""}`}
            >
              {LEVEL[cur.level]?.label ?? "平"}
            </span>
            {cur.stars.length > 0 && <span className="text-[12px] text-ink-faint">{cur.stars.join("、")}</span>}
          </div>
          <RichReading text={cur.text} />
        </div>
      </div>
    </section>
  );
}
