"use client";

import type { LiuYaoResult, LiuYaoYao } from "@/lib/divination";
import { ELEMENT_VAR } from "@/lib/divination";

/**
 * 六爻装卦盘面:自上而下(上爻→初爻)六行,列对齐——
 * 六神 | 六亲 | 纳甲干支(五行色)| 爻画(动爻着金,老阳○/老阴×)| 世应 | 动变(→ 变爻六亲干支)。
 * 有动爻才出「动变」列;移动端压缩字号与列距,单行不折。
 */

/** 干支 + 五行(五行按语义色)。 */
function GanZhi({ stem, branch, element, size = 13 }: { stem: string; branch: string; element: string; size?: number }) {
  return (
    <span className="tnum whitespace-nowrap leading-none" style={{ fontSize: size }}>
      <span className="text-ink">{stem}{branch}</span>
      <span className="ml-0.5" style={{ color: ELEMENT_VAR[element] ?? "var(--ink)" }}>{element}</span>
    </span>
  );
}

/** 爻画:阳实阴断;动爻着金并标 ○(老阳)/×(老阴)。 */
function YaoBar({ yao }: { yao: LiuYaoYao }) {
  const fill = yao.moving ? "var(--gold)" : "var(--ink)";
  const bar = {
    height: 8,
    background: fill,
    boxShadow: yao.moving ? "inset 0 0 0 1px var(--gold-dim)" : "inset 0 0 0 1px var(--line-strong)",
    borderRadius: 1,
  };
  return (
    <span className="flex min-w-0 flex-1 items-center gap-1.5">
      <span className="flex flex-1 items-center" style={{ gap: 10 }}>
        {yao.yang ? (
          <span className="w-full" style={bar} />
        ) : (
          <>
            <span className="flex-1" style={bar} />
            <span className="flex-1" style={bar} />
          </>
        )}
      </span>
      {/* 动爻标记:老阳○ 老阴× */}
      <span className="w-3 shrink-0 text-center text-[12px] leading-none text-gold" aria-hidden>
        {yao.moving ? (yao.yang ? "○" : "×") : ""}
      </span>
    </span>
  );
}

/** 世/应徽标。 */
function ShiYing({ yao }: { yao: LiuYaoYao }) {
  if (yao.isShi) {
    return (
      <span className="rounded-[2px] px-1 py-0.5 text-[10px] leading-none tracking-[0.1em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
        世
      </span>
    );
  }
  if (yao.isYing) {
    return (
      <span className="rounded-[2px] px-1 py-0.5 text-[10px] leading-none tracking-[0.1em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
        应
      </span>
    );
  }
  return null;
}

export function LiuYaoPan({ result }: { result: LiuYaoResult }) {
  const hasBian = !!result.bianName;
  const movingNums = result.movingNums ?? []; // 后端保证 [],此处再防御一层
  // 自上而下渲染:上爻(6)在顶,初爻(1)在底。
  const rows = [...(result.yaos ?? [])].sort((a, b) => b.pos - a.pos);
  // 动变列取定宽,保证六行爻画右缘对齐(每行各自成 grid,max-content 会因内容宽度不一错位)
  const cols = hasBian
    ? "grid-cols-[30px_32px_50px_minmax(56px,1fr)_22px_92px] sm:grid-cols-[34px_36px_54px_minmax(72px,1fr)_24px_108px]"
    : "grid-cols-[30px_32px_50px_minmax(56px,1fr)_22px] sm:grid-cols-[34px_36px_54px_minmax(72px,1fr)_24px]";

  return (
    <div className="rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)] sm:px-7 md:px-9">
      {/* 卦头:本卦(宫·世次)→ 变卦 */}
      <div className="flex flex-wrap items-baseline gap-x-4 gap-y-1.5">
        <div className="flex items-baseline gap-2.5">
          <span className="text-[11px] font-medium tracking-[0.24em] text-ink-faint">本卦</span>
          <span className="font-display text-[27px] font-semibold leading-none text-ink sm:text-[31px]">
            {result.benName}
          </span>
          <span className="whitespace-nowrap text-[12px] text-ink-secondary">
            {result.palace} · {result.palaceSeq}
          </span>
        </div>
        {hasBian && (
          <div className="flex items-baseline gap-2.5">
            <span aria-hidden className="text-[15px] text-gold">→</span>
            <span className="text-[11px] font-medium tracking-[0.24em] text-ink-faint">变卦</span>
            <span className="font-display text-[27px] font-semibold leading-none text-ink sm:text-[31px]">
              {result.bianName}
            </span>
          </div>
        )}
      </div>

      {/* 断卦基准:月建 / 日辰 */}
      <div className="tnum mt-3 flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] tracking-[0.06em] text-ink-faint">
        <span>农历 {result.lunarText}</span>
        <span>· 月建 {result.monthJian}</span>
        <span>· 日辰 {result.dayStem}{result.dayBranch}</span>
        {movingNums.length === 0 && <span>· 六爻安静</span>}
      </div>

      {/* 爻列表(自上而下) */}
      <div className="mt-6 flex flex-col">
        {rows.map((y) => (
          <div
            key={y.pos}
            className={`grid ${cols} items-center gap-x-1.5 rounded-[4px] px-1 py-2 sm:gap-x-3 ${
              y.moving ? "bg-[var(--gold-glow)]" : ""
            }`}
            aria-label={`第${y.pos}爻 ${y.liuShen} ${y.liuQin} ${y.stem}${y.branch}${y.element}${
              y.isShi ? " 世" : y.isYing ? " 应" : ""
            }${y.moving ? " 动" : ""}`}
          >
            <span className="text-[11px] leading-none text-ink-faint">{y.liuShen}</span>
            <span className="text-[12px] leading-none text-ink-secondary">{y.liuQin}</span>
            <GanZhi stem={y.stem} branch={y.branch} element={y.element} />
            <YaoBar yao={y} />
            <span className="flex justify-center">
              <ShiYing yao={y} />
            </span>
            {hasBian && (
              <span className="whitespace-nowrap text-right text-[11px] leading-none text-ink-secondary">
                {y.moving && y.bianYao ? (
                  <>
                    <span aria-hidden className="mr-1 text-gold">→</span>
                    {y.bianYao.liuQin}{" "}
                    <GanZhi stem={y.bianYao.stem} branch={y.bianYao.branch} element={y.bianYao.element} size={11} />
                  </>
                ) : null}
              </span>
            )}
          </div>
        ))}
      </div>

      {/* 图例 */}
      <p className="mt-5 text-[11px] leading-relaxed text-ink-faint">
        自上而下为上爻至初爻;○ 老阳动、× 老阴动,动爻化出右侧变爻。六神依日辰{result.dayStem}日起。
      </p>
    </div>
  );
}
