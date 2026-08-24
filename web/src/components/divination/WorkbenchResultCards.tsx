"use client";

import {
  RELATION_TONE,
  type MeihuaResult,
  type MeihuaJudgment,
  type MeihuaRoleLore,
  type LiuYaoJudgment,
} from "@/lib/divination";
import { toneBadgeClass } from "./tone";
import { JudgeSections } from "./JudgeSections";

// 占卜工作台 · 结果侧卡片(自 Workbench.tsx 拆出:体用/断语/经文/类象)。

/** 体用生克卡。 */
export function TiYongCard({ result }: { result: MeihuaResult }) {
  const rel = RELATION_TONE[result.relation];
  return (
    <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <p className="text-[12px] font-medium tracking-[0.24em] text-gold">体用生克</p>
      <div className="mt-4 grid grid-cols-2 gap-3">
        <TiYongCell
          role="体"
          position={result.tiIsUpper ? "上卦" : "下卦"}
          trigram={result.tiTrigram.name}
          element={result.tiTrigram.element}
        />
        <TiYongCell
          role="用"
          position={result.tiIsUpper ? "下卦" : "上卦"}
          trigram={result.yongTrigram.name}
          element={result.yongTrigram.element}
        />
      </div>
      <div className="mt-5 flex flex-wrap items-center gap-3">
        <span
          className={`inline-flex items-center rounded-[2px] px-2.5 py-1 text-[13px] font-medium tracking-[0.06em] ${toneBadgeClass(
            rel.tone,
          )}`}
        >
          {rel.label}
        </span>
        <p className="flex-1 text-[14px] leading-relaxed text-ink-secondary">{result.verdict}</p>
      </div>
    </div>
  );
}

function TiYongCell({
  role,
  position,
  trigram,
  element,
}: {
  role: string;
  position: string;
  trigram: string;
  element: string;
}) {
  return (
    <div className="flex flex-col gap-1 rounded-[6px] bg-bg px-4 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
      <span className="text-[11px] tracking-[0.1em] text-ink-faint">
        {role} · {position}
      </span>
      <div className="flex items-baseline gap-2">
        <span className="font-display text-[25px] font-semibold text-ink">{trigram}</span>
        <span className="text-[13px] text-ink-secondary">五行 · {element}</span>
      </div>
    </div>
  );
}

const JUDGE_LEVEL: Record<string, { label: string; cls: string }> = {
  good: { label: "吉", cls: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]" },
  neutral: { label: "平", cls: "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)]" },
  caution: { label: "慎", cls: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]" },
};

/** 断卦骨架卡:《梅花易数·体用总诀》确定性推演(卦气/体党/互变/事类/应期)。 */
export function JudgeCard({ j }: { j: MeihuaJudgment }) {
  const lv = JUDGE_LEVEL[j.level] ?? JUDGE_LEVEL.neutral;
  return (
    <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">断卦 · 体用总诀</p>
        <span className="flex items-center gap-2">
          <span className="text-[11px] text-ink-faint">体气【{j.tiQi}】· 所问【{j.topic}】</span>
          <span className={`rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none ${lv.cls}`}>{lv.label}</span>
        </span>
      </div>
      <p className="mt-4 font-reading text-[16px] leading-[1.9] text-ink">{j.conclusion}</p>
      <ul className="mt-4 flex flex-col gap-2.5 border-t border-line pt-4">
        {j.points.map((pt, i) => (
          <li key={i} className="flex gap-2 text-[14px] leading-relaxed text-ink-secondary">
            <span aria-hidden className="mt-[9px] h-[3px] w-[3px] shrink-0 rounded-full bg-gold-dim" />
            {pt}
          </li>
        ))}
      </ul>
      <p className="mt-4 rounded-[6px] bg-bg px-3.5 py-2.5 text-[13px] leading-relaxed text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
        {j.yingQi}
      </p>
      {/* 分节深断:卦象总论/体用之辨/过程与结局/类象取应/应期(免费确定性层) */}
      <JudgeSections sections={j.sections} />
    </div>
  );
}

