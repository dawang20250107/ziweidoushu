/**
 * 类型化 API client。
 * 后端统一响应包裹:{ ok, data } / { ok: false, error: { code, message } }。
 */
import type {
  BirthInfo, ChartResponse, Horoscope, BookMeta, Book, Chapter,
  SearchHit, InterpretResult, FamousPerson, HemingResponse, Chart, Pattern,
  WorldCity, ProvinceCities, HoroscopeReading,
} from "./types";

const BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";

export class ApiError extends Error {
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

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  const body = (await res.json()) as Envelope<T>;
  if (!res.ok || !body.ok) {
    throw new ApiError(
      body.error?.code ?? "unknown",
      body.error?.message ?? `请求失败(${res.status})`,
      res.status,
    );
  }
  return body.data as T;
}

function get<T>(path: string): Promise<T> {
  return request<T>(path);
}

function post<T>(path: string, payload: unknown): Promise<T> {
  return request<T>(path, { method: "POST", body: JSON.stringify(payload) });
}

// ── 排盘 ──────────────────────────────────────────────

export function fetchChart(birth: BirthInfo): Promise<ChartResponse> {
  return post<ChartResponse>("/api/v1/chart", birth);
}

export function fetchWorldCities(): Promise<{ cities: WorldCity[] }> {
  return get("/api/v1/world-cities");
}

export function fetchChinaCities(): Promise<{ provinces: ProvinceCities[] }> {
  return get("/api/v1/cities");
}

export function fetchHoroscope(
  birth: BirthInfo,
  target: { year: number; month: number; day: number; hour: number },
): Promise<{ horoscope: Horoscope; reading?: HoroscopeReading }> {
  return post<{ horoscope: Horoscope; reading?: HoroscopeReading }>("/api/v1/horoscope", { ...birth, target });
}

// ── 合盘与名人 ────────────────────────────────────────

export function fetchHeming(a: BirthInfo, b: BirthInfo): Promise<HemingResponse> {
  return post<HemingResponse>("/api/v1/heming", { a, b });
}

export function fetchFamousList(): Promise<{ persons: FamousPerson[] }> {
  return get("/api/v1/famous");
}

export function fetchFamousChart(id: string): Promise<{
  person: FamousPerson;
  chart: Chart;
  patterns: Pattern[];
}> {
  return get(`/api/v1/famous/${encodeURIComponent(id)}/chart`);
}

// ── 古籍 ──────────────────────────────────────────────

export function fetchBooks(): Promise<{ books: BookMeta[]; stats: Record<string, number> }> {
  return get("/api/v1/books");
}

export function fetchBook(slug: string): Promise<Book> {
  return get(`/api/v1/books/${encodeURIComponent(slug)}`);
}

export function fetchChapter(slug: string, idx: number): Promise<{
  book: { title: string; slug: string };
  index: number;
  total: number;
  chapter: Chapter;
}> {
  return get(`/api/v1/books/${encodeURIComponent(slug)}/chapters/${idx}`);
}

export function searchClassics(q: string, limit = 30): Promise<{
  query: string;
  count: number;
  hits: SearchHit[];
}> {
  return get(`/api/v1/search?q=${encodeURIComponent(q)}&limit=${limit}`);
}

// ── AI 解读(SSE 流式)────────────────────────────────

export interface InterpretParams extends BirthInfo {
  topic?: string;
  question?: string;
}

/**
 * 流式解读:onDelta 逐段回调;返回完整结果。
 * AbortSignal 支持取消(离开页面/重新提问)。
 */
export async function streamInterpret(
  params: InterpretParams,
  onDelta: (text: string) => void,
  signal?: AbortSignal,
): Promise<InterpretResult> {
  const res = await fetch(`${BASE}/api/v1/ai/interpret`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...params, stream: true }),
    signal,
  });
  if (!res.ok || !res.body) {
    throw new ApiError("ai_failed", `AI 服务不可用(${res.status})`, res.status);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let full = "";
  let meta: { provider: string; degraded: boolean } = { provider: "", degraded: false };

  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    // SSE 帧以空行分隔
    const frames = buffer.split("\n\n");
    buffer = frames.pop() ?? "";
    for (const frame of frames) {
      let event = "message";
      let data = "";
      for (const line of frame.split("\n")) {
        if (line.startsWith("event: ")) event = line.slice(7).trim();
        else if (line.startsWith("data: ")) data += line.slice(6);
      }
      if (!data) continue;
      if (event === "delta") {
        const payload = JSON.parse(data) as { text: string };
        full += payload.text;
        onDelta(payload.text);
      } else if (event === "done") {
        meta = JSON.parse(data) as { provider: string; degraded: boolean };
      } else if (event === "error") {
        const payload = JSON.parse(data) as { message: string };
        throw new ApiError("ai_failed", payload.message, 502);
      }
    }
  }
  return { text: full, provider: meta.provider, degraded: meta.degraded };
}
