"use client";

import Link from "next/link";
import { FONT_SIZES, type FontSize } from "./prefs";

const SIZE_LABEL: Record<FontSize, string> = { 18: "小", 20: "中", 23: "大" };

/**
 * 阅读器顶部工具条:面包屑 + 书签入口 + 字号三档。
 * 移动端下滑时收纳(translate 隐入顶栏之后),上滑即现;桌面端(md+)恒显。
 */
export function ReaderToolbar({
  slug,
  bookTitle,
  chapterTitle,
  size,
  onSize,
  onOpenBookmarks,
  bookmarkCount = 0,
  collapsed = false,
}: {
  slug: string;
  bookTitle: string;
  chapterTitle: string;
  size: FontSize;
  onSize: (s: FontSize) => void;
  onOpenBookmarks?: () => void;
  bookmarkCount?: number;
  collapsed?: boolean;
}) {
  return (
    <div
      className={[
        "sticky top-14 z-30 border-b border-line bg-bg/90 backdrop-blur",
        "transition-transform duration-200 ease-out md:!translate-y-0",
        collapsed ? "-translate-y-full" : "translate-y-0",
      ].join(" ")}
    >
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

        <div className="flex shrink-0 items-center gap-2">
          {onOpenBookmarks && (
            <button
              type="button"
              onClick={onOpenBookmarks}
              aria-label="书签"
              className="flex min-h-[38px] items-center gap-1 rounded-[6px] px-2.5 py-1 text-[13px] text-ink-secondary transition-colors hover:text-gold sm:min-h-0"
            >
              书签
              {bookmarkCount > 0 && (
                <span className="tnum rounded-[6px] bg-bg-raised px-1.5 text-[11px] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
                  {bookmarkCount}
                </span>
              )}
            </button>
          )}

          <div
            className="flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]"
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
                  "min-h-[38px] px-3 py-1 text-[13px] transition-colors sm:min-h-0",
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
    </div>
  );
}
