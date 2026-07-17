"use client";

import { use, useEffect, useState } from "react";
import Link from "next/link";
import { fetchBook, ApiError } from "@/lib/api";
import type { Book } from "@/lib/types";
import { loadProgress, type ReadingProgress } from "@/components/library/prefs";
import { SkeletonLines } from "@/components/library/Skeleton";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import { listReadingProgress } from "@/lib/reading";

/** 书详情:简介 + 目录(各章段落数)+「继续阅读」(localStorage 进度)。 */
export default function BookDetailPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = use(params);
  const [book, setBook] = useState<Book | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [progress, setProgress] = useState<ReadingProgress | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");
    fetchBook(slug)
      .then((b) => {
        if (!cancelled) setBook(b);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "书籍加载失败");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [slug]);

  // 进度:登录取服务端(跨端续读),未登录或失败取本地。登录态变化即重算。
  useEffect(() => {
    let cancelled = false;
    const sync = () => {
      if (currentUser()) {
        listReadingProgress()
          .then((list) => {
            if (cancelled) return;
            const hit = list.find((p) => p.bookSlug === slug);
            setProgress(hit ? { chapterIdx: hit.chapterIdx, paragraphId: hit.paragraphId } : loadProgress(slug));
          })
          .catch(() => {
            if (!cancelled) setProgress(loadProgress(slug));
          });
      } else {
        setProgress(loadProgress(slug));
      }
    };
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => {
      cancelled = true;
      window.removeEventListener(AUTH_EVENT, sync);
    };
  }, [slug]);

  const totalParagraphs = book?.chapters.reduce((n, c) => n + c.paragraphs.length, 0) ?? 0;
  const resumeValid = progress != null && book != null && progress.chapterIdx < book.chapters.length;
  const resumeChapter = resumeValid ? book!.chapters[progress!.chapterIdx] : null;
  const resumeHref = resumeValid
    ? `/library/${slug}/${progress!.chapterIdx}` +
      (progress!.paragraphId ? `#p-${progress!.paragraphId}` : "")
    : "";

  return (
    <div className="mx-auto max-w-3xl px-5 py-8 md:py-12">
      <nav className="mb-5 flex items-center gap-1.5 text-[13px] text-ink-faint" aria-label="面包屑">
        <Link href="/library" className="transition-colors hover:text-gold">
          书架
        </Link>
        <span aria-hidden>/</span>
        <span className="truncate text-ink-secondary">{book?.title ?? "…"}</span>
      </nav>

      {error && (
        <p className="rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {loading && !book && (
        <div className="flex flex-col gap-8">
          <SkeletonLines lines={4} />
          <SkeletonLines lines={8} />
        </div>
      )}

      {book && (
        <>
          <header className="border-b border-line pb-8">
            <h1 className="font-display text-3xl font-semibold leading-tight text-ink">{book.title}</h1>
            <div className="tnum mt-3 flex flex-wrap items-center gap-2.5 text-[13px] text-ink-secondary">
              {book.dynasty && <span>{book.dynasty}</span>}
              {book.author && (
                <>
                  <span aria-hidden className="text-ink-faint">·</span>
                  <span>{book.author}</span>
                </>
              )}
              <span aria-hidden className="text-ink-faint">·</span>
              <span className="text-ink-faint">
                {book.chapters.length} 章 · {totalParagraphs} 段
                {book.wordCount > 0 ? ` · ${book.wordCount.toLocaleString()} 字` : ""}
              </span>
            </div>

            {book.intro && (
              <p className="mt-5 font-reading text-[15px] leading-[1.9] text-ink-secondary">
                {book.intro}
              </p>
            )}

            <div className="mt-6 flex flex-wrap gap-3">
              {resumeChapter ? (
                <Link
                  href={resumeHref}
                  className="rounded-[6px] px-4 py-2 text-[14px] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-colors hover:text-gold-bright"
                  style={{ background: "var(--gold-glow)" }}
                >
                  继续阅读 · 第 {progress!.chapterIdx + 1} 章「{resumeChapter.title}」
                </Link>
              ) : (
                book.chapters.length > 0 && (
                  <Link
                    href={`/library/${slug}/0`}
                    className="rounded-[6px] px-4 py-2 text-[14px] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-colors hover:text-gold-bright"
                    style={{ background: "var(--gold-glow)" }}
                  >
                    开始阅读
                  </Link>
                )
              )}
            </div>
          </header>

          <section className="mt-8" aria-label="目录">
            <h2 className="mb-3 text-[12px] tracking-[0.24em] text-gold">目录</h2>
            <ol className="divide-y divide-line">
              {book.chapters.map((ch, i) => {
                const current = progress?.chapterIdx === i;
                return (
                  <li key={i}>
                    <Link
                      href={`/library/${slug}/${i}`}
                      className="flex items-baseline gap-3 rounded-[6px] px-2 py-3 transition-colors hover:bg-bg-raised"
                      aria-current={current ? "true" : undefined}
                    >
                      <span
                        className={`tnum w-7 shrink-0 text-[13px] ${current ? "text-gold" : "text-gold-dim"}`}
                      >
                        {String(i + 1).padStart(2, "0")}
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="font-display text-[16px] text-ink">{ch.title}</span>
                        {ch.subtitle && (
                          <span className="ml-2 text-[13px] text-ink-faint">{ch.subtitle}</span>
                        )}
                      </span>
                      <span className="tnum shrink-0 text-[12px] text-ink-faint">
                        {ch.paragraphs.length} 段
                      </span>
                    </Link>
                  </li>
                );
              })}
            </ol>
          </section>
        </>
      )}
    </div>
  );
}
