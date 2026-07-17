"use client";

import { useState } from "react";
import type { SiZhuView } from "@/lib/types";

/**
 * 四柱视角(八字附加层):与紫微同源同盘——四柱干支之上展开
 * 五行、十神、藏干、纳音与五行分布。可折叠,默认收起。
 */

// 五行 → 语义色变量(木青 火朱 土赭 金曜 水墨蓝)
const ELEMENT_VAR: Record<string, string> = {
  木: "var(--ok)",
  火: "var(--danger)",
  土: "var(--warn)",
  金: "var(--gold)",
  水: "var(--info)",
};

function ElementChar({ char, element, size = 28 }: { char: string; element: string; size?: number }) {
  return (
    <span
      className="font-display font-semibold leading-none"
      style={{ color: ELEMENT_VAR[element] ?? "var(--ink)", fontSize: size }}
    >
      {char}
    </span>
  );
}

export function SiZhuPanel({ siZhu }: { siZhu: SiZhuView }) {
  const [open, setOpen] = useState(false);
  const total = Object.values(siZhu.elementCount).reduce((a, b) => a + b, 0) || 8;

  return (
    <section className="rounded-[10px] bg-bg-raised shadow-[0_0_0_1px_var(--line)]">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        className="flex min-h-[52px] w-full items-center justify-between px-5 py-3 text-left"
      >
        <span className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
          <span className="text-[12px] tracking-[0.24em] text-gold">四柱视角</span>
          <span className="text-[13px] text-ink-faint">
            八字与紫微同源同盘 · 日主
            <span className="ml-1 font-display text-[15px]" style={{ color: ELEMENT_VAR[siZhu.dayMasterElement] }}>
              {siZhu.dayMaster}{siZhu.dayMasterElement}
            </span>
          </span>
          {siZhu.geJu && (
            <span className="rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
              月令{siZhu.geJu.name}
            </span>
          )}
        </span>
        <svg
          width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
          aria-hidden
          className={`text-ink-faint transition-transform ${open ? "rotate-180" : ""}`}
        >
          <path d="M6 9l6 6 6-6" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>

      {open && (
        <div className="border-t border-line px-5 pb-6 pt-4">
          {/* 四柱表 */}
          <div className="grid grid-cols-4 gap-2 sm:gap-4">
            {siZhu.pillars.map((p) => (
              <div key={p.name} className="flex flex-col items-center gap-1.5 rounded-[8px] bg-bg px-2 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
                <span className="text-[11px] tracking-[0.16em] text-ink-faint">{p.name}</span>
                <ElementChar char={p.stem} element={p.stemElement} />
                <ElementChar char={p.branch} element={p.branchElement} />
                <span
                  className={[
                    "mt-1 rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none",
                    p.stemShiShen === "日主"
                      ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                      : "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]",
                  ].join(" ")}
                >
                  {p.stemShiShen}
                </span>
                {/* 藏干 */}
                <div className="mt-1 flex flex-col items-center gap-0.5">
                  {p.hidden.map((h) => (
                    <span key={h.stem} className="text-[11px] leading-tight text-ink-faint">
                      <span style={{ color: ELEMENT_VAR[h.element] }}>{h.stem}</span>
                      <span className="ml-1">{h.shiShen}</span>
                    </span>
                  ))}
                </div>
                <span className="mt-1 text-[11px] text-ink-faint">{p.naYin}</span>
              </div>
            ))}
          </div>

          {/* 月令取格(子平真诠) */}
          {siZhu.geJu && (
            <div className="mt-5 rounded-[8px] bg-bg px-4 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
              <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
                <span className="text-[12px] tracking-[0.08em] text-ink-faint">月令取格</span>
                <span className="font-display text-[17px] font-semibold text-ink">{siZhu.geJu.name}</span>
                <span className="text-[12px] text-ink-faint">{siZhu.geJu.basis}</span>
              </div>
              <p className="mt-2 text-[13px] leading-relaxed text-ink-secondary">{siZhu.geJu.note}</p>
              <p className="mt-1.5 text-[11px] text-ink-faint">
                {siZhu.geJu.source} · 取格述格局之体,吉凶成败仍须通盘参详
              </p>
            </div>
          )}

          {/* 五行分布 */}
          <div className="mt-5">
            <p className="mb-2 text-[12px] tracking-[0.08em] text-ink-faint">五行分布(四干四支)</p>
            <div className="flex flex-col gap-1.5">
              {(["木", "火", "土", "金", "水"] as const).map((el) => {
                const n = siZhu.elementCount[el] ?? 0;
                return (
                  <div key={el} className="flex items-center gap-2">
                    <span className="w-4 text-[13px]" style={{ color: ELEMENT_VAR[el] }}>{el}</span>
                    <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-bg">
                      <div
                        className="h-full rounded-full transition-[width] duration-500"
                        style={{ width: `${(n / total) * 100}%`, background: ELEMENT_VAR[el], opacity: 0.75 }}
                      />
                    </div>
                    <span className="tnum w-6 text-right text-[12px] text-ink-secondary">{n}</span>
                  </div>
                );
              })}
            </div>
            {Object.entries(siZhu.elementCount).some(([, v]) => v === 0) && (
              <p className="mt-2 text-[12px] text-ink-faint">
                缺
                {Object.entries(siZhu.elementCount)
                  .filter(([, v]) => v === 0)
                  .map(([k]) => k)
                  .join("、")}
                (以本气计,藏干未入分布)
              </p>
            )}
          </div>
        </div>
      )}
    </section>
  );
}
