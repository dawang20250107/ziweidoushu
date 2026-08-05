"use client";

import type { DaLiuRenResult } from "@/lib/divination";

/**
 * 大六壬式盘:对标紫微十二宫盘面质感的 4×4 环布(南上北下,巳午未申居上)。
 * 每格为一地盘定位:大字天盘之神、天将、地盘支锚点;三传所落之格金环高亮并标
 * 初/中/末,贵人格金字,干支寄坐标「干/支」;中宫呈课骨(课体/日时将贵/旬空/三传)。
 */

const LEVEL_CLS: Record<string, string> = {
  good: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]",
  neutral: "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)]",
  caution: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]",
};
const LEVEL_LABEL: Record<string, string> = { good: "吉", neutral: "平", caution: "慎" };
const CHUAN_NAMES = ["初", "中", "末"];
const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];
// 4×4 环布之地盘支序(-1 为中宫,自动流布:辰/酉、卯/戌各居中宫两侧)
const RING = [5, 6, 7, 8, 4, 9, 3, 10, 2, 1, 0, 11];

export function LiurenPan({ result: r }: { result: DaLiuRenResult }) {
  const ganSeat = r.ke?.[0]?.lower; // 日干寄宫(第一课之下)
  const zhiSeat = r.ke?.[2]?.lower; // 日支(第三课之下)
  const kong = new Set(r.xunKong ?? []);

  // 某地盘位上之天盘神所属三传序(可多传同神,如八专中末相并)
  const chuanAt = (shen: string): number[] =>
    r.chuan.map((c, i) => (c === shen ? i : -1)).filter((i) => i >= 0);

  const cell = (bi: number) => {
    const b = BRANCHES[bi];
    const shen = r.tianPan[bi];
    const jiang = r.tianJiang?.[bi];
    const chuans = chuanAt(shen);
    const isChuan = chuans.length > 0;
    const isGui = jiang === "贵人";
    return (
      <div
        key={b}
        className={[
          "relative flex min-h-[88px] flex-col items-center justify-center gap-1 rounded-[6px] bg-bg px-1 py-2 sm:min-h-[104px]",
          isChuan
            ? "shadow-[inset_0_0_0_1px_var(--gold-dim),0_0_18px_var(--gold-glow)]"
            : "shadow-[inset_0_0_0_1px_var(--line)]",
        ].join(" ")}
      >
        {/* 干/支 坐标 */}
        <span className="absolute left-1.5 top-1.5 flex gap-1">
          {b === ganSeat && (
            <span className="rounded-[2px] px-1 py-0.5 text-[9px] leading-none text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
              干
            </span>
          )}
          {b === zhiSeat && (
            <span className="rounded-[2px] px-1 py-0.5 text-[9px] leading-none text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)]">
              支
            </span>
          )}
        </span>
        {/* 三传标记 */}
        {isChuan && (
          <span className="absolute right-1.5 top-1.5 rounded-[2px] bg-[var(--gold-glow)] px-1 py-0.5 text-[9px] font-medium leading-none text-gold">
            {chuans.map((i) => CHUAN_NAMES[i]).join("·")}
          </span>
        )}

        <span
          className={`font-display text-[23px] leading-none sm:text-[27px] ${isChuan || isGui ? "font-semibold text-gold" : "text-ink"}`}
        >
          {shen}
          {kong.has(shen) && (
            <span className="ml-0.5 align-super text-[9px] font-normal text-danger">空</span>
          )}
        </span>
        {jiang && (
          <span
            className={`text-[10px] leading-none ${isGui ? "font-medium text-gold" : "text-ink-secondary"}`}
          >
            {jiang}
          </span>
        )}
        <span className="absolute bottom-1.5 right-2 text-[11px] leading-none text-ink-faint">{b}</span>
      </div>
    );
  };

  return (
    <div className="rounded-[10px] bg-bg-raised p-2.5 shadow-[0_0_0_1px_var(--line)] sm:p-3">
      <div className="grid grid-cols-4 gap-1.5 sm:gap-2">
        {RING.slice(0, 5).map(cell)}
        {/* 中宫课骨 */}
        <div className="col-span-2 row-span-2 flex flex-col items-center justify-center gap-2 rounded-[6px] bg-bg px-3 py-3 text-center shadow-[inset_0_0_0_1px_var(--line)]">
          <p className="text-[10px] font-medium tracking-[0.3em] text-ink-faint">大六壬 · 月将加时</p>
          <div className="flex items-center gap-2">
            <span className="font-display text-[23px] font-semibold text-ink sm:text-[27px]">{r.keType}课</span>
            {r.judgment && (
              <span
                className={`rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none ${LEVEL_CLS[r.judgment.level] ?? LEVEL_CLS.neutral}`}
              >
                {LEVEL_LABEL[r.judgment.level] ?? "平"}
              </span>
            )}
          </div>
          <p className="tnum text-[12px] leading-relaxed text-ink-secondary">
            {r.dayStem}
            {r.dayBranch}日 · {r.hourBranch}时占
            {r.hourNote && <span className="text-gold">({r.hourNote})</span>}
            <br />
            月将{r.monthGen}
            {r.guiIsDay != null && ` · ${r.guiIsDay ? "昼贵" : "夜贵"}`}
            {r.xunKong?.length === 2 && ` · 旬空${r.xunKong[0]}${r.xunKong[1]}`}
          </p>
          <div className="flex items-center gap-2.5">
            {r.chuan.map((c, i) => (
              <span key={i} className="flex flex-col items-center gap-0.5">
                <span className="font-display text-[16px] leading-none text-gold">
                  {r.chuanDunGan?.[i] && (
                    <span className="mr-px text-[10px] text-ink-faint">{r.chuanDunGan[i]}</span>
                  )}
                  {c}
                </span>
                <span className="text-[9px] leading-none text-ink-faint">{CHUAN_NAMES[i]}传</span>
              </span>
            ))}
          </div>
        </div>
        {RING.slice(5).map(cell)}
      </div>
      <p className="mt-2 px-1 text-[11px] leading-relaxed text-ink-faint">
        每格大字为天盘之神,右下角为地盘定位;金环为三传所落,「干/支」为四课起处;天将随贵人顺逆布十二位。
      </p>
    </div>
  );
}
