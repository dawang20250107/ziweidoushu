"use client";

import { useEffect } from "react";
import type { Chart, Pattern } from "@/lib/types";
import { branchName } from "@/lib/chart-helpers";
import { PalaceDetail } from "./DetailPanel";

/**
 * 宫位详情抽屉:选宫时浮出,不占页面栏位——
 * 桌面自右缘滑入的竖幅面板(盘面保持全宽可见,三方四正金线不被遮挡),
 * 移动端为底部上滑抽屉(带遮罩,点遮罩或 Esc 关闭)。内部滚动,长内容不撑页面。
 */
export function PalaceDrawer({
  chart, patterns, selectedBranch, onClose,
}: {
  chart: Chart;
  patterns: Pattern[];
  selectedBranch: number | null;
  onClose: () => void;
}) {
  const palace =
    selectedBranch != null
      ? chart.palaces.find((p) => p.branch === selectedBranch) ?? null
      : null;

  useEffect(() => {
    if (!palace) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [palace, onClose]);

  if (!palace) return null;

  return (
    <>
      {/* 移动端遮罩(桌面无遮罩:盘面保持可点,便于连续切宫对照) */}
      <div aria-hidden className="fixed inset-0 z-30 bg-black/50 lg:hidden" onClick={onClose} />
      <aside
        role="dialog"
        aria-label={`${palace.name}(${branchName(palace.branch)}宫)详情`}
        className={[
          "drawer-panel fixed z-40 overflow-y-auto overscroll-contain bg-bg shadow-[0_0_0_1px_var(--line-strong),0_18px_60px_rgba(0,0,0,0.5)]",
          // 移动端:底部抽屉
          "inset-x-0 bottom-0 max-h-[74vh] rounded-t-[14px] p-4 pb-6",
          // 桌面:右缘竖幅
          "lg:inset-x-auto lg:bottom-6 lg:right-5 lg:top-24 lg:max-h-none lg:w-[360px] lg:rounded-[10px] lg:p-4",
        ].join(" ")}
      >
        <div className="mb-3 flex items-center justify-between">
          <span className="text-[12px] tracking-[0.24em] text-gold">宫位详情</span>
          <button
            type="button"
            onClick={onClose}
            aria-label="关闭宫位详情"
            className="flex h-8 w-8 items-center justify-center rounded-[6px] text-ink-secondary transition-colors hover:text-ink hover:shadow-[inset_0_0_0_1px_var(--line)]"
          >
            ✕
          </button>
        </div>
        {/* 移动端抽屉提手 */}
        <span aria-hidden className="absolute left-1/2 top-1.5 h-1 w-9 -translate-x-1/2 rounded-full bg-[var(--line-strong)] lg:hidden" />
        <PalaceDetail palace={palace} patterns={patterns} />
      </aside>
    </>
  );
}
