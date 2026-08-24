import { fetchStarKnowledge, type StarKnowledge, type StarLoreEntry } from "./api";

export type { StarKnowledge, StarLoreEntry };

// 亮度七等要义(通行口径,前端静态图例)。
export const BRIGHTNESS_MEANING: Record<string, string> = {
  庙: "光华最盛,吉力全发",
  旺: "明亮有力,发挥顺畅",
  得: "得地可用,稳定施展",
  利: "尚可施展,略打折扣",
  平: "平常之力,吉凶随伴",
  不: "乏力失色,须借吉扶",
  陷: "失辉无力,缺点易显",
};

const FLOW_PREFIX = "运流月日时";

let cache: Promise<StarKnowledge> | null = null;

/** 全量星曜知识(档案/四大十二神/流曜),模块级缓存,一次拉取全站复用。 */
export function getStarKnowledge(): Promise<StarKnowledge> {
  if (!cache) {
    cache = fetchStarKnowledge().catch((e) => {
      cache = null; // 失败不缓存,允许下次重试
      throw e;
    });
  }
  return cache;
}

/** 按星名查档:本命星曜直查;流曜(运魁/流禄…)按去前缀字义查 flow。 */
export function loreOf(lib: StarKnowledge | null, name: string): StarLoreEntry | null {
  if (!lib) return null;
  const direct = lib.lore[name];
  if (direct) return direct;
  const r = [...name];
  if (r.length === 2 && FLOW_PREFIX.includes(r[0])) {
    const f = lib.flow[r[1]];
    if (f) return { si: "流曜", gist: f };
  }
  return null;
}
