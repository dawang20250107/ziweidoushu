/**
 * 计费 API client + 类型契约。
 * 后端统一响应包裹:{ ok, data } / { ok: false, error: { code, message } }。
 * 公开接口(products)走匿名 fetch;其余走 authFetch(自动带 Bearer + 401 重放)。
 */
import { authFetch } from "@/lib/auth";

const BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";

// ── 类型(与 Go 后端 JSON 逐字段对应)────────────────────

export type ProductKind = "subscription" | "credits";
export type PaidTier = "pro" | "master";
export type Tier = "free" | "pro" | "master";

export interface Product {
  id: string;
  title: string;
  kind: ProductKind;
  tier?: PaidTier;
  durationDays?: number;
  creditType?: string;
  creditAmount?: number;
  priceCents: number;
  originalPriceCents?: number;
}

export interface ProductsResponse {
  products: Product[];
  devPayEnabled: boolean;
}

export type OrderStatus =
  | "created"
  | "paying"
  | "paid"
  | "fulfilled"
  | "closed"
  | "refunded";

export interface Order {
  id: string;
  orderNo: string;
  productId: string;
  amountCents: number;
  status: OrderStatus;
  channel?: string;
  paidAt?: string;
  fulfilledAt?: string;
  expiresAt: string;
  createdAt: string;
}

export interface MeUser {
  id: string;
  nickname: string;
  avatar: string;
  tier: Tier;
  tierExpiresAt?: string;
}

export interface Entitlement {
  tier: string;
  startsAt: string;
  endsAt: string;
}

export interface EntitlementsResponse {
  tier: Tier;
  entitlements: Entitlement[] | null;
  credits: { deep_report?: number };
}

// ── 错误类型 ──────────────────────────────────────────

export class BillingError extends Error {
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

function unwrap<T>(res: Response, body: Envelope<T>): T {
  if (!res.ok || !body.ok) {
    throw new BillingError(
      body.error?.code ?? "unknown",
      body.error?.message ?? `请求失败(${res.status})`,
      res.status,
    );
  }
  return body.data as T;
}

/** 匿名 GET(公开接口)。 */
async function getPublic<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
  });
  return unwrap(res, (await res.json()) as Envelope<T>);
}

/** 鉴权请求。 */
async function authRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await authFetch(`${BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  return unwrap(res, (await res.json()) as Envelope<T>);
}

// ── 接口 ──────────────────────────────────────────────

export function fetchProducts(): Promise<ProductsResponse> {
  return getPublic<ProductsResponse>("/api/v1/products");
}

export function createOrder(productId: string): Promise<{ order: Order; devPayEnabled: boolean }> {
  return authRequest("/api/v1/orders", { method: "POST", body: JSON.stringify({ productId }) });
}

export function fetchOrders(): Promise<{ orders: Order[] }> {
  return authRequest("/api/v1/orders");
}

export function devPayOrder(id: string): Promise<{ order: Order }> {
  return authRequest(`/api/v1/orders/${encodeURIComponent(id)}/dev-pay`, { method: "POST" });
}

export function fetchMe(): Promise<{ user: MeUser }> {
  return authRequest("/api/v1/me");
}

export function fetchEntitlements(): Promise<EntitlementsResponse> {
  return authRequest("/api/v1/me/entitlements");
}

// ── 展示辅助(纯函数,组件与页面共用)────────────────────

/** 分 → ¥xx.xx */
export function formatPrice(cents: number): string {
  return `¥${(cents / 100).toFixed(2)}`;
}

/** 折扣百分比(originalPriceCents 更高时返回省下的百分比,否则 0)。 */
export function discountPercent(priceCents: number, originalPriceCents?: number): number {
  if (!originalPriceCents || originalPriceCents <= priceCents) return 0;
  return Math.round((1 - priceCents / originalPriceCents) * 100);
}

/** 订阅时长 → 中文标签(年/月/天)。 */
export function durationLabel(days?: number): string {
  if (!days) return "";
  if (days >= 360) return `${Math.max(1, Math.round(days / 365))} 年`;
  if (days >= 28) return `${Math.round(days / 30)} 个月`;
  return `${days} 天`;
}

/** 会员层级中文名。 */
export const TIER_LABEL: Record<Tier, string> = {
  free: "体验版",
  pro: "Pro",
  master: "大师版",
};

/** 次卡类型中文名。 */
export const CREDIT_LABEL: Record<string, string> = {
  deep_report: "深度报告",
};

export function creditLabel(type?: string): string {
  if (!type) return "次数";
  return CREDIT_LABEL[type] ?? type;
}

/** 订单状态中文名。 */
export const ORDER_STATUS_LABEL: Record<OrderStatus, string> = {
  created: "待支付",
  paying: "支付中",
  paid: "已支付",
  fulfilled: "已完成",
  closed: "已关闭",
  refunded: "已退款",
};

/** 权益点文案(定价卡与成功态共用)。 */
export const PRO_BENEFITS = ["无限命盘档案", "深度运限下钻", "古籍全文检索"];
export const MASTER_BENEFITS = ["Pro 全部权益", "双人合盘分析", "优先 AI 解读队列"];
export const CREDIT_BENEFITS = ["深度报告 AI 生成", "生成失败自动退还", "购买次数永久有效"];

/** 依产品推导权益点列表。 */
export function benefitsFor(product: Product): string[] {
  if (product.kind === "credits") return CREDIT_BENEFITS;
  return product.tier === "master" ? MASTER_BENEFITS : PRO_BENEFITS;
}

/** 本地化日期(年月日)。 */
export function formatDate(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" });
}

/** 本地化日期时间(年月日 时:分)。 */
export function formatDateTime(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