/** 周易经文卡:卦辞与动爻爻辞(维基文库公版通行本,动爻辞加粗为断卦要义)。 */
export function JingWenCard({ rows }: { rows: { label: string; text: string; strong?: boolean }[] }) {
  return (
    <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <div className="flex items-baseline justify-between">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">周易经文</p>
        <span className="text-[11px] text-ink-faint">通行本原文 · 公版</span>
      </div>
      <dl className="mt-4 flex flex-col gap-2.5">
        {rows.map((row, i) => (
          <div key={i} className="flex flex-col gap-0.5 sm:flex-row sm:gap-3">
            <dt className="shrink-0 text-[12px] leading-[1.9] tracking-[0.06em] text-ink-faint sm:w-24">
              {row.label}
            </dt>
            <dd
              className={`font-reading text-[15px] leading-[1.9] ${row.strong ? "font-medium text-ink" : "text-ink-secondary"}`}
            >
              {row.text}
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

/** 六爻断卦骨架卡:用神旺衰/伏神/卦性/动变/世应/应期确定性推演(免费层)。 */
export function LiuYaoJudgeCard({ j, xingZhi }: { j: LiuYaoJudgment; xingZhi?: string }) {
  const lv = JUDGE_LEVEL[j.level] ?? JUDGE_LEVEL.neutral;
  // 应期已单独落底部框,列表内滤重
  const points = j.points.filter((pt) => !pt.startsWith("应期:"));
  return (
    <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">断卦 · 用神旺衰</p>
        <span className="flex items-center gap-2">
          {xingZhi && <span className="text-[11px] text-ink-faint">本卦【{xingZhi}】</span>}
          <span className={`rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none ${lv.cls}`}>{lv.label}</span>
        </span>
      </div>
      <p className="mt-4 font-reading text-[16px] leading-[1.9] text-ink">{j.conclusion}</p>
      {points.length > 0 && (
        <ul className="mt-4 flex flex-col gap-2.5 border-t border-line pt-4">
          {points.map((pt, i) => (
            <li key={i} className="flex gap-2 text-[14px] leading-relaxed text-ink-secondary">
              <span aria-hidden className="mt-[9px] h-[3px] w-[3px] shrink-0 rounded-full bg-gold-dim" />
              {pt}
            </li>
          ))}
        </ul>
      )}
      {j.yingQi && (
        <p className="mt-4 rounded-[6px] bg-bg px-3.5 py-2.5 text-[13px] leading-relaxed text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
          应期:{j.yingQi}
        </p>
      )}
      {/* 分节深断:取用/旺衰/元忌/动变/世应/逐爻/应期(免费确定性层) */}
      <JudgeSections sections={j.sections} />
    </div>
  );
}

const LORE_ROWS: { key: keyof Omit<MeihuaRoleLore, "role" | "name">; label: string }[] = [
  { key: "renlun", label: "人物" },
  { key: "shenti", label: "身体" },
  { key: "jingwu", label: "器物" },
  { key: "fangwei", label: "方位" },
  { key: "xing", label: "性情" },
];

/** 万物类象卡:体/用/变三角色卦的邵子类占速查(断辞落到具体人事物)。 */
export function LoreCard({ lore }: { lore: MeihuaRoleLore[] }) {
  return (
    <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <div className="flex items-baseline justify-between">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">万物类象</p>
        <span className="text-[11px] text-ink-faint">邵子八卦类占义 · 取象参考</span>
      </div>
      {/* Tailwind 类须静态可析,按数量映射 */}
      <div
        className={`mt-4 grid grid-cols-1 gap-3 ${
          lore.length >= 3 ? "sm:grid-cols-3" : lore.length === 2 ? "sm:grid-cols-2" : ""
        }`}
      >
        {lore.map((l) => (
          <div key={l.role + l.name} className="rounded-[6px] bg-bg px-4 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
            <div className="flex items-baseline gap-2">
              <span className="rounded-[2px] px-1.5 py-0.5 text-[10px] tracking-[0.08em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
                {l.role}
              </span>
              <span className="font-display text-[19px] font-semibold text-ink">{l.name}</span>
            </div>
            <dl className="mt-3 flex flex-col gap-1.5">
              {LORE_ROWS.map((row) => (
                <div key={row.key} className="flex gap-2 text-[12.5px] leading-relaxed">
                  <dt className="shrink-0 text-ink-faint">{row.label}</dt>
                  <dd className="text-ink-secondary">{l[row.key]}</dd>
                </div>
              ))}
            </dl>
          </div>
        ))}
      </div>
    </div>
  );
}
