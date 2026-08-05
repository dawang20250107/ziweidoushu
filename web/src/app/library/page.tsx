"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { fetchBooks, ApiError } from "@/lib/api";
import type { BookMeta } from "@/lib/types";
import { SearchBox } from "@/components/library/SearchBox";
import { BookCard, type CardProgress } from "@/components/library/BookCard";
import { ResumeHero } from "@/components/library/ResumeHero";
import { SkeletonCard } from "@/components/library/Skeleton";
import { loadProgress } from "@/components/library/prefs";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import { listReadingProgress } from "@/lib/reading";

/**
 * 星阁:典籍星牌网格 + 顶部续读 hero(登录且有进度时)+ 全文检索入口。
 * 每枚星牌带阅读进度;登录取服务端进度(跨端),未登录取本地 prefs。
 */
export default function LibraryPage() {
  const router = useRouter();
  const [q, setQ] = useState("");
  const [books, setBooks] = useState<BookMeta[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [progressMap, setProgressMap] = useState<Record<string, CardProgress>>({});
  const [recentSlug, setRecentSlug] = useState<string | null>(null); // 续读 hero:最近在读(仅登录)

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

  // 进度:登录取服务端(跨端 + 最近在读),未登录取本地。登录态变化即重算。
  useEffect(() => {
    if (books.length === 0) return;
    let cancelled = false;

    const computeLocal = () => {
      const map: Record<string, CardProgress> = {};
      for (const b of books) {
        const p = loadProgress(b.slug);
        if (p) map[b.slug] = { chapterIdx: p.chapterIdx, paragraphId: p.paragraphId };
      }
      if (!cancelled) {
        setProgressMap(map);
        setRecentSlug(null); // 续读 hero 仅登录态呈现
      }
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
            // 最近在读:按 updatedAt 倒序取首(ISO 串字典序即时间序)
            const recent = [...list].sort((a, b) =>
              (b.updatedAt ?? "").localeCompare(a.updatedAt ?? ""),
            )[0];
            setRecentSlug(recent ? recent.bookSlug : null);
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

  // 续读 hero 数据:最近在读的书 + 有效进度
  const heroBook = recentSlug ? books.find((b) => b.slug === recentSlug) ?? null : null;
  const heroProgress = heroBook ? progressMap[heroBook.slug] ?? null : null;
  const heroValid =
    heroBook != null &&
    heroProgress != null &&
    heroProgress.chapterIdx >= 0 &&
    heroProgress.chapterIdx < heroBook.chapters;

  return (
    <div className="mx-auto max-w-6xl px-5 pb-24 pt-14 md:pt-20">
      {/* 顶部:续读 hero 或星阁标题区 */}
      {heroValid ? (
        <ResumeHero book={heroBook!} progress={heroProgress!} />
      ) : (
        <header className="max-w-2xl">
          <p className="text-[12px] tracking-[0.24em] text-gold">古籍 · 典藏</p>
          <h1 className="mt-4 font-display text-5xl font-semibold leading-[1.05] text-ink md:text-6xl">
            星阁
          </h1>
          <p className="mt-6 text-[17px] leading-[1.7] text-ink-secondary">
            紫微斗数历代典籍,列如夜空一阁星牌。拾级而上,于宣纸上细读古人星语。
          </p>
        </header>
      )}

      {/* 星牌书架 */}
      <section className="mt-24" aria-label="典籍星牌">
        <div className="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-[12px] tracking-[0.24em] text-gold-dim">古籍</p>
            <h2 className="mt-1.5 font-display text-xl font-semibold text-ink">
              典籍{books.length > 0 ? ` · ${books.length} 卷` : ""}
            </h2>
          </div>
          <div className="w-full sm:w-72">
            <SearchBox
              value={q}
              onChange={setQ}
              onSubmit={toSearch}
              placeholder="检索全文…书名、星曜、格局"
            />
          </div>
        </div>

        {error && (
          <p className="mt-8 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
            {error}
          </p>
        )}

        {loading && (
          <div className="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        )}

        {!loading && books.length > 0 && (
          <div className="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {books.map((b) => (
              /* 星牌随下滑逐一入焦(scroll-driven,渐进增强) */
              <div key={b.slug} className="reveal">
                <BookCard book={b} progress={progressMap[b.slug] ?? null} />
              </div>
            ))}
          </div>
        )}

        {!loading && !error && books.length === 0 && (
          <p className="py-16 text-center text-ink-faint">暂无典籍。</p>
        )}
      </section>
    </div>
  );
}
