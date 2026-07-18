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

/** 旺衰小字 + 空/破/暗徽记(装卦客观标注层,依增删卜易口径)。 */
function StateMarks({ yao }: { yao: LiuYaoYao }) {
  const flags: { ch: string; cls: string; label: string }[] = [];
  if (yao.xunKong) flags.push({ ch: "空", cls: "text-warn", label: "旬空" });
  if (yao.yuePo) flags.push({ ch: "破", cls: "text-danger", label: "月破" });
  if (yao.riPo) flags.push({ ch: "破", cls: "text-danger", label: "日破" });
  if (yao.anDong) flags.push({ ch: "暗", cls: "text-gold", label: "暗动" });
  if (yao.dayStage === "长生") flags.push({ ch: "长", cls: "text-ok", label: "日辰长生" });
  if (yao.dayStage === "墓") flags.push({ ch: "墓", cls: "text-warn", label: "日辰入墓" });
  if (yao.dayStage === "绝") flags.push({ ch: "绝", cls: "text-danger", label: "日辰临绝" });
  return (
    <span className="ml-1 inline-flex items-center gap-0.5 align-middle">
      {yao.monthState && (
        <span className="text-[10px] leading-none text-ink-faint" title={`对月建${yao.monthState}`}>
          {yao.monthState}
        </span>
      )}
      {flags.map((f, i) => (
        <span key={i} className={`text-[10px] leading-none ${f.cls}`} title={f.label} aria-label={f.label}>
          {f.ch}
        </span>
      ))}
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

/** 动变作用徽记(单字,悬停见全称)。 */
const BIAN_MARK: Record<string, { ch: string; cls: string }> = {
  化进神: { ch: "进", cls: "text-gold" },
  化退神: { ch: "退", cls: "text-warn" },
  伏吟: { ch: "伏", cls: "text-warn" },
  反吟: { ch: "反", cls: "text-danger" },
  化长生: { ch: "长", cls: "text-ok" },
  化墓: { ch: "墓", cls: "text-warn" },
  化绝: { ch: "绝", cls: "text-danger" },
  化合: { ch: "合", cls: "text-ok" },
  回头生: { ch: "生", cls: "text-ok" },
  回头克: { ch: "克", cls: "text-danger" },
};

function BianMark({ relation }: { relation?: string }) {
  if (!relation || !BIAN_MARK[relation]) return null;
  const m = BIAN_MARK[relation];
  return (
    <span className={`ml-0.5 text-[10px] leading-none ${m.cls}`} title={relation} aria-label={relation}>
      {m.ch}
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
    ? "grid-cols-[30px_32px_72px_minmax(44px,1fr)_22px_92px] sm:grid-cols-[34px_36px_84px_minmax(64px,1fr)_24px_108px]"
    : "grid-cols-[30px_32px_72px_minmax(44px,1fr)_22px] sm:grid-cols-[34px_36px_84px_minmax(64px,1fr)_24px]";

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

      {/* 断卦基准:月建 / 日辰 / 用神建议 */}
      <div className="tnum mt-3 flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] tracking-[0.06em] text-ink-faint">
        <span>农历 {result.lunarText}</span>
        <span>· 月建 {result.monthJian}</span>
        <span>· 日辰 {result.dayStem}{result.dayBranch}</span>
        {movingNums.length === 0 && <span>· 六爻安静</span>}
        {result.yongShen && (
          <span title={result.yongShenBasis}>
            · 用神建议 <span className="text-gold">{result.yongShen}</span>
            {(result.yongShenPos ?? []).length > 0
              ? `(第 ${(result.yongShenPos ?? []).join("、")} 爻)`
              : result.yongShen !== "世爻"
                ? "(不上卦)"
                : ""}
          </span>
        )}
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
            <span className="whitespace-nowrap leading-none">
              <GanZhi stem={y.stem} branch={y.branch} element={y.element} />
              <StateMarks yao={y} />
            </span>
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
                    <BianMark relation={y.bianRelation} />
                  </>
                ) : null}
              </span>
            )}
          </div>
        ))}
      </div>

      {/* 图例 */}
      <p className="mt-5 text-[11px] leading-relaxed text-ink-faint">
        自上而下为上爻至初爻;○ 老阳动、× 老阴动,动爻化出右侧变爻。干支旁小字为对月建旺衰
        (旺相休囚死),空=旬空、破=月破/日破、暗=暗动、长/墓/绝=对日辰四态;变爻旁
        进/退/伏/反/长/墓/绝/合/生/克为动变作用(增删卜易口径,悬停见全称)。六神依日辰{result.dayStem}日起。
      </p>
    </div>
  );
}
