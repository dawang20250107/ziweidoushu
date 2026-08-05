"use client";

import Link from "next/link";
import type { Chart, Pattern, Palace } from "@/lib/types";
import { branchName, brightnessVar, groupStars, sihuaBadgeStyle, stemName } from "@/lib/chart-helpers";

export const LEVEL_LABEL: Record<string, { text: string; cls: string }> = {
  excellent: { text: "上格", cls: "text-gold-bright shadow-[inset_0_0_0_1px_var(--gold-dim)]" },
  good: { text: "吉格", cls: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]" },
  neutral: { text: "中性", cls: "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)]" },
  caution: { text: "凶格", cls: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]" },
};

export function PatternCard({ p }: { p: Pattern }) {
  const level = LEVEL_LABEL[p.level] ?? LEVEL_LABEL.neutral;
  return (
    <div className="rounded-[6px] bg-bg-raised p-3 shadow-[0_0_0_1px_var(--line)]">
      <div className="mb-1 flex items-center gap-2">
        <span className="font-display text-[15px] font-semibold">{p.name}</span>
        <span className={`rounded-[2px] px-1.5 text-[10px] tracking-[0.08em] ${level.cls}`}>{level.text}</span>
      </div>
      <p className="text-[13px] leading-relaxed text-ink-secondary">{p.description}</p>
      {p.conditions?.breaking && p.conditions.breaking.length > 0 && (
        <p className="mt-1 text-[12px] text-danger">破格:{p.conditions.breaking.join(";")}</p>
      )}
      {p.source && <p className="mt-1.5 font-reading text-[12px] text-ink-faint">{p.source}</p>}
    </div>
  );
}

/** 单宫详情内容(无外框定位,供侧栏/抽屉复用):星曜细目 + 关联格局 + 问 AI。 */
export function PalaceDetail({ palace, patterns }: { palace: Palace; patterns: Pattern[] }) {
  const { major, assist, adjective } = groupStars(palace.stars);
  // 年支系补充杂曜(大耗/龙德/劫煞)并入杂曜组展示
  const allAdjective = [...adjective, ...(palace.extraStars ?? [])];
  const related = patterns.filter((p) => (p.palaces ?? []).includes(palace.name));
  const aiQuestion = `请重点分析我命盘的【${palace.name}】(${stemName(palace.stem)}${branchName(palace.branch)}宫)。`;

  return (
    <div className="flex flex-col gap-3">
      <div className="rounded-[6px] bg-bg-raised p-4 shadow-[0_0_0_1px_var(--line)]">
        <div className="mb-2 flex items-baseline justify-between">
          <span className="font-display text-lg font-semibold">{palace.name}</span>
          <span className="text-[13px] text-ink-secondary">
            {stemName(palace.stem)}{branchName(palace.branch)}
            <span className="tnum ml-2 text-ink-faint">大限 {palace.daXianStart}-{palace.daXianEnd}</span>
          </span>
        </div>

        {palace.isEmpty && palace.borrowedStars && (
          <p className="mb-2 rounded-[4px] bg-bg px-2 py-1.5 text-[12px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
            空宫,借对宫【{palace.borrowedFromName}】主星:{palace.borrowedStars.join("、")}
          </p>
        )}

        {[
          { label: "主星", stars: major },
          { label: "辅曜", stars: assist },
          { label: "杂曜", stars: allAdjective },
        ].map(
          (group) =>
            group.stars.length > 0 && (
              <div key={group.label} className="mb-2">
                <p className="mb-1 text-[11px] tracking-[0.08em] text-ink-faint">{group.label}</p>
                <div className="flex flex-wrap gap-x-3 gap-y-1">
                  {group.stars.map((s) => (
                    <span key={s.name} className="text-[14px]" style={{ color: brightnessVar(s.brightness) }}>
                      {s.name}
                      {s.brightness && <span className="ml-0.5 text-[11px] opacity-80">({s.brightness})</span>}
                      {s.siHua && (
                        <sup
                          className="ml-0.5 rounded-[2px] px-[3px] text-[10px] font-semibold not-italic"
                          style={sihuaBadgeStyle(s.siHua)}
                        >
                          {s.siHua}
                        </sup>
                      )}
                    </span>
                  ))}
                </div>
              </div>
            ),
        )}

        <div className="mt-2 flex gap-3 border-t border-line pt-2 text-[12px] text-ink-faint">
          {palace.changsheng12 && <span>长生:{palace.changsheng12}</span>}
          {palace.boshi12 && <span>博士:{palace.boshi12}</span>}
        </div>

        <Link
          href={`/chat?q=${encodeURIComponent(aiQuestion)}`}
          className="mt-3 block rounded-[6px] bg-gold px-4 py-2 text-center text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
        >
          就此宫问 AI
        </Link>
      </div>

      {related.length > 0 && (
        <div>
          <p className="mb-2 text-[11px] tracking-[0.08em] text-ink-faint">关联格局</p>
          <div className="flex flex-col gap-2">
            {related.map((p) => (
              <PatternCard key={p.name} p={p} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

/** 宫位详情侧栏(名人盘库等双栏页仍在用;排盘工作台已改用 PalaceDrawer)。 */
export function DetailPanel({
  chart, patterns, selectedBranch,
}: {
  chart: Chart;
  patterns: Pattern[];
  selectedBranch: number | null;
}) {
  const palace: Palace | null =
    selectedBranch != null
      ? chart.palaces.find((p) => p.branch === selectedBranch) ?? null
      : null;

  if (!palace) {
    return (
      <div className="rounded-[6px] bg-bg-raised p-4 text-[13px] text-ink-faint shadow-[0_0_0_1px_var(--line)]">
        <p className="mb-2 font-display text-[15px] font-semibold text-ink">格局总览</p>
        {patterns.length === 0 && <p>此盘未识别出典型格局。</p>}
        <div className="flex flex-col gap-2">
          {patterns.map((p) => (
            <PatternCard key={p.name} p={p} />
          ))}
        </div>
        <p className="mt-3">点击任意宫位查看星曜细目与三方四正。</p>
      </div>
    );
  }

  return <PalaceDetail palace={palace} patterns={patterns} />;
}
