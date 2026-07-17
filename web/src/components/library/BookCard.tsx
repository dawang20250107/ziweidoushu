"use client";

import Link from "next/link";
import type { BookMeta } from "@/lib/types";

/** 卡片进度(章级);paragraphId 用于直达上次段落。 */
export interface CardProgress {
  chapterIdx: number;
  paragraphId: string;
}

/**
 * 书目卡:书名(宋体)· 朝代 · 作者 · 简介 · 章节/段落统计。
 * 有进度时整卡直达上次章节,并显示「读到 第 N 章」+ 细进度条;未读则「开始阅读」。
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
      className="group block rounded-[6px] bg-bg-raised p-5 shadow-[0_0_0_1px_var(--line)] transition-shadow hover:shadow-[0_0_0_1px_var(--gold-dim)]"
    >
      <div className="flex items-baseline justify-between gap-3">
        <h2 className="font-display text-xl font-semibold text-ink transition-colors group-hover:text-gold">
          {book.title}
        </h2>
        {book.dynasty && (
          <span className="shrink-0 text-[12px] tracking-[0.08em] text-ink-faint">{book.dynasty}</span>
        )}
      </div>
      {book.author && <p className="mt-1 text-[13px] text-ink-secondary">{book.author}</p>}
      {book.intro && (
        <p className="mt-3 line-clamp-3 font-reading text-[14px] leading-relaxed text-ink-secondary">
          {book.intro}
        </p>
      )}
      <div className="tnum mt-4 flex items-center gap-2.5 border-t border-line pt-3 text-[12px] text-ink-faint">
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

      {reading ? (
        <div className="mt-3">
          <div className="flex items-center justify-between text-[12px]">
            <span className="tnum text-gold">读到 第 {progress!.chapterIdx + 1} 章</span>
            <span className="text-ink-faint transition-colors group-hover:text-gold">继续阅读 →</span>
          </div>
          <div className="mt-1.5 h-[3px] overflow-hidden rounded-full bg-line" aria-hidden>
            <div className="h-full rounded-full bg-gold" style={{ width: `${fraction * 100}%` }} />
          </div>
        </div>
      ) : (
        <p className="mt-3 text-[12px] font-medium text-gold-dim transition-colors group-hover:text-gold">
          开始阅读 →
        </p>
      )}
    </Link>
  );
}
