"use client";

import type { Chart, Horoscope } from "@/lib/types";
import { branchName } from "@/lib/chart-helpers";

export interface TimelineSelection {
  /** 目标公历年;null = 关闭运限叠加(纯本命盘) */
  year: number | null;
  /** 公历月 1-12;选中后叠加流月层 */
  month?: number | null;
  /** 公历日;选中后叠加流日层 */
  day?: number | null;
  /** 时辰索引 0-11(子…亥);选中后叠加流时层 */
  hour?: number | null;
}

const rowBtn = (active: boolean) =>
  [
    "tnum rounded-[2px] px-2 py-0.5 text-[12px] transition-colors",
    active
      ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[0_0_0_1px_var(--gold-dim)]"
      : "text-ink-secondary shadow-[0_0_0_1px_var(--line)] hover:text-ink",
  ].join(" ");

/**
 * 运限时间轴:大限(十二段)→ 流年(十年)→ 流月 → 流日 → 流时,逐层下钻。
 * 流月/流日/流时按公历定位目标日期,盘面按农历运限计算(倪师口径:
 * 生年四化与流年四化为准,更深层四化为飞星派研究字段)。
 */
export function TimelineBar({
  chart, selection, horoscope, onChange,
}: {
  chart: Chart;
  selection: TimelineSelection;
  horoscope?: Horoscope | null;
  onChange: (sel: TimelineSelection) => void;
}) {
  const birthYear = chart.birthInfo.year;
  const { year: selectedYear, month, day, hour } = selection;
  // 选中年份 → 虚岁 → 所在大限
  const nominalAge = selectedYear != null ? selectedYear - birthYear + 1 : null;
  const activeDx = nominalAge != null
    ? chart.daXians.find((d) => nominalAge >= d.startAge && nominalAge <= d.endAge)
    : null;
  const daysInMonth = selectedYear != null && month != null
    ? new Date(selectedYear, month, 0).getDate()
    : 0;

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
          <span className="w-8 shrink-0 pt-0.5 text-[11px] text-ink-faint">流年</span>
          {Array.from({ length: 10 }, (_, i) => {
            const age = activeDx.startAge + i;
            const year = birthYear + age - 1;
            return (
              <button
                key={year}
                type="button"
                onClick={() => onChange({ year })}
                aria-pressed={selectedYear === year}
                className={rowBtn(selectedYear === year)}
              >
                {year}
                <span className="ml-1 opacity-70">{age}岁</span>
              </button>
            );
          })}
        </div>
      )}

      {/* 流月(公历定位) */}
      {selectedYear != null && (
        <div className="mt-2 flex flex-wrap gap-1 border-t border-line pt-2">
          <span className="w-8 shrink-0 pt-0.5 text-[11px] text-ink-faint">流月</span>
          {Array.from({ length: 12 }, (_, i) => {
            const m = i + 1;
            return (
              <button
                key={m}
                type="button"
                onClick={() => onChange(month === m ? { year: selectedYear } : { year: selectedYear, month: m })}
                aria-pressed={month === m}
                className={rowBtn(month === m)}
              >
                {m}月
              </button>
            );
          })}
        </div>
      )}

      {/* 流日 */}
      {selectedYear != null && month != null && (
        <div className="mt-2 flex flex-wrap gap-1 border-t border-line pt-2">
          <span className="w-8 shrink-0 pt-0.5 text-[11px] text-ink-faint">流日</span>
          {Array.from({ length: daysInMonth }, (_, i) => {
            const d = i + 1;
            return (
              <button
                key={d}
                type="button"
                onClick={() =>
                  onChange(day === d
                    ? { year: selectedYear, month }
                    : { year: selectedYear, month, day: d })}
                aria-pressed={day === d}
                className={rowBtn(day === d)}
              >
                {d}
              </button>
            );
          })}
        </div>
      )}

      {/* 流时(十二时辰) */}
      {selectedYear != null && month != null && day != null && (
        <div className="mt-2 flex flex-wrap gap-1 border-t border-line pt-2">
          <span className="w-8 shrink-0 pt-0.5 text-[11px] text-ink-faint">流时</span>
          {Array.from({ length: 12 }, (_, i) => (
            <button
              key={i}
              type="button"
              onClick={() =>
                onChange(hour === i
                  ? { year: selectedYear, month, day }
                  : { year: selectedYear, month, day, hour: i })}
              aria-pressed={hour === i}
              className={rowBtn(hour === i)}
            >
              {branchName(i)}时
            </button>
          ))}
        </div>
      )}

      {/* 激活层摘要:各层干支 + 农历定位 */}
      {horoscope && selectedYear != null && (
        <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 border-t border-line pt-2 text-[12px]">
          {([
            ["大限", horoscope.decadal, true],
            ["流年", horoscope.yearly, true],
            ["流月", horoscope.monthly, month != null],
            ["流日", horoscope.daily, day != null],
            ["流时", horoscope.hourly, hour != null],
          ] as const).map(([label, scope, show]) =>
            show ? (
              <span key={label} className="text-ink-secondary">
                <span className="text-ink-faint">{label}</span>
                <span className="ml-1 font-display text-gold">{scope.stem}{scope.branch}</span>
              </span>
            ) : null,
          )}
          <span className="text-ink-faint">农历 {horoscope.targetLunarText}</span>
        </div>
      )}
    </div>
  );
}
