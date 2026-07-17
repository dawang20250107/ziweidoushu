/**
 * 占卜(梅花易数 + 小六壬)API client + 类型契约。
 * 后端统一信封:{ ok, data } / { ok: false, error: { code, message } }。
 * 起卦与卦象展示免费(匿名 fetch);AI 深度解卦按次付费(authFetch,消耗 divination)。
 *
 * 起卦由服务端推导,客户端不可伪造。关键契约:AI 解卦须回传与用户所见「同一卦」——
 *   时间卦:记住起卦返回的 castAt,原样回传;数字卦:回传同一组 numbers。
 */
import { authFetch } from "@/lib/auth";

const BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";

// ── 类型(与 Go 后端 meihua 包逐字段对应)──────────────

/** 八卦(先天数 1-8)。lines 为三爻,自下而上,true=阳爻。 */
export interface Trigram {
  num: number;
  name: string; // 乾/兑/离/震/巽/坎/艮/坤
  symbol: string; // ☰☱☲☳☴☵☶☷
  nature: string; // 天/泽/火/雷/风/水/山/地
  element: string; // 金/木/水/火/土
  lines: boolean[]; // ×3
}

/** 六十四卦。lines 为六爻,自下而上(1-6 爻),true=阳爻。 */
export interface Hexagram {
  name: string;
  upper: Trigram;
  lower: Trigram;
  lines: boolean[]; // ×6
}

/** 体用生克关系。 */
export type Relation = "用生体" | "比和" | "体克用" | "体生用" | "用克体";

/** 一次梅花起卦的完整卦象。 */
export interface MeihuaResult {
  method: "time" | "number";
  question?: string;
  lunarText?: string; // 时间卦:「午年六月初四日申时」
  numbers?: number[]; // 数字卦原始数
  ben: Hexagram; // 本卦
  hu: Hexagram; // 互卦
  bian: Hexagram; // 变卦
  moving: number; // 动爻 1-6
  tiTrigram: Trigram; // 体卦
  yongTrigram: Trigram; // 用卦
  tiIsUpper: boolean; // 体在上卦
  relation: Relation;
  verdict: string; // 吉凶倾向一句话
}

/** 小六壬六位之一。 */
export interface LiuRenPos {
  name: string; // 大安/留连/速喜/赤口/小吉/空亡
  luck: "吉" | "凶" | "平";
  meaning: string;
}

/** 小六壬占算结果。 */
export interface XiaoLiuRenResult {
  question?: string;
  lunarText: string; // 「正月初一子时」
  steps: string[]; // 月/日/时三步落位名 ×3
  result: LiuRenPos;
  path: LiuRenPos[]; // 三步完整落位 ×3
}

/** AI 解卦读物。 */
export interface DivineReading {
  text: string;
  provider: string;
}

// ── 错误类型 ──────────────────────────────────────────

/** 带业务错误码的异常(no_credits / ai_unavailable / cast_failed …)。 */
export class DivinationError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

interface Envelope<T> {
  ok: boolean;
  data?: T;
  error?: { code: string; message: string };
}

/** 解析统一信封;非 2xx 或 ok:false 抛 DivinationError。 */
async function parse<T>(res: Response): Promise<T> {
  let body: Envelope<T>;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    throw new DivinationError("network", `请求失败(${res.status})`, res.status);
  }
  if (!res.ok || !body.ok) {
    throw new DivinationError(
      body.error?.code ?? "unknown",
      body.error?.message ?? `请求失败(${res.status})`,
      res.status,
    );
  }
  return body.data as T;
}

const JSON_HEADERS = { "Content-Type": "application/json" };

// ── 起卦入参 ──────────────────────────────────────────

export interface CastInput {
  method: "time" | "number";
  numbers?: number[]; // 数字卦:[n,n] 或 [n,n,n]
  castAt?: number; // 时间卦:unix 秒(回传同一卦时用)
  question?: string;
}

// ── 接口 ──────────────────────────────────────────────

/** 梅花易数起卦(免费,匿名)。返回卦象与服务端起卦时刻 castAt。 */
export async function castMeihua(input: CastInput): Promise<{ result: MeihuaResult; castAt: number }> {
  const res = await fetch(`${BASE}/api/v1/divination/meihua`, {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
  return parse(res);
}

/** 小六壬快占(免费,匿名)。 */
export async function castXiaoLiuRen(input?: { question?: string; castAt?: number }): Promise<{ result: XiaoLiuRenResult }> {
  const res = await fetch(`${BASE}/api/v1/divination/xiaoliuren`, {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input ?? {}),
  });
  return parse(res);
}

/** AI 深度解卦(需登录,消耗 1 次 divination)。question 必填;卦象参数须与所见同一卦。 */
export async function divineAI(
  input: CastInput & { question: string },
): Promise<{ result: MeihuaResult; reading: DivineReading; remainingCredits: number }> {
  const res = await authFetch(`${BASE}/api/v1/ai/divine`, {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
  return parse(res);
}

/** 剩余解卦次数(需登录)。读 data.credits.divination,缺省 0。 */
export async function fetchDivinationCredits(): Promise<number> {
  const res = await authFetch(`${BASE}/api/v1/me/entitlements`);
  const data = await parse<{ credits?: Record<string, number | undefined> }>(res);
  return data.credits?.divination ?? 0;
}

// ── 展示辅助(纯函数,页面与组件共用)────────────────

/** 语义色调:金(gold)/成(ok)/戒(warn)/危(danger)。 */
export type Tone = "gold" | "ok" | "warn" | "danger";

/** 体用生克 → 吉凶断语 + 语义色。 */
export const RELATION_TONE: Record<Relation, { label: string; tone: Tone }> = {
  用生体: { label: "用生体 · 大吉", tone: "gold" },
  比和: { label: "比和 · 吉", tone: "ok" },
  体克用: { label: "体克用 · 小吉", tone: "ok" },
  体生用: { label: "体生用 · 小凶", tone: "warn" },
  用克体: { label: "用克体 · 凶", tone: "danger" },
};

/** 小六壬吉凶 → 语义色。 */
export function luckTone(luck: LiuRenPos["luck"]): Tone {
  if (luck === "吉") return "ok";
  if (luck === "凶") return "danger";
  return "warn"; // 平
}

/** 卦名之上下卦类象:如「上乾天 · 下兑泽」。 */
export function trigramLine(h: Hexagram): string {
  return `上${h.upper.name}${h.upper.nature} · 下${h.lower.name}${h.lower.nature}`;
}
