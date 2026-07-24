"use client";

import type { HemingMatchReading } from "@/lib/types";

/**
 * 合盘契合度面板:确定性比对双盘(年命相合/四化互飞/夫妻宫呼应/相处建议)。
 * 契合分与分向断语随双盘星情而定,不走 LLM、始终可用。
 */

const LEVEL: Record<string, string> = {
  good: "text-ok",
  caution: "text-danger",
  neutral: "text-ink-secondary",
};

export function HemingReadingPanel({ reading }: { reading: HemingMatchReading | null | undefined }) {
  if (!reading || reading.sections.length === 0) return null;
  const pct = Math.max(0, Math.min(100, reading.score));

  return (
    <section className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)] md:p-8">
      <div className="flex flex-col items-center gap-4 sm:flex-row sm:items-center sm:gap-6">
        {/* 契合度环 */}
        <div
          className="relative grid h-28 w-28 shrink-0 place-items-center rounded-full"
          style={{
            background: `conic-gradient(var(--gold) ${pct * 3.6}deg, var(--line) 0deg)`,
          }}
          role="img"
          aria-label={`契合度 ${pct} 分`}
        >
          <div className="grid h-[92px] w-[92px] place-items-center rounded-full bg-bg-raised">
            <span className="font-display text-[30px] font-semibold leading-none text-gold">{pct}</span>
            <span className="mt-1 text-[11px] tracking-[0.16em] text-ink-faint">契合度</span>
          </div>
        </div>
        <div className="min-w-0 flex-1 text-center sm:text-left">
          <div className="flex items-baseline justify-center gap-2 sm:justify-start">
            <span className="text-[12px] tracking-[0.24em] text-gold">合盘契合</span>
            <span className="font-display text-[18px] font-semibold text-ink">{reading.level}</span>
          </div>
          <p className="mt-2 text-[14px] leading-[1.9] text-ink-secondary">{reading.summary}</p>
        </div>
      </div>

      <dl className="mt-6 flex flex-col gap-2.5">
        {reading.sections.map((s) => (
          <div key={s.key} className="rounded-[8px] bg-bg px-4 py-3 shadow-[inset_0_0_0_1px_var(--line)]">
            <dt className={`mb-1 font-display text-[14px] ${LEVEL[s.level] ?? "text-ink-secondary"}`}>{s.title}</dt>
            <dd className="text-[13.5px] leading-[1.85] text-ink-secondary">{s.text}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
