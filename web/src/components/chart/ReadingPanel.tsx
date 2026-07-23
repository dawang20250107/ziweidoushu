"use client";

import { useState } from "react";
import type { Reading } from "@/lib/types";

/**
 * 多维断语面板:逐宫(命/财/官/夫妻/迁移/福德/疾厄/田宅/子女/兄弟/交友/父母)+ 当前大限,
 * 断语由命盘「主星 × 庙旺 × 四化 × 煞吉」组合而成,随盘而异。默认展开总论 + 前几维,
 * 点选维度切换。这是确定性的「准头」骨架,与 AI 深度解读互补。
 */

const LEVEL: Record<string, { label: string; cls: string }> = {
  good: { label: "吉", cls: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]" },
  caution: { label: "慎", cls: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]" },
  neutral: { label: "平", cls: "text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]" },
};

export function ReadingPanel({ reading }: { reading: Reading }) {
  const [active, setActive] = useState(0);
  if (!reading || reading.sections.length === 0) return null;
  const cur = reading.sections[active] ?? reading.sections[0];

  return (
    <section className="rounded-[10px] bg-bg-raised shadow-[0_0_0_1px_var(--line)]">
      <div className="flex items-baseline justify-between px-5 pt-4">
        <span className="text-[12px] tracking-[0.24em] text-gold">多维断语</span>
        <span className="text-[11px] text-ink-faint">主星 × 庙旺 × 四化 × 煞吉 · 随盘而异</span>
      </div>

      {/* 命格总论 */}
      <div className="mx-5 mt-3 rounded-[8px] bg-bg px-4 py-3 shadow-[inset_0_0_0_1px_var(--line)]">
        <p className="mb-1 text-[11px] tracking-[0.16em] text-ink-faint">命格总论</p>
        <p className="text-[13.5px] leading-relaxed text-ink-secondary">{reading.overview}</p>
      </div>

      {/* 维度切换 */}
      <div className="mt-4 flex gap-1.5 overflow-x-auto px-5 pb-1">
        {reading.sections.map((s, i) => (
          <button
            key={s.key}
            type="button"
            onClick={() => setActive(i)}
            aria-pressed={i === active}
            className={[
              "flex shrink-0 items-center gap-1 rounded-full px-3 py-1.5 text-[12px] transition-colors",
              i === active
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
      <div className="border-t border-line px-5 pb-6 pt-4">
        <div className="mb-2 flex flex-wrap items-baseline gap-x-3 gap-y-1">
          <span className="font-display text-[16px] font-semibold text-ink">{cur.title}</span>
          {cur.stars.length > 0 && (
            <span className="text-[12px] text-ink-faint">{cur.stars.join("、")}</span>
          )}
        </div>
        <p className="text-[14px] leading-[1.9] text-ink-secondary">{cur.text}</p>
      </div>
    </section>
  );
}
