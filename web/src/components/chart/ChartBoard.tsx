"use client";

import type { Chart, Horoscope, Star } from "@/lib/types";
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
}

/** 4×4 星盘:外环十二宫(地支固定位)+ 中宫命主信息。 */
export function ChartBoard({
  chart, density, selectedBranch, onSelectBranch, horoscope, overlayScopes = ["decadal", "yearly"],
}: ChartBoardProps) {
  const sanFang = selectedBranch != null ? new Set(sanFangBranches(selectedBranch)) : null;

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
      role="group"
      aria-label="紫微斗数命盘"
      className="grid gap-1.5"
      style={{
        gridTemplateAreas: BOARD_GRID_TEMPLATE,
        gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
        gridTemplateRows: "repeat(4, minmax(132px, auto))",
      }}
    >
      {chart.palaces.map((palace) => (
        <div key={palace.branch} style={{ gridArea: GRID_AREA_BY_BRANCH[palace.branch] }} className="flex">
          <PalaceCell
            palace={palace}
            density={density}
            selected={selectedBranch === palace.branch}
            inSanFang={sanFang != null && selectedBranch !== palace.branch && sanFang.has(palace.branch)}
            overlayStars={overlayStarsByBranch.get(palace.branch)}
            overlayNames={overlayNamesByBranch.get(palace.branch)}
            enterDelay={ENTER_ORDER_BY_BRANCH[palace.branch] * 36}
            onSelect={(b) => onSelectBranch(selectedBranch === b ? null : b)}
          />
        </div>
      ))}
      <div style={{ gridArea: "center" }} className="flex">
        <ChartCenter chart={chart} horoscope={horoscope ?? undefined} />
      </div>
    </div>
  );
}
