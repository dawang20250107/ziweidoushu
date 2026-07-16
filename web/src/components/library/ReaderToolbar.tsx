"use client";

import Link from "next/link";
import { FONT_SIZES, type FontSize } from "./prefs";

const SIZE_LABEL: Record<FontSize, string> = { 18: "小", 20: "中", 23: "大" };

/** 阅读器顶部工具条:面包屑 + 字号三档(移动端固定于此,不遮正文)。 */
export function ReaderToolbar({
  slug,
  bookTitle,
  chapterTitle,
  size,
  onSize,
}: {
  slug: string;
  bookTitle: string;
  chapterTitle: string;
  size: FontSize;
  onSize: (s: FontSize) => void;
}) {
  return (
    <div className="sticky top-14 z-30 border-b border-line bg-bg/90 backdrop-blur">
      <div className="mx-auto flex max-w-[74ch] items-center justify-between gap-3 px-5 py-2">
        <nav
          className="flex min-w-0 items-center gap-1.5 text-[13px] text-ink-faint"
          aria-label="面包屑"
        >
          <Link href="/library" className="shrink-0 transition-colors hover:text-gold">
            书架
          </Link>
          <span aria-hidden>/</span>
          <Link
            href={`/library/${slug}`}
            className="max-w-[7rem] shrink-0 truncate font-display transition-colors hover:text-gold"
          >
            {bookTitle || "…"}
          </Link>
          <span aria-hidden>/</span>
          <span className="truncate text-ink-secondary">{chapterTitle || "…"}</span>
        </nav>

        <div
          className="flex shrink-0 overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]"
          role="radiogroup"
          aria-label="正文字号"
        >
          {FONT_SIZES.map((s) => (
            <button
              key={s}
              type="button"
              role="radio"
              aria-checked={size === s}
              aria-label={`字号${SIZE_LABEL[s]}`}
              onClick={() => onSize(s)}
              className={[
                "px-3 py-1 text-[13px] transition-colors",
                size === s ? "text-gold" : "text-ink-secondary hover:text-ink",
              ].join(" ")}
              style={size === s ? { background: "var(--gold-glow)" } : undefined}
            >
              {SIZE_LABEL[s]}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
