"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { fetchBooks, ApiError } from "@/lib/api";
import type { BookMeta } from "@/lib/types";
import { SearchBox } from "@/components/library/SearchBox";
import { BookCard, type CardProgress } from "@/components/library/BookCard";
import { SkeletonCard } from "@/components/library/Skeleton";
import { loadProgress } from "@/components/library/prefs";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import { listReadingProgress } from "@/lib/reading";

/** 书架:典籍卡片网格 + 顶部检索入口(回车跳检索页)+ 每本书阅读进度。 */
export default function LibraryPage() {
  const router = useRouter();
  const [q, setQ] = useState("");
  const [books, setBooks] = useState<BookMeta[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [progressMap, setProgressMap] = useState<Record<string, CardProgress>>({});

  useEffect(() => {
    let cancelled = false;
    fetchBooks()
      .then((r) => {
        if (!cancelled) setBooks(r.books);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "书目加载失败");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // 进度:登录取服务端(跨端),未登录取本地。登录态变化即重算。
  useEffect(() => {
    if (books.length === 0) return;
    let cancelled = false;

    const computeLocal = () => {
      const map: Record<string, CardProgress> = {};
      for (const b of books) {
        const p = loadProgress(b.slug);
        if (p) map[b.slug] = { chapterIdx: p.chapterIdx, paragraphId: p.paragraphId };
      }
      if (!cancelled) setProgressMap(map);
    };

    const sync = () => {
      if (currentUser()) {
        listReadingProgress()
          .then((list) => {
            if (cancelled) return;
            const map: Record<string, CardProgress> = {};
            for (const p of list) {
              map[p.bookSlug] = { chapterIdx: p.chapterIdx, paragraphId: p.paragraphId };
            }
            setProgressMap(map);
          })
          .catch(() => computeLocal());
      } else {
        computeLocal();
      }
    };

    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => {
      cancelled = true;
      window.removeEventListener(AUTH_EVENT, sync);
    };
  }, [books]);

  function toSearch(query: string) {
    const s = query.trim();
    if (s) router.push(`/library/search?q=${encodeURIComponent(s)}`);
  }

  return (
    <div className="mx-auto max-w-5xl px-5 py-8 md:py-12">
      <header className="mb-8">
        <p className="text-[12px] tracking-[0.24em] text-gold">古籍 · 典藏</p>
        <h1 className="mt-2 font-display text-3xl font-semibold text-ink">书架</h1>
        <p className="mt-2 text-ink-secondary">紫微斗数历代典籍,宣纸上细读。</p>
        <div className="mt-5 max-w-xl">
          <SearchBox
            value={q}
            onChange={setQ}
            onSubmit={toSearch}
            placeholder="检索全文…书名、星曜、格局"
          />
        </div>
      </header>

      {error && (
        <p className="rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {loading && (
        <div className="grid gap-4 sm:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      )}

      {!loading && books.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2">
          {books.map((b) => (
            <BookCard key={b.slug} book={b} progress={progressMap[b.slug] ?? null} />
          ))}
        </div>
      )}

      {!loading && !error && books.length === 0 && (
        <p className="py-16 text-center text-ink-faint">暂无典籍。</p>
      )}
    </div>
  );
}
