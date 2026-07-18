"use client";

import type { Hexagram } from "@/lib/divination";
import { trigramLine } from "@/lib/divination";

/**
 * 单卦卦画:六爻自下而上,阳爻实线、阴爻断线;动爻着星曜金并注「动」。
 * 爻线用 --ink(阳/阴笔画)与 --line-strong(描边);emphasis 时放大(本卦「大」)。
 */

interface Size {
  bar: number; // 爻线粗细
  gap: number; // 爻行间距
  slit: number; // 阴爻中缝
  gutter: string; // 左右留白(容纳「动」标)
  width: string; // 卦画总宽
  name: string; // 卦名字号类
}

const SIZE_LG: Size = { bar: 12, gap: 9, slit: 16, gutter: "w-8", width: "w-[204px]", name: "text-[31px]" };
const SIZE_MD: Size = { bar: 9, gap: 7, slit: 12, gutter: "w-6", width: "w-[152px]", name: "text-[25px]" };

/** 一根爻:yang=实线,否则断线;moving 时整行着金并标「动」。 */
function YaoLine({ yang, moving, s }: { yang: boolean; moving: boolean; s: Size }) {
  const fill = moving ? "var(--gold)" : "var(--ink)";
  const ring = moving ? "inset 0 0 0 1px var(--gold-dim)" : "inset 0 0 0 1px var(--line-strong)";
  const bar = { height: s.bar, background: fill, boxShadow: ring, borderRadius: 1 };
  return (
    <div
      className={`flex items-center rounded-[2px] ${moving ? "bg-[var(--gold-glow)]" : ""}`}
      style={{ paddingTop: s.gap / 2, paddingBottom: s.gap / 2 }}
    >
      {/* 左侧留白(与右侧「动」标对称,保卦画居中) */}
      <span className={`${s.gutter} shrink-0`} aria-hidden />
      {/* 爻线本体 */}
      <span className="flex flex-1 items-center justify-center" style={{ gap: s.slit }}>
        {yang ? (
          <span className="w-full" style={bar} />
        ) : (
          <>
            <span className="flex-1" style={bar} />
            <span className="flex-1" style={bar} />
          </>
        )}
      </span>
      {/* 右侧「动」标 */}
      <span className={`${s.gutter} shrink-0`}>
        {moving && (
          <span className="rounded-[2px] px-1 py-0.5 text-[10px] leading-none tracking-[0.1em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
            动
          </span>
        )}
      </span>
    </div>
  );
}

export function HexagramView({
  hexagram,
  moving,
  label,
  emphasis = false,
}: {
  hexagram: Hexagram;
  moving: number; // 动爻 1-6
  label: string; // 本卦 / 互卦 / 变卦
  emphasis?: boolean;
}) {
  const s = emphasis ? SIZE_LG : SIZE_MD;
  // lines 自下而上(index 0 = 初爻);卦画自上而下渲染,故倒序。
  const rows = hexagram.lines.map((yang, i) => ({ yang, idx: i })).reverse();

  return (
    <figure className="flex flex-col items-center gap-3">
      <figcaption className="text-[11px] font-medium tracking-[0.24em] text-ink-faint">{label}</figcaption>
      <div
        className={`${s.width} rounded-[8px] bg-bg px-1 py-3 shadow-[inset_0_0_0_1px_var(--line)]`}
        role="img"
        aria-label={`${label}:${hexagram.name}`}
      >
        {rows.map((r) => (
          <YaoLine key={r.idx} yang={r.yang} moving={r.idx + 1 === moving} s={s} />
        ))}
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className={`font-display ${s.name} font-semibold leading-none text-ink`}>{hexagram.name}</span>
        <span className="text-[12px] text-ink-secondary">{trigramLine(hexagram)}</span>
      </div>
    </figure>
  );
}
