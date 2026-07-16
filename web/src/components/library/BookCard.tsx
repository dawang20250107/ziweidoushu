"use client";

import Link from "next/link";
import type { BookMeta } from "@/lib/types";

/** 书目卡:书名(宋体)· 朝代 · 作者 · 简介 · 章节/段落统计。 */
export function BookCard({ book }: { book: BookMeta }) {
  return (
    <Link
      href={`/library/${book.slug}`}
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
    </Link>
  );
}
