"use client";

/**
 * 鉴权客户端:access 存内存、refresh 存 localStorage,
 * 401 自动刷新一次并重放请求;登录态变化通过自定义事件广播。
 */

export interface AuthUser {
  id: string;
  nickname: string;
  avatar: string;
  tier: "free" | "pro" | "master";
}

interface TokenPair {
  access: string;
  refresh: string;
  expiresAt: string;
}

const REFRESH_KEY = "ziwei-refresh";
const USER_KEY = "ziwei-user";
export const AUTH_EVENT = "ziwei-auth-changed";

let accessToken: string | null = null;

function emitChange() {
  window.dispatchEvent(new Event(AUTH_EVENT));
}

export function currentUser(): AuthUser | null {
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? (JSON.parse(raw) as AuthUser) : null;
  } catch {
    return null;
  }
}

function saveSession(tokens: TokenPair, user: AuthUser) {
  accessToken = tokens.access;
  localStorage.setItem(REFRESH_KEY, tokens.refresh);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
  emitChange();
}

export function clearSession() {
  accessToken = null;
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(USER_KEY);
  emitChange();
}

interface Envelope<T> {
  ok: boolean;
  data?: T;
  error?: { code: string; message: string };
}

async function post<T>(path: string, payload: unknown, access?: string): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (access) headers.Authorization = `Bearer ${access}`;
  const res = await fetch(path, { method: "POST", headers, body: JSON.stringify(payload) });
  const body = (await res.json()) as Envelope<T>;
  if (!res.ok || !body.ok) {
    throw new Error(body.error?.message ?? `请求失败(${res.status})`);
  }
  return body.data as T;
}

// ── 登录流程 ──────────────────────────────────────────

export async function sendCode(phone: string): Promise<{ devCode?: string }> {
  return post("/api/v1/auth/sms/send", { phone });
}

export async function verifyCode(phone: string, code: string): Promise<AuthUser> {
  const data = await post<{ tokens: TokenPair; user: AuthUser }>("/api/v1/auth/sms/verify", { phone, code });
  saveSession(data.tokens, data.user);
  return data.user;
}

/** 刷新访问令牌;失败即清会话。
 * 单飞:并发调用共享同一次刷新请求。refresh 是旋转式一次性令牌,
 * 若并发各自刷新,第二个请求会触发服务端复用检测导致全端下线。 */
let refreshInFlight: Promise<boolean> | null = null;

function refreshAccess(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    const refresh = localStorage.getItem(REFRESH_KEY);
    if (!refresh) return false;
    try {
      const data = await post<{ tokens: TokenPair; user: AuthUser }>("/api/v1/auth/refresh", { refresh });
      saveSession(data.tokens, data.user);
      return true;
    } catch {
      clearSession();
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

export async function logout(): Promise<void> {
  const refresh = localStorage.getItem(REFRESH_KEY);
  try {
    const access = await ensureAccess();
    if (access) {
      await post("/api/v1/auth/logout", { refresh: refresh ?? "" }, access);
    }
  } catch {
    // 服务端撤销失败不阻塞本地登出
  } finally {
    clearSession();
  }
}

/** 取有效 access(必要时刷新);未登录返回 null。 */
export async function ensureAccess(): Promise<string | null> {
  if (accessToken) return accessToken;
  if (await refreshAccess()) return accessToken;
  return null;
}

/** 带鉴权的 fetch:401 时自动刷新一次并重放。 */
export async function authFetch(input: string, init?: RequestInit): Promise<Response> {
  const access = await ensureAccess();
  const withAuth = (token: string | null): RequestInit => ({
    ...init,
    headers: { ...init?.headers, ...(token ? { Authorization: `Bearer ${token}` } : {}) },
  });
  let res = await fetch(input, withAuth(access));
  if (res.status === 401 && access) {
    accessToken = null;
    const renewed = await ensureAccess();
    if (renewed) res = await fetch(input, withAuth(renewed));
  }
  return res;
}
