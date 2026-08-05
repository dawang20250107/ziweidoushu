"use client";

import type { Chart, Horoscope, Pattern } from "@/lib/types";
import { branchName } from "@/lib/chart-helpers";

const CHIP_LEVEL: Record<string, string> = {
  excellent: "text-gold-bright shadow-[inset_0_0_0_1px_var(--gold-dim)]",
  good: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]",
  neutral: "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)]",
  caution: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]",
};

/** 中宫:命主信息与四柱 + 格局徽章锚点(悬停点亮关联宫位);运限激活时显示目标日期与虚岁。 */
export function ChartCenter({
  chart, horoscope, patterns, onPatternHover,
}: {
  chart: Chart;
  horoscope?: Horoscope;
  patterns?: Pattern[];
  onPatternHover?: (names: string[] | null) => void;
}) {
  const b = chart.birthInfo;
  const gender = b.gender === "male" ? "男" : "女";
  const pillars = chart.fourPillars;

  return (
    <div className="flex w-full flex-col items-center justify-center rounded-[6px] bg-bg px-4 py-5 text-center shadow-[inset_0_0_0_1px_var(--line)]">
      <div className="font-display text-xl font-semibold">
        {b.name || "命主"}
        <span className="ml-2 text-[13px] font-normal text-ink-secondary">{gender}命</span>
      </div>
      <div className="tnum mt-1 text-[13px] text-ink-secondary">
        {b.year}-{b.month}-{b.day} {chart.timeName}
      </div>
      <div className="mt-0.5 text-[13px] text-ink-faint">{chart.lunarDateText}</div>

      <div className="tnum mt-3 flex gap-2 font-display text-[15px] tracking-wide">
        {[pillars.year, pillars.month, pillars.day, pillars.hour].map((p, i) => (
          <span key={i} className="rounded-[2px] bg-bg-raised px-1.5 py-0.5 shadow-[0_0_0_1px_var(--line)]">
            {p}
          </span>
        ))}
      </div>

      <dl className="mt-3 grid grid-cols-2 gap-x-6 gap-y-0.5 text-[13px]">
        <div className="flex gap-1.5">
          <dt className="text-ink-faint">五行局</dt>
          <dd className="text-ink">{chart.wuxingJuName}</dd>
        </div>
        <div className="flex gap-1.5">
          <dt className="text-ink-faint">生肖</dt>
          <dd className="text-ink">{chart.zodiac}</dd>
        </div>
        <div className="flex gap-1.5">
          <dt className="text-ink-faint">命主</dt>
          <dd className="text-ink">{chart.mingZhu}</dd>
        </div>
        <div className="flex gap-1.5">
          <dt className="text-ink-faint">身主</dt>
          <dd className="text-ink">{chart.shenZhu}</dd>
        </div>
        <div className="flex gap-1.5">
          <dt className="text-ink-faint">命宫</dt>
          <dd className="text-ink">{branchName(chart.mingGongBranch)}宫</dd>
        </div>
        <div className="flex gap-1.5">
          <dt className="text-ink-faint">身宫</dt>
          <dd className="text-ink">{branchName(chart.shenGongBranch)}宫</dd>
        </div>
      </dl>

      {patterns && patterns.length > 0 && (
        <div className="mt-3 flex flex-wrap items-center justify-center gap-1.5">
          {patterns.slice(0, 4).map((p) => (
            <a
              key={p.name}
              href="#patterns-overview"
              onMouseEnter={() => onPatternHover?.(p.palaces ?? [])}
              onMouseLeave={() => onPatternHover?.(null)}
              onFocus={() => onPatternHover?.(p.palaces ?? [])}
              onBlur={() => onPatternHover?.(null)}
              className={`rounded-full px-2 py-0.5 text-[11px] transition-opacity hover:opacity-80 ${CHIP_LEVEL[p.level] ?? CHIP_LEVEL.neutral}`}
            >
              {p.name}
            </a>
          ))}
          {patterns.length > 4 && (
            <a href="#patterns-overview" className="tnum text-[11px] text-ink-faint hover:text-ink-secondary">
              +{patterns.length - 4}
            </a>
          )}
        </div>
      )}

      {horoscope && (
        <div className="mt-3 w-full border-t border-line pt-2 text-[12px]">
          <span className="text-gold">运限</span>
          <span className="tnum ml-2 text-ink-secondary">{horoscope.targetSolarDate}</span>
          <span className="ml-2 text-ink-faint">
            虚岁 <span className="tnum">{horoscope.nominalAge}</span> · {horoscope.decadal.name}行
            {branchName(horoscope.decadal.palaceBranch)}宫
          </span>
        </div>
      )}
    </div>
  );
}
