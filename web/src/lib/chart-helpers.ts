/** 星盘渲染辅助:宫格布局、三方四正、亮度/四化配色。 */
import { BRANCHES, STEMS } from "./types";
import type { SiHua, Star } from "./types";

/**
 * 4×4 盘面的地支固定位(grid-area 名):
 *   巳 午 未 申
 *   辰 [中宫]  酉
 *   卯 [中宫]  戌
 *   寅 丑 子 亥
 */
export const GRID_AREA_BY_BRANCH: Record<number, string> = {
  5: "si", 6: "wu", 7: "wei", 8: "shen",
  4: "chen", 9: "you",
  3: "mao", 10: "xu",
  2: "yin", 1: "chou", 0: "zi", 11: "hai",
};

export const BOARD_GRID_TEMPLATE = `
  "si   wu     wei    shen"
  "chen center center you"
  "mao  center center xu"
  "yin  chou   zi     hai"
`;

/** 星盘入场级联顺序(安星视觉:自寅宫顺行) */
export const ENTER_ORDER_BY_BRANCH: Record<number, number> = {
  2: 0, 3: 1, 4: 2, 5: 3, 6: 4, 7: 5, 8: 6, 9: 7, 10: 8, 11: 9, 0: 10, 1: 11,
};

/** 三方四正:本宫 + 官禄位(+4)+ 财帛位(+8)+ 对宫(+6) */
export function sanFangBranches(branch: number): number[] {
  return [branch, (branch + 4) % 12, (branch + 8) % 12, (branch + 6) % 12];
}

/** 庙旺得利平不陷 → 亮度 token(CSS 变量名) */
export function brightnessVar(brightness?: string): string {
  switch (brightness) {
    case "庙":
      return "var(--brightness-miao)";
    case "旺":
      return "var(--brightness-wang)";
    case "得":
    case "利":
    case "平":
      return "var(--brightness-mid)";
    case "不":
      return "var(--brightness-bu)";
    case "陷":
      return "var(--brightness-xian)";
    default:
      return "var(--ink)";
  }
}

export function sihuaVar(siHua: SiHua): string {
  switch (siHua) {
    case "禄":
      return "var(--sihua-lu)";
    case "权":
      return "var(--sihua-quan)";
    case "科":
      return "var(--sihua-ke)";
    case "忌":
      return "var(--sihua-ji)";
  }
}

export function stemName(stem: number): string {
  return STEMS[stem] ?? "";
}

export function branchName(branch: number): string {
  return BRANCHES[branch] ?? "";
}

/** 按类型分组宫内星曜(渲染分行) */
export function groupStars(stars: Star[]): {
  major: Star[];
  assist: Star[]; // 六吉/禄马/六煞(lucky+sha)
  adjective: Star[]; // 杂曜
} {
  const major: Star[] = [];
  const assist: Star[] = [];
  const adjective: Star[] = [];
  for (const s of stars ?? []) {
    if (s.type === "major") major.push(s);
    else if (s.type === "lucky" || s.type === "sha") assist.push(s);
    else adjective.push(s);
  }
  return { major, assist, adjective };
}

export type Density = "simple" | "pro" | "master";

export const DENSITY_LABELS: Record<Density, string> = {
  simple: "简洁",
  pro: "专业",
  master: "大师",
};
