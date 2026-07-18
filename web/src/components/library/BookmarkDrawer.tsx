"use client";

import { useEffect } from "react";
import Link from "next/link";
import type { Bookmark } from "@/lib/reading";

/**
 * 书签抽屉:右侧滑入,列出本书书签(第 N 章 + 摘录),点击跳转、可删除。
 * 未登录时不做本地书签,改为引导登录。
 */
export function BookmarkDrawer({
  open,
  onClose,
  slug,
  signedIn,
  bookmarks,
  loading,
  onDelete,
  loginHref,
}: {
  open: boolean;
  onClose: () => void;
  slug: string;
  signedIn: boolean;
  bookmarks: Bookmark[];
  loading: boolean;
  onDelete: (id: string) => void;
  loginHref: string;
}) {
  // Esc 关闭 + 打开时锁滚动
  useEffect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = prev;
    };
  }, [open, onClose]);

  return (
    <div
      // overflow-hidden:收起态面板 translate-x-full 移出视口,不裁剪会把页面横向撑宽
      className={`fixed inset-0 z-50 overflow-hidden ${open ? "" : "pointer-events-none"}`}
      aria-hidden={!open}
    >
      {/* 遮罩 */}
      <div
        onClick={onClose}
        className={`absolute inset-0 bg-bg-overlay/60 backdrop-blur-[1px] transition-opacity duration-200 ${
          open ? "opacity-100" : "opacity-0"
        }`}
      />

      {/* 面板 */}
      <aside
        role="dialog"
        aria-modal="true"
        aria-label="本书书签"
        className={`absolute right-0 top-0 flex h-full w-[min(88vw,22rem)] flex-col border-l border-line bg-bg shadow-[var(--elevation-2)] transition-transform duration-200 ease-out ${
          open ? "translate-x-0" : "translate-x-full"
        }`}
      >
        <header className="flex items-center justify-between border-b border-line px-5 py-3.5">
          <div>
            <p className="text-[11px] tracking-[0.24em] text-gold">书签</p>
            <h2 className="mt-0.5 font-display text-[17px] font-semibold text-ink">本书收藏</h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="关闭"
            className="min-h-[38px] rounded-[6px] px-2.5 py-1 text-[13px] text-ink-faint transition-colors hover:text-gold sm:min-h-0"
          >
            关闭
          </button>
        </header>

        <div className="flex-1 overflow-y-auto px-5 py-4">
          {!signedIn ? (
            <div className="rounded-[10px] bg-bg-raised px-5 py-10 text-center shadow-[0_0_0_1px_var(--line)]">
              <p className="font-display text-[16px] font-semibold text-ink">登录后使用书签</p>
              <p className="mt-2 text-[13px] leading-relaxed text-ink-secondary">
                书签随账号跨端同步,任意设备接着读、随手收藏。
              </p>
              <Link
                href={loginHref}
                className="mt-5 inline-flex min-h-[40px] items-center rounded-[6px] bg-gold px-5 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
              >
                去登录
              </Link>
            </div>
          ) : loading ? (
            <p className="py-10 text-center text-[13px] text-ink-faint">载入中…</p>
          ) : bookmarks.length === 0 ? (
            <div className="rounded-[10px] bg-bg-raised px-5 py-10 text-center shadow-[0_0_0_1px_var(--line)]">
              <p className="font-display text-[16px] font-semibold text-ink">还没有书签</p>
              <p className="mt-2 text-[13px] leading-relaxed text-ink-secondary">
                阅读时点段尾「书签」即可收藏此处,稍后从这里一键回到。
              </p>
            </div>
          ) : (
            <ul className="flex flex-col gap-2">
              {bookmarks.map((b) => (
                <li
                  key={b.id}
                  className="group flex items-start gap-2 rounded-[6px] bg-bg-raised p-3 shadow-[0_0_0_1px_var(--line)] transition-shadow hover:shadow-[0_0_0_1px_var(--gold-dim)]"
                >
                  <Link
                    href={`/library/${slug}/${b.chapterIdx}${b.paragraphId ? `#p-${b.paragraphId}` : ""}`}
                    onClick={onClose}
                    className="min-w-0 flex-1"
                  >
                    <span className="tnum text-[11px] tracking-[0.12em] text-gold">
                      第 {b.chapterIdx + 1} 章
                    </span>
                    <span className="mt-1 line-clamp-3 font-reading text-[14px] leading-relaxed text-ink-secondary">
                      {b.excerpt || "(无摘录)"}
                    </span>
                  </Link>
                  <button
                    type="button"
                    onClick={() => onDelete(b.id)}
                    aria-label="删除书签"
                    className="shrink-0 rounded-[6px] px-2 py-1 text-[12px] text-ink-faint opacity-60 transition-colors hover:text-danger group-hover:opacity-100"
                  >
                    删除
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </aside>
    </div>
  );
}
