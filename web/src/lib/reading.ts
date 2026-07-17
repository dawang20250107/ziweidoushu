"use client";

/**
 * 古籍阅读:跨端续读进度 + 书签的 API 封装。
 * 统一响应信封 { ok, data, error:{code,message} };鉴权走 authFetch(自动带 Bearer、401 刷新重放)。
 * 进度写入内置 ≥2s 节流器(仅登录时才打服务端);本地记忆仍由 prefs.ts 承担,二者互补。
 */
import { authFetch, currentUser } from "@/lib/auth";

// ── 类型 ──────────────────────────────────────────────

/** 服务端阅读进度(单本书)。 */
export interface ServerProgress {
  bookSlug: string;
  chapterIdx: number;
  paragraphId: string;
  updatedAt: string;
}

/** 阅读书签。 */
export interface Bookmark {
  id: string;
  bookSlug: string;
  chapterIdx: number;
  paragraphId: string;
  excerpt: string;
  createdAt: string;
}

/** 带业务错误码的异常(unauthorized / invalid_chapter …)。 */
export class ReadingError extends Error {
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

async function parse<T>(res: Response): Promise<T> {
  let body: Envelope<T>;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    throw new ReadingError("network", `请求失败(${res.status})`, res.status);
  }
  if (!res.ok || !body.ok) {
    throw new ReadingError(
      body.error?.code ?? "unknown",
      body.error?.message ?? `请求失败(${res.status})`,
      res.status,
    );
  }
  return body.data as T;
}

const JSON_HEADERS = { "Content-Type": "application/json" };

/** 是否已登录(书签/服务端进度仅登录可用)。 */
export function isSignedIn(): boolean {
  return !!currentUser();
}

// ── 续读进度 ──────────────────────────────────────────

export async function listReadingProgress(): Promise<ServerProgress[]> {
  const res = await authFetch("/api/v1/me/reading");
  const data = await parse<{ progress: ServerProgress[] }>(res);
  return data.progress ?? [];
}

export async function putReadingProgress(
  slug: string,
  chapterIdx: number,
  paragraphId: string,
): Promise<void> {
  const res = await authFetch(`/api/v1/me/reading/${encodeURIComponent(slug)}`, {
    method: "PUT",
    headers: JSON_HEADERS,
    body: JSON.stringify({ chapterIdx, paragraphId }),
  });
  await parse<{ ok: boolean }>(res);
}

// ── 书签 ──────────────────────────────────────────────

export async function listBookmarks(book?: string): Promise<Bookmark[]> {
  const qs = book ? `?book=${encodeURIComponent(book)}` : "";
  const res = await authFetch(`/api/v1/me/bookmarks${qs}`);
  const data = await parse<{ bookmarks: Bookmark[] }>(res);
  return data.bookmarks ?? [];
}

export async function addBookmark(input: {
  bookSlug: string;
  chapterIdx: number;
  paragraphId: string;
  excerpt: string;
}): Promise<Bookmark> {
  const res = await authFetch("/api/v1/me/bookmarks", {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
  const data = await parse<{ bookmark: Bookmark }>(res);
  return data.bookmark;
}

export async function deleteBookmark(id: string): Promise<void> {
  const res = await authFetch(`/api/v1/me/bookmarks/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  await parse<{ deleted: boolean }>(res);
}

// ── 进度节流器:滚动高频触发 → 每 ≥2s 落一次服务端 ──────────────

const MIN_INTERVAL = 2000;
interface SyncSlot {
  last: number;
  timer?: number;
  pending?: { chapterIdx: number; paragraphId: string };
}
const slots: Record<string, SyncSlot> = {};

function fire(slug: string) {
  const slot = slots[slug];
  if (!slot?.pending) return;
  const { chapterIdx, paragraphId } = slot.pending;
  slot.pending = undefined;
  slot.timer = undefined;
  slot.last = Date.now();
  // 静默失败:进度非关键路径,失败不打扰阅读(下次滚动会再试)
  void putReadingProgress(slug, chapterIdx, paragraphId).catch(() => {});
}

/**
 * 节流上报进度到服务端(仅登录时)。前沿即时、后沿补发,保证最终一致。
 * 未登录静默跳过,由本地 prefs 记忆兜底。
 */
export function syncProgressToServer(
  slug: string,
  chapterIdx: number,
  paragraphId: string,
): void {
  if (!isSignedIn()) return;
  const slot = (slots[slug] ??= { last: 0 });
  slot.pending = { chapterIdx, paragraphId };
  const elapsed = Date.now() - slot.last;
  if (elapsed >= MIN_INTERVAL) {
    fire(slug);
  } else if (slot.timer == null) {
    slot.timer = window.setTimeout(() => fire(slug), MIN_INTERVAL - elapsed);
  }
}

/** 立刻落盘挂起的进度(离开章节/卸载时调用),避免丢最后位置。 */
export function flushProgressToServer(slug: string): void {
  const slot = slots[slug];
  if (!slot?.pending || !isSignedIn()) return;
  if (slot.timer != null) {
    window.clearTimeout(slot.timer);
    slot.timer = undefined;
  }
  fire(slug);
}
