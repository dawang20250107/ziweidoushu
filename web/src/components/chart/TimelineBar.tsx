"use client";

import type { Chart } from "@/lib/types";
import { branchName } from "@/lib/chart-helpers";

export interface TimelineSelection {
  /** 目标公历年;null = 关闭运限叠加(纯本命盘) */
  year: number | null;
}

/**
 * 运限时间轴:大限段(十二段)→ 段内流年(十年)。
 * 选中流年即定运限目标日期(取该公历年 7 月 15 日,避开农历年界歧义)。
 */
export function TimelineBar({
  chart, selection, onChange,
}: {
  chart: Chart;
  selection: TimelineSelection;
  onChange: (sel: TimelineSelection) => void;
}) {
  const birthYear = chart.birthInfo.year;
  const selectedYear = selection.year;
  // 选中年份 → 虚岁 → 所在大限
  const nominalAge = selectedYear != null ? selectedYear - birthYear + 1 : null;
  const activeDx = nominalAge != null
    ? chart.daXians.find((d) => nominalAge >= d.startAge && nominalAge <= d.endAge)
    : null;

  return (
    <div className="rounded-[6px] bg-bg-raised p-3 shadow-[0_0_0_1px_var(--line)]">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-[12px] tracking-[0.08em] text-ink-faint">运限时间轴</span>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => onChange({ year: new Date().getFullYear() })}
            className="rounded-[2px] border border-line-strong px-2 py-0.5 text-[12px] text-ink-secondary transition-colors hover:border-gold-dim hover:text-gold"
          >
            今年
          </button>
          {selectedYear != null && (
            <button
              type="button"
              onClick={() => onChange({ year: null })}
              className="rounded-[2px] border border-line-strong px-2 py-0.5 text-[12px] text-ink-secondary transition-colors hover:border-gold-dim hover:text-gold"
            >
              回本命盘
            </button>
          )}
        </div>
      </div>

      {/* 大限段 */}
      <div className="scrollbar-thin flex gap-1 overflow-x-auto pb-1">
        {chart.daXians.map((dx) => {
          const active = activeDx != null && dx.startAge === activeDx.startAge;
          return (
            <button
              key={dx.startAge}
              type="button"
              onClick={() => onChange({ year: birthYear + dx.startAge - 1 })}
              aria-pressed={active}
              className={[
                "flex min-w-[72px] flex-col items-center rounded-[4px] px-2 py-1.5 transition-colors",
                active
                  ? "bg-[var(--gold-glow)] text-gold shadow-[0_0_0_1px_var(--gold-dim)]"
                  : "text-ink-secondary shadow-[0_0_0_1px_var(--line)] hover:text-ink hover:shadow-[0_0_0_1px_var(--line-strong)]",
              ].join(" ")}
            >
              <span className="tnum text-[13px] font-medium">
                {dx.startAge}-{dx.endAge}
              </span>
              <span className="text-[11px]">
                {dx.palaceName}·{branchName(dx.palaceBranch)}
              </span>
            </button>
          );
        })}
      </div>

      {/* 流年(所选大限内十年) */}
      {activeDx && (
        <div className="mt-2 flex flex-wrap gap-1 border-t border-line pt-2">
          {Array.from({ length: 10 }, (_, i) => {
            const age = activeDx.startAge + i;
            const year = birthYear + age - 1;
            const active = selectedYear === year;
            return (
              <button
                key={year}
                type="button"
                onClick={() => onChange({ year })}
                aria-pressed={active}
                className={[
                  "tnum rounded-[2px] px-2 py-0.5 text-[12px] transition-colors",
                  active
                    ? "bg-gold font-medium text-[#161206]"
                    : "text-ink-secondary shadow-[0_0_0_1px_var(--line)] hover:text-ink",
                ].join(" ")}
              >
                {year}
                <span className="ml-1 opacity-70">{age}岁</span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
