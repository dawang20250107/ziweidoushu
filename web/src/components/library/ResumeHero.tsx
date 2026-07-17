"use client";

import Link from "next/link";
import type { BookMeta } from "@/lib/types";
import type { CardProgress } from "./BookCard";
import { Constellation } from "./Constellation";

/**
 * 续读 hero:登录且有进度时置于星阁顶部的大幅横卡。
 * 书名大字(font-display)· 朝代作者 · 金色细进度条 · 大「继续阅读」主 CTA。
 * CTA 为本视区唯一辉光(bg-gold + .glow-gold)。
 */
export function ResumeHero({ book, progress }: { book: BookMeta; progress: CardProgress }) {
  const chapterNo = progress.chapterIdx + 1;
  const href =
    `/library/${book.slug}/${progress.chapterIdx}` +
    (progress.paragraphId ? `#p-${progress.paragraphId}` : "");
  const fraction =
    book.chapters > 0 ? Math.min(1, Math.max(0, chapterNo / book.chapters)) : 0;

  return (
    <section
      aria-label="继续阅读"
      className="relative overflow-hidden rounded-[10px] bg-bg-raised p-8 shadow-[0_0_0_1px_var(--line-strong)] md:p-12"
    >
      {/* 右上角星象氛围(在读态点亮) */}
      <div className="pointer-events-none absolute -right-4 -top-2 hidden w-64 opacity-60 sm:block" aria-hidden>
        <Constellation seed={book.slug} active className="h-24" />
      </div>

      <div className="relative max-w-2xl">
        <p className="text-[12px] tracking-[0.24em] text-gold">最近在读</p>

        <h2 className="mt-4 font-display text-[31px] font-semibold leading-[1.15] text-ink [text-wrap:balance] sm:text-4xl md:text-5xl">
          {book.title}
        </h2>

        <div className="tnum mt-4 flex flex-wrap items-center gap-2.5 text-[14px] text-ink-secondary">
          {book.dynasty && <span>{book.dynasty}</span>}
          {book.author && (
            <>
              <span aria-hidden className="text-ink-faint">·</span>
              <span>{book.author}</span>
            </>
          )}
        </div>

        {/* 进度 */}
        <div className="mt-8 max-w-md">
          <div className="tnum flex items-baseline justify-between text-[13px]">
            <span className="text-gold">读到 第 {chapterNo} 章</span>
            <span className="text-ink-faint">共 {book.chapters} 章</span>
          </div>
          <div
            className="mt-2.5 h-[3px] overflow-hidden rounded-full"
            style={{ background: "var(--line)" }}
            aria-hidden
          >
            <div className="h-full rounded-full bg-gold" style={{ width: `${fraction * 100}%` }} />
          </div>
        </div>

        <div className="mt-8">
          <Link
            href={href}
            className="glow-gold inline-flex min-h-[44px] items-center gap-2 rounded-[6px] bg-gold px-6 text-[15px] font-semibold text-[#161206] transition-colors hover:bg-gold-bright"
          >
            继续阅读
            <svg
              viewBox="0 0 20 20"
              className="h-4 w-4"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.8"
              aria-hidden
            >
              <path d="M4 10h11M11 5.5 15.5 10 11 14.5" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </Link>
        </div>
      </div>
    </section>
  );
}
