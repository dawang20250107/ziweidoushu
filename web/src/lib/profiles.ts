/**
 * 命盘档案 & 深度报告的 API 封装。
 * 统一响应信封 { ok, data, error:{code,message} };鉴权走 authFetch(自动带 Bearer、401 刷新重放)。
 * 本模块保持自包含(仅依赖 @/lib/auth),供 SaveProfileButton 等零外部依赖组件安全引用。
 */
import { authFetch } from "@/lib/auth";

// ── 类型 ──────────────────────────────────────────────

/** 生辰输入(与排盘一致):hour 为时辰索引 0-12。 */
export interface BirthRequest {
  year: number;
  month: number;
  day: number;
  hour: number;
  gender: "male" | "female";
  name?: string;
  longitude?: number;
  province?: string;
  city?: string;
  trueSolarTime?: boolean;
}

export type Relation = "self" | "family" | "friend" | "client" | "other";

export interface Profile {
  id: string;
  label: string;
  relation: Relation;
  birthInput: BirthRequest;
  engineVersion: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ReportResult {
  text: string;
  provider: string;
  degraded: boolean;
}

export interface Entitlements {
  tier: string;
  entitlements: Record<string, unknown>;
  credits: { deep_report?: number };
}

/** 带业务错误码的异常(profile_limit / no_credits / ai_unavailable / invalid_birth …)。 */
export class ProfileError extends Error {
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

/** 解析统一信封;非 2xx 或 ok:false 抛 ProfileError。 */
async function parse<T>(res: Response): Promise<T> {
  let body: Envelope<T>;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    throw new ProfileError("network", `请求失败(${res.status})`, res.status);
  }
  if (!res.ok || !body.ok) {
    throw new ProfileError(
      body.error?.code ?? "unknown",
      body.error?.message ?? `请求失败(${res.status})`,
      res.status,
    );
  }
  return body.data as T;
}

const JSON_HEADERS = { "Content-Type": "application/json" };

// ── 档案 CRUD ─────────────────────────────────────────

/** limit=0 表示不限;>0 时前端显示「n/limit」。 */
export async function listProfiles(): Promise<{ profiles: Profile[]; limit: number }> {
  const res = await authFetch("/api/v1/profiles");
  return parse(res);
}

export async function createProfile(
  input: BirthRequest & { label: string; relation?: Relation; isDefault?: boolean },
): Promise<{ profile: Profile }> {
  const res = await authFetch("/api/v1/profiles", {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
  return parse(res);
}

export async function deleteProfile(id: string): Promise<{ deleted: boolean }> {
  const res = await authFetch(`/api/v1/profiles/${encodeURIComponent(id)}`, { method: "DELETE" });
  return parse(res);
}

export async function setDefaultProfile(id: string): Promise<{ ok: boolean }> {
  const res = await authFetch(`/api/v1/profiles/${encodeURIComponent(id)}/default`, { method: "POST" });
  return parse(res);
}

// ── 深度报告 & 权益 ───────────────────────────────────

export async function generateReport(
  input: BirthRequest & { topic: string },
): Promise<{ report: ReportResult; topic: string; remainingCredits: number }> {
  const res = await authFetch("/api/v1/ai/report", {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
  return parse(res);
}

export async function fetchEntitlements(): Promise<Entitlements> {
  const res = await authFetch("/api/v1/me/entitlements");
  return parse(res);
}

// ── 跨页契约:最近一次排盘生辰 ────────────────────────

export const BIRTH_KEY = "ziwei-birth";

/** 读取排盘页写入的最近生辰;无或损坏返回 null。 */
export function readRecentBirth(): BirthRequest | null {
  try {
    const raw = localStorage.getItem(BIRTH_KEY);
    return raw ? (JSON.parse(raw) as BirthRequest) : null;
  } catch {
    return null;
  }
}

/** 把档案生辰写回,供排盘页 /chart 读取重排。 */
export function writeRecentBirth(b: BirthRequest): void {
  try {
    localStorage.setItem(BIRTH_KEY, JSON.stringify(b));
  } catch {
    // localStorage 不可用时静默(隐私模式等)
  }
}

// ── 展示辅助 ──────────────────────────────────────────

/** 关系徽标选项(下拉与徽标共用)。 */
export const RELATIONS: { key: Relation; label: string }[] = [
  { key: "self", label: "本人" },
  { key: "family", label: "家人" },
  { key: "friend", label: "朋友" },
  { key: "client", label: "客户" },
  { key: "other", label: "其他" },
];

export function relationLabel(r?: Relation): string {
  return RELATIONS.find((x) => x.key === r)?.label ?? "其他";
}

// 时辰索引 → 地支(0=早子…11=亥…12=晚子,均取地支「子丑寅卯辰巳午未申酉戌亥」)
const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];

/** 生辰摘要:如「1990 年 8 月 16 日 · 子时 · 女」。 */
export function formatBirthSummary(b: BirthRequest): string {
  const branch = BRANCHES[((b.hour % 12) + 12) % 12] ?? "";
  const gender = b.gender === "male" ? "男" : "女";
  return `${b.year} 年 ${b.month} 月 ${b.day} 日 · ${branch}时 · ${gender}`;
}
