"use client";

import { useState } from "react";
import type { HoroscopeReading } from "@/lib/types";
import { RichReading, LEVEL_ACCENT } from "@/components/chart/RichReading";

/**
 * 运限逐层断语面板:大限 → 流年 → 流月 → 流日 → 流时。
 * 随时间轴选择的目标日期而生成(确定性)。每层「命宫落本命何宫 + 流曜 + 流年四化」,
 * 与本命多维断语互补:本命看一生格局,本层看某一时点运势。
 */

const LEVEL: Record<string, { label: string; cls: string }> = {
  good: { label: "吉", cls: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]" },
  caution: { label: "慎", cls: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]" },
  neutral: { label: "平", cls: "text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]" },
};

export function HoroscopeReadingPanel({ reading }: { reading: HoroscopeReading | null }) {
  const [active, setActive] = useState(0);
  if (!reading || reading.sections.length === 0) return null;
  const idx = Math.min(active, reading.sections.length - 1);
  const cur = reading.sections[idx];

  return (
    <section className="rounded-[10px] bg-bg-raised shadow-[0_0_0_1px_var(--line)]">
      <div className="flex items-baseline justify-between px-5 pt-4">
        <span className="text-[12px] tracking-[0.24em] text-gold">运限断语</span>
        <span className="text-[11px] text-ink-faint">大限 → 流年 → 流月 → 流日 → 流时</span>
      </div>

      <div className="mt-3 flex gap-1.5 overflow-x-auto px-5 pb-1">
        {reading.sections.map((s, i) => (
          <button
            key={s.key}
            type="button"
            onClick={() => setActive(i)}
            aria-pressed={i === idx}
            className={[
              "flex shrink-0 items-center gap-1 rounded-full px-3 py-1.5 text-[12px] transition-colors",
              i === idx
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

      <div className="border-t border-line px-5 pb-6 pt-4">
        <div
          key={cur.key}
          className="palace-enter rounded-[8px] bg-bg px-4 py-3.5 shadow-[inset_0_0_0_1px_var(--line)]"
          style={{ borderLeft: `2px solid ${LEVEL_ACCENT[cur.level] ?? "var(--line-strong)"}` }}
        >
          <div className="mb-2.5 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <span className="font-display text-[16px] font-semibold text-ink">{cur.title}</span>
            {cur.stars.length > 0 && <span className="text-[12px] text-ink-faint">{cur.stars.join("、")}</span>}
          </div>
          <RichReading text={cur.text} />
        </div>
      </div>
    </section>
  );
}
