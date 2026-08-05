"use client";

import type { Pattern } from "@/lib/types";
import { PatternCard } from "./DetailPanel";

/**
 * 格局总览(全宽卡片墙):从盘面侧栏下沉为独立分区——
 * 长文在多列宽卡中自然铺开,不再把盘面旁边撑成高低不齐的窄柱。
 * 中宫格局徽章锚点跳转至此(id="patterns-overview");
 * 悬停/聚焦卡片时点亮盘上关联宫位(onPatternHover 联动)。
 */
export function PatternsOverview({
  patterns, onPatternHover,
}: {
  patterns: Pattern[];
  onPatternHover?: (names: string[] | null) => void;
}) {
  if (patterns.length === 0) return null;
  return (
    <section id="patterns-overview" className="scroll-mt-24">
      <div className="mb-3 flex items-baseline justify-between">
        <span className="text-[12px] tracking-[0.24em] text-gold">格局总览</span>
        <span className="tnum text-[11px] text-ink-faint">{patterns.length} 个格局</span>
      </div>
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {patterns.map((p) => (
          <div
            key={p.name}
            className="reveal"
            tabIndex={-1}
            onMouseEnter={() => onPatternHover?.(p.palaces ?? [])}
            onMouseLeave={() => onPatternHover?.(null)}
          >
            <PatternCard p={p} />
          </div>
        ))}
      </div>
    </section>
  );
}
