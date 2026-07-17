"use client";

import Link from "next/link";
import type { BookMeta } from "@/lib/types";
import { Constellation } from "./Constellation";

/** 卡片进度(章级);paragraphId 用于直达上次段落。 */
export interface CardProgress {
  chapterIdx: number;
  paragraphId: string;
}

const RING_R = 9;
const RING_C = 2 * Math.PI * RING_R;

/**
 * 星牌:一枚立式典籍卡,如夜空中的星牌。
 * 顶部星象顶饰(各书点位不同)· 朝代眉标 · 书名大字 · 简介 · 章段字统计。
 * 有进度时整卡直达上次章节,并显示进度环 +「读到 第 N 章」;未读则「开始阅读」。
 * hover:整卡浮起(.lift)、边线转金、星象与书名点亮。
 */
export function BookCard({ book, progress }: { book: BookMeta; progress?: CardProgress | null }) {
  const reading = progress != null && progress.chapterIdx >= 0 && progress.chapterIdx < book.chapters;
  const href = reading
    ? `/library/${book.slug}/${progress!.chapterIdx}` +
      (progress!.paragraphId ? `#p-${progress!.paragraphId}` : "")
    : `/library/${book.slug}`;
  const fraction =
    reading && book.chapters > 0
      ? Math.min(1, Math.max(0, (progress!.chapterIdx + 1) / book.chapters))
      : 0;

  return (
    <Link
      href={href}
      aria-label={
        reading
          ? `${book.title} · 继续阅读 第 ${progress!.chapterIdx + 1} 章`
          : `${book.title} · 开始阅读`
      }
      className="lift group relative flex flex-col overflow-hidden rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)] hover:shadow-[0_0_0_1px_var(--gold-dim)]"
    >
      {/* 星象顶饰 */}
      <div className="mb-5 h-12 opacity-80 transition-opacity duration-500 group-hover:opacity-100">
        <Constellation seed={book.slug} active={reading} className="h-full" />
      </div>

      {/* 朝代眉标 */}
      {book.dynasty && (
        <p className="text-[11px] tracking-[0.24em] text-gold-dim transition-colors group-hover:text-gold">
          {book.dynasty}
        </p>
      )}

      {/* 书名 */}
      <h2 className="mt-2 font-display text-[26px] font-semibold leading-[1.2] text-ink transition-colors group-hover:text-gold">
        {book.title}
      </h2>

      {book.author && <p className="mt-2 text-[13px] text-ink-secondary">{book.author}</p>}

      {book.intro && (
        <p className="mt-4 line-clamp-3 font-reading text-[14px] leading-[1.75] text-ink-secondary">
          {book.intro}
        </p>
      )}

      {/* 底部:统计 + 进度(mt-auto 把底栏压到卡片下缘,对齐网格) */}
      <div className="mt-auto pt-6">
        <div className="tnum flex items-center gap-2.5 text-[12px] text-ink-faint">
          <span>{book.chapters} 章</span>
          <span aria-hidden>·</span>
          <span>{book.paragraphs} 段</span>
          {book.wordCount > 0 && (
            <>
              <span aria-hidden>·</span>
              <span>{book.wordCount.toLocaleString()} 字</span>
            </>
          )}
        </div>

        <div className="mt-4 flex items-center justify-between gap-3 border-t border-line pt-4">
          {reading ? (
            <>
              <span className="flex items-center gap-2.5">
                <svg viewBox="0 0 24 24" className="h-7 w-7 -rotate-90" aria-hidden>
                  <circle cx="12" cy="12" r={RING_R} fill="none" stroke="var(--line-strong)" strokeWidth="2.4" />
                  <circle
                    cx="12"
                    cy="12"
                    r={RING_R}
                    fill="none"
                    stroke="var(--gold)"
                    strokeWidth="2.4"
                    strokeLinecap="round"
                    strokeDasharray={RING_C}
                    strokeDashoffset={RING_C * (1 - fraction)}
                  />
                </svg>
                <span className="tnum text-[13px] text-gold">读到 第 {progress!.chapterIdx + 1} 章</span>
              </span>
              <span className="shrink-0 text-[13px] text-ink-faint transition-colors group-hover:text-gold">
                继续 →
              </span>
            </>
          ) : (
            <span className="text-[13px] font-medium text-gold-dim transition-colors group-hover:text-gold">
              开始阅读 →
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
