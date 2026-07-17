"use client";

import { use, useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { fetchChapter, ApiError } from "@/lib/api";
import type { Chapter, Paragraph } from "@/lib/types";
import { ReaderToolbar } from "@/components/library/ReaderToolbar";
import { ReaderParagraph } from "@/components/library/ReaderParagraph";
import { BookmarkDrawer } from "@/components/library/BookmarkDrawer";
import { SkeletonLines } from "@/components/library/Skeleton";
import { useFontSize, saveProgress, loadProgress } from "@/components/library/prefs";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import {
  listReadingProgress,
  listBookmarks,
  addBookmark,
  deleteBookmark,
  syncProgressToServer,
  flushProgressToServer,
  ReadingError,
  type Bookmark,
  type ServerProgress,
} from "@/lib/reading";

interface ChapterData {
  book: { title: string; slug: string };
  index: number;
  total: number;
  chapter: Chapter;
}

interface ToastState {
  text: string;
  actionLabel?: string;
  onAction?: () => void;
  href?: string;
}

/** 命中锚点段落时,以 --gold-glow 背景 2 秒淡出的短暂金晕。 */
function glow(el: HTMLElement) {
  if (typeof el.animate !== "function") return;
  const color = getComputedStyle(document.documentElement).getPropertyValue("--gold-glow").trim();
  if (!color) return;
  el.animate(
    [{ backgroundColor: color }, { backgroundColor: "transparent" }],
    { duration: 2000, easing: "cubic-bezier(0.22, 1, 0.36, 1)" },
  );
}

/**
 * 阅读器(核心):三档字号、锚点/续读定位、三层文本、书签、章节导航、
 * 滚动进度条、返回顶部、移动端工具栏收纳。进度本地即时 + 服务端节流跨端同步。
 */
export default function ReaderPage({
  params,
}: {
  params: Promise<{ slug: string; chapter: string }>;
}) {
  const { slug, chapter } = use(params);
  const idx = Number(chapter);
  const router = useRouter();
  const [size, setSize] = useFontSize();
  const [data, setData] = useState<ChapterData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [signedIn, setSignedIn] = useState(false);
  const [bookmarks, setBookmarks] = useState<Bookmark[]>([]);
  const [bookmarksLoading, setBookmarksLoading] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [toast, setToast] = useState<ToastState | null>(null);

  const [scrollProgress, setScrollProgress] = useState(0);
  const [showBackToTop, setShowBackToTop] = useState(false);
  const [collapsed, setCollapsed] = useState(false);

  const canPersistRef = useRef(false);
  const serverProgRef = useRef<{ slug: string; list: ServerProgress[] } | null>(null);

  const loginHref = `/login?next=${encodeURIComponent(`/library/${slug}/${idx}`)}`;

  // ── 登录态感知 ──────────────────────────────────────
  useEffect(() => {
    const sync = () => setSignedIn(!!currentUser());
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, []);

  // ── 拉取章节 ────────────────────────────────────────
  useEffect(() => {
    if (!Number.isInteger(idx) || idx < 0) {
      setError("章节不存在");
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError("");
    setData(null);
    fetchChapter(slug, idx)
      .then((r) => {
        if (!cancelled) setData(r);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "章节加载失败");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [slug, idx]);

  // ── 本书书签(仅登录)────────────────────────────────
  useEffect(() => {
    if (!signedIn) {
      setBookmarks([]);
      return;
    }
    let cancelled = false;
    setBookmarksLoading(true);
    listBookmarks(slug)
      .then((bs) => {
        if (!cancelled) setBookmarks(bs);
      })
      .catch(() => {
        if (!cancelled) setBookmarks([]);
      })
      .finally(() => {
        if (!cancelled) setBookmarksLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [slug, signedIn]);

  const total = data?.total ?? 0;
  const hasPrev = idx > 0;
  const hasNext = data != null && idx < total - 1;
  const prevHref = `/library/${slug}/${idx - 1}`;
  const nextHref = `/library/${slug}/${idx + 1}`;

  // 当前章节内被收藏的段落 id 集合(段落金标记依据)
  const bookmarkedPids = useMemo(() => {
    const s = new Set<string>();
    for (const b of bookmarks) if (b.chapterIdx === idx) s.add(b.paragraphId);
    return s;
  }, [bookmarks, idx]);

  const backToTop = useCallback(() => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  }, []);

  // ── 键盘 ←/→ 翻章 ──────────────────────────────────
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.metaKey || e.ctrlKey || e.altKey || e.shiftKey) return;
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (e.key === "ArrowLeft" && hasPrev) {
        e.preventDefault();
        router.push(prevHref);
      } else if (e.key === "ArrowRight" && hasNext) {
        e.preventDefault();
        router.push(nextHref);
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [hasPrev, hasNext, prevHref, nextHref, router]);

  // ── 滚动:进度条 + 返回顶部 + 移动端工具栏收纳 ──────────
  useEffect(() => {
    setCollapsed(false);
    let lastY = window.scrollY;
    let raf = 0;
    const onScroll = () => {
      if (raf) return;
      raf = requestAnimationFrame(() => {
        raf = 0;
        const y = window.scrollY;
        const doc = document.documentElement;
        const max = doc.scrollHeight - doc.clientHeight;
        setScrollProgress(max > 0 ? Math.min(1, Math.max(0, y / max)) : 0);
        setShowBackToTop(y > 600);
        if (y > lastY + 8 && y > 220) setCollapsed(true);
        else if (y < lastY - 8) setCollapsed(false);
        lastY = y;
      });
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
    return () => {
      window.removeEventListener("scroll", onScroll);
      if (raf) cancelAnimationFrame(raf);
    };
  }, [idx]);

  // ── 续读定位 + 首个可见段落进度记录 ────────────────────
  useEffect(() => {
    if (!data) return;
    let cancelled = false;
    canPersistRef.current = false;
    const els = Array.from(document.querySelectorAll<HTMLElement>("[data-pid]"));

    // 首个可见段 → 进度(本地即时 + 服务端节流)。gate 前不落盘,避免覆盖续读源。
    const visible = new Set<string>();
    let timer: number | undefined;
    const persist = () => {
      if (!canPersistRef.current) return;
      const first = els.find((el) => el.dataset.pid && visible.has(el.dataset.pid)) ?? els[0];
      const pid = first?.dataset.pid ?? "";
      saveProgress(slug, { chapterIdx: idx, paragraphId: pid });
      syncProgressToServer(slug, idx, pid);
    };

    const hash = window.location.hash;

    async function restore() {
      // URL 带 #p-xxx:优先锚点定位并金晕高亮
      if (hash.startsWith("#p-")) {
        const el = document.getElementById(hash.slice(1));
        if (el) {
          requestAnimationFrame(() => {
            if (cancelled) return;
            el.scrollIntoView({ behavior: "smooth", block: "start" });
            glow(el);
          });
        }
        canPersistRef.current = true;
        persist();
        return;
      }

      // 无锚点:取该书进度(登录取服务端,未登录取本地)
      let prog: { chapterIdx: number; paragraphId: string } | null = null;
      if (signedIn) {
        try {
          let list = serverProgRef.current?.slug === slug ? serverProgRef.current.list : null;
          if (!list) {
            list = await listReadingProgress();
            serverProgRef.current = { slug, list };
          }
          const hit = list.find((p) => p.bookSlug === slug);
          if (hit) prog = { chapterIdx: hit.chapterIdx, paragraphId: hit.paragraphId };
        } catch {
          /* 服务端不可用时落本地兜底 */
        }
        if (!prog) prog = loadProgress(slug);
      } else {
        prog = loadProgress(slug);
      }
      if (cancelled) return;

      const firstPid = els[0]?.dataset.pid ?? "";
      if (prog && prog.chapterIdx === idx && prog.paragraphId && prog.paragraphId !== firstPid) {
        const el = document.getElementById(`p-${prog.paragraphId}`);
        if (el) {
          requestAnimationFrame(() => {
            if (cancelled) return;
            el.scrollIntoView({ behavior: "smooth", block: "start" });
            glow(el);
          });
          setToast({ text: "已回到上次阅读位置", actionLabel: "回到开头", onAction: backToTop });
          canPersistRef.current = true;
          persist();
          return;
        }
      }
      window.scrollTo({ top: 0 });
      canPersistRef.current = true;
      persist();
    }
    void restore();

    const io = new IntersectionObserver(
      (entries) => {
        for (const en of entries) {
          const pid = (en.target as HTMLElement).dataset.pid;
          if (!pid) continue;
          if (en.isIntersecting) visible.add(pid);
          else visible.delete(pid);
        }
        window.clearTimeout(timer);
        timer = window.setTimeout(persist, 300);
      },
      { rootMargin: "-140px 0px -55% 0px" },
    );
    els.forEach((el) => io.observe(el));
    return () => {
      cancelled = true;
      io.disconnect();
      window.clearTimeout(timer);
      flushProgressToServer(slug);
    };
  }, [data, slug, idx, signedIn, backToTop]);

  // ── 提示条数秒自动消失 ──────────────────────────────
  useEffect(() => {
    if (!toast) return;
    const t = window.setTimeout(() => setToast(null), 5200);
    return () => window.clearTimeout(t);
  }, [toast]);

  // ── 书签增删(乐观更新)──────────────────────────────
  const handleToggleBookmark = useCallback(
    async (p: Paragraph) => {
      if (!signedIn) {
        setToast({ text: "登录后可添加书签", actionLabel: "去登录", href: loginHref });
        return;
      }
      const existing = bookmarks.find((b) => b.chapterIdx === idx && b.paragraphId === p.id);
      try {
        if (existing) {
          setBookmarks((prev) => prev.filter((b) => b.id !== existing.id));
          await deleteBookmark(existing.id);
        } else {
          const bm = await addBookmark({
            bookSlug: slug,
            chapterIdx: idx,
            paragraphId: p.id,
            excerpt: p.text,
          });
          setBookmarks((prev) => (prev.some((b) => b.id === bm.id) ? prev : [bm, ...prev]));
        }
      } catch (e) {
        // 失败回滚:重新拉取真实状态
        try {
          setBookmarks(await listBookmarks(slug));
        } catch {
          /* ignore */
        }
        setToast({ text: e instanceof ReadingError ? e.message : "书签操作失败,请重试" });
      }
    },
    [signedIn, bookmarks, idx, slug, loginHref],
  );

  const handleDeleteBookmark = useCallback(
    async (id: string) => {
      setBookmarks((prev) => prev.filter((b) => b.id !== id));
      try {
        await deleteBookmark(id);
      } catch {
        try {
          setBookmarks(await listBookmarks(slug));
        } catch {
          /* ignore */
        }
      }
    },
    [slug],
  );

  return (
    <div>
      {/* 滚动进度条(顶栏之下细金线) */}
      <div className="fixed inset-x-0 top-14 z-40 h-[2px]" aria-hidden>
        <div className="h-full bg-gold" style={{ width: `${scrollProgress * 100}%` }} />
      </div>

      <ReaderToolbar
        slug={slug}
        bookTitle={data?.book.title ?? ""}
        chapterTitle={data?.chapter.title ?? ""}
        size={size}
        onSize={setSize}
        onOpenBookmarks={() => setDrawerOpen(true)}
        bookmarkCount={bookmarks.length}
        collapsed={collapsed}
      />

      <article
        className="mx-auto max-w-[68ch] px-5 py-8 md:py-12"
        style={{ fontSize: `${size}px`, lineHeight: 1.9 }}
      >
        {error && (
          <p className="rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
            {error}
          </p>
        )}

        {loading && !data && <SkeletonLines lines={10} />}

        {data && (
          <div key={idx} className="palace-enter">
            <header className="mb-12 text-center md:mb-16">
              <p className="tnum text-[12px] tracking-[0.28em] text-gold">
                第 {idx + 1} / {total} 章
              </p>
              <h1 className="mt-4 font-display text-[28px] font-semibold leading-tight text-ink md:text-[32px]">
                {data.chapter.title}
              </h1>
              {data.chapter.subtitle && (
                <p className="mt-2 font-reading text-[14px] text-ink-secondary">
                  {data.chapter.subtitle}
                </p>
              )}
              <span className="mx-auto mt-6 block h-px w-10 bg-gold-dim" aria-hidden />
            </header>

            <div className="font-reading">
              {data.chapter.paragraphs.map((p) => (
                <ReaderParagraph
                  key={p.id}
                  p={p}
                  bookmarked={bookmarkedPids.has(p.id)}
                  onToggleBookmark={handleToggleBookmark}
                />
              ))}
            </div>

            <nav className="mt-16 border-t border-line pt-10 font-body" aria-label="章节导航">
              <div className="grid gap-3 sm:grid-cols-2">
                {hasPrev ? (
                  <Link
                    href={prevHref}
                    className="group flex min-h-[56px] flex-col justify-center rounded-[10px] bg-bg-raised px-5 py-3 shadow-[0_0_0_1px_var(--line)] transition-shadow hover:shadow-[0_0_0_1px_var(--gold-dim)]"
                  >
                    <span className="text-[11px] tracking-[0.2em] text-ink-faint">上一章</span>
                    <span className="tnum mt-1 text-[14px] text-ink-secondary transition-colors group-hover:text-gold">
                      ← 第 {idx} 章
                    </span>
                  </Link>
                ) : (
                  <span aria-hidden className="hidden sm:block" />
                )}
                {hasNext ? (
                  <Link
                    href={nextHref}
                    className="group flex min-h-[56px] flex-col justify-center rounded-[10px] bg-bg-raised px-5 py-3 text-right shadow-[0_0_0_1px_var(--line)] transition-shadow hover:shadow-[0_0_0_1px_var(--gold-dim)]"
                  >
                    <span className="text-[11px] tracking-[0.2em] text-ink-faint">下一章</span>
                    <span className="tnum mt-1 text-[14px] text-ink-secondary transition-colors group-hover:text-gold">
                      第 {idx + 2} 章 →
                    </span>
                  </Link>
                ) : (
                  <span aria-hidden className="hidden sm:block" />
                )}
              </div>
              <div className="mt-6 text-center">
                <Link
                  href={`/library/${slug}`}
                  className="inline-flex min-h-[44px] items-center rounded-[6px] px-4 text-[13px] text-ink-faint transition-colors hover:text-gold"
                >
                  返回目录
                </Link>
              </div>
            </nav>
          </div>
        )}
      </article>

      {/* 返回顶部 + 阅读进度百分比 */}
      {showBackToTop && (
        <button
          type="button"
          onClick={backToTop}
          aria-label="返回顶部"
          title="返回顶部"
          className="tnum fixed bottom-6 right-5 z-30 flex h-12 w-12 flex-col items-center justify-center rounded-full bg-bg-raised text-gold shadow-[var(--elevation-2)] ring-1 ring-[var(--gold-dim)] transition-colors hover:text-gold-bright"
        >
          <span className="text-[12px] font-medium leading-none">{Math.round(scrollProgress * 100)}%</span>
          <span className="mt-0.5 text-[10px] leading-none text-ink-faint">顶</span>
        </button>
      )}

      {/* 轻提示条 */}
      {toast && (
        <div className="fixed inset-x-0 bottom-6 z-40 flex justify-center px-4" role="status">
          <div className="flex items-center gap-3 rounded-[10px] bg-bg-overlay px-4 py-2.5 text-[13px] text-ink shadow-[var(--elevation-2)] ring-1 ring-[var(--line-strong)]">
            <span>{toast.text}</span>
            {toast.actionLabel &&
              (toast.href ? (
                <Link
                  href={toast.href}
                  onClick={() => setToast(null)}
                  className="shrink-0 font-medium text-gold transition-colors hover:text-gold-bright"
                >
                  {toast.actionLabel}
                </Link>
              ) : (
                <button
                  type="button"
                  onClick={() => {
                    toast.onAction?.();
                    setToast(null);
                  }}
                  className="shrink-0 font-medium text-gold transition-colors hover:text-gold-bright"
                >
                  {toast.actionLabel}
                </button>
              ))}
          </div>
        </div>
      )}

      <BookmarkDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        slug={slug}
        signedIn={signedIn}
        bookmarks={bookmarks}
        loading={bookmarksLoading}
        onDelete={handleDeleteBookmark}
        loginHref={loginHref}
      />
    </div>
  );
}
