"use client";

import { useLayoutEffect, useRef, useState } from "react";
import type { Chart, Horoscope, Pattern, Star } from "@/lib/types";
import {
  BOARD_GRID_TEMPLATE, ENTER_ORDER_BY_BRANCH, GRID_AREA_BY_BRANCH,
  sanFangBranches, type Density,
} from "@/lib/chart-helpers";
import { PalaceCell } from "./PalaceCell";
import { ChartCenter } from "./ChartCenter";

export interface ChartBoardProps {
  chart: Chart;
  density: Density;
  selectedBranch: number | null;
  onSelectBranch: (branch: number | null) => void;
  /** 运限叠加(时间轴激活时传入) */
  horoscope?: Horoscope | null;
  /** 叠加显示哪些层(默认大限+流年) */
  overlayScopes?: Array<"decadal" | "yearly" | "monthly" | "daily" | "hourly">;
  /** 已识别格局(中宫徽章锚点) */
  patterns?: Pattern[];
  /** 格局联动:悬停格局时点亮其关联宫位(宫名列表),其余降暗 */
  highlightNames?: string[] | null;
  /** 中宫格局徽章悬停回调(向上冒泡驱动 highlightNames) */
  onPatternHover?: (names: string[] | null) => void;
}

interface ConnectLine {
  key: string;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
}

/** 4×4 星盘:外环十二宫(地支固定位)+ 中宫命主信息 + 三方四正金线。 */
export function ChartBoard({
  chart, density, selectedBranch, onSelectBranch, horoscope, overlayScopes = ["decadal", "yearly"], patterns,
  highlightNames, onPatternHover,
}: ChartBoardProps) {
  const sanFang = selectedBranch != null ? new Set(sanFangBranches(selectedBranch)) : null;
  const hlSet = highlightNames && highlightNames.length > 0 ? new Set(highlightNames) : null;
  const boardRef = useRef<HTMLDivElement>(null);
  const [lines, setLines] = useState<ConnectLine[]>([]);

  // 选宫 → 量取宫格中心,画三方四正连线(布局变化与缩放时重量)
  useLayoutEffect(() => {
    const board = boardRef.current;
    if (!board || selectedBranch == null) {
      setLines([]);
      return;
    }
    function measure() {
      if (!board) return;
      const origin = board.getBoundingClientRect();
      const center = (branch: number) => {
        const el = board.querySelector(`[data-branch="${branch}"]`);
        if (!el) return null;
        const r = el.getBoundingClientRect();
        return { x: r.left - origin.left + r.width / 2, y: r.top - origin.top + r.height / 2 };
      };
      const from = center(selectedBranch!);
      if (!from) return;
      const next: ConnectLine[] = [];
      for (const b of sanFangBranches(selectedBranch!)) {
        if (b === selectedBranch) continue;
        const to = center(b);
        if (to) next.push({ key: `${selectedBranch}-${b}`, x1: from.x, y1: from.y, x2: to.x, y2: to.y });
      }
      setLines(next);
    }
    measure();
    // ResizeObserver:选宫让位平移/窗口缩放期间连线逐帧追踪宫心,不会错位
    const ro = new ResizeObserver(measure);
    ro.observe(board);
    window.addEventListener("resize", measure);
    return () => {
      ro.disconnect();
      window.removeEventListener("resize", measure);
    };
  }, [selectedBranch, density, horoscope, chart]);

  // 组装每宫的运限叠加数据
  const overlayStarsByBranch = new Map<number, Star[]>();
  const overlayNamesByBranch = new Map<number, { scope: string; name: string }[]>();
  if (horoscope) {
    for (const key of overlayScopes) {
      const scope = horoscope[key];
      if (!scope) continue;
      if (scope.stars) {
        scope.stars.forEach((cell, branch) => {
          if (!cell || cell.length === 0) return;
          const list = overlayStarsByBranch.get(branch) ?? [];
          list.push(...cell);
          overlayStarsByBranch.set(branch, list);
        });
      }
      scope.palaceNames.forEach((name, branch) => {
        const list = overlayNamesByBranch.get(branch) ?? [];
        list.push({ scope: scope.name, name });
        overlayNamesByBranch.set(branch, list);
      });
    }
  }

  return (
    <div
      ref={boardRef}
      role="group"
      aria-label="紫微斗数命盘"
      className="relative grid gap-1.5"
      style={{
        gridTemplateAreas: BOARD_GRID_TEMPLATE,
        gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
        gridTemplateRows: "repeat(4, minmax(132px, auto))",
      }}
    >
      {chart.palaces.map((palace) => (
        <div
          key={palace.branch}
          data-branch={palace.branch}
          style={{ gridArea: GRID_AREA_BY_BRANCH[palace.branch] }}
          className="flex"
        >
          <PalaceCell
            palace={palace}
            density={density}
            selected={selectedBranch === palace.branch}
            inSanFang={sanFang != null && selectedBranch !== palace.branch && sanFang.has(palace.branch)}
            patternGlow={hlSet != null && hlSet.has(palace.name)}
            dimmed={
              (sanFang != null && !sanFang.has(palace.branch)) ||
              (hlSet != null && !hlSet.has(palace.name))
            }
            overlayStars={overlayStarsByBranch.get(palace.branch)}
            overlayNames={overlayNamesByBranch.get(palace.branch)}
            enterDelay={ENTER_ORDER_BY_BRANCH[palace.branch] * 36}
            onSelect={(b) => onSelectBranch(selectedBranch === b ? null : b)}
          />
        </div>
      ))}
      <div style={{ gridArea: "center" }} className="flex">
        <ChartCenter chart={chart} horoscope={horoscope ?? undefined} patterns={patterns} onPatternHover={onPatternHover} />
      </div>

      {/* 三方四正金线:选宫时从本宫射向对宫与三合宫 */}
      {lines.length > 0 && (
        <svg aria-hidden className="pointer-events-none absolute inset-0 z-10 h-full w-full">
          {lines.map((l) => (
            <g key={l.key}>
              {/* 辉光底衬:宽而淡的金光,让连线像光束而非细线 */}
              <line
                x1={l.x1} y1={l.y1} x2={l.x2} y2={l.y2}
                pathLength={1}
                className="sanfang-line"
                stroke="var(--gold)"
                strokeWidth="5"
                strokeLinecap="round"
                opacity="0.14"
              />
              <line
                x1={l.x1} y1={l.y1} x2={l.x2} y2={l.y2}
                pathLength={1}
                className="sanfang-line"
                stroke="var(--gold-dim)"
                strokeWidth="1.5"
                opacity="0.85"
              />
              <circle cx={l.x2} cy={l.y2} r="5.5" fill="var(--gold)" opacity="0.18" />
              <circle cx={l.x2} cy={l.y2} r="3" fill="var(--gold)" opacity="0.9" />
            </g>
          ))}
          {lines[0] && (
            <>
              <circle cx={lines[0].x1} cy={lines[0].y1} r="7" fill="var(--gold)" opacity="0.2" />
              <circle cx={lines[0].x1} cy={lines[0].y1} r="4" fill="var(--gold-bright)" />
            </>
          )}
        </svg>
      )}
    </div>
  );
}
