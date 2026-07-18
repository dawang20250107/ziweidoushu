"use client";

import { useState } from "react";
import type { Paragraph } from "@/lib/types";

type Layer = "" | "translation" | "niNote";

// 移动端触控目标加高,sm 起收回紧凑密度
const chipCls = "rounded-[2px] px-2 py-1.5 text-[12px] tracking-[0.08em] transition-colors sm:py-0.5";

/**
 * 单段正文:锚点 id="p-{id}" 供定位/金晕;有 translation/niNote 时段尾显示切换,
 * 点开在段下以次文本小字展开。三层文本互斥,再点收起。
 * 书签:段尾低调「书签」入口(悬停/触屏皆可点),已加书签段落左缘一道金色细标记。
 */
export function ReaderParagraph({
  p,
  bookmarked = false,
  onToggleBookmark,
}: {
  p: Paragraph;
  bookmarked?: boolean;
  onToggleBookmark?: (p: Paragraph) => void;
}) {
  const [layer, setLayer] = useState<Layer>("");
  const hasT = Boolean(p.translation);
  const hasN = Boolean(p.niNote);

  function toggle(k: Exclude<Layer, "">) {
    setLayer((cur) => (cur === k ? "" : k));
  }

  return (
    <div
      id={`p-${p.id}`}
      data-pid={p.id}
      className="group -mx-3 mb-6 scroll-mt-28 rounded-[6px] px-3 transition-shadow"
      // 金色细标记:已加书签段落左缘一道金线(与金晕背景动画互不干扰)
      style={bookmarked ? { boxShadow: "inset 2px 0 0 0 var(--gold)" } : undefined}
    >
      <p className="text-ink">{p.text}</p>

      <div className="mt-2 flex flex-wrap items-center gap-2 font-body">
        {hasT && (
          <button
            type="button"
            onClick={() => toggle("translation")}
            aria-pressed={layer === "translation"}
            className={`${chipCls} ${
              layer === "translation" ? "text-gold" : "text-ink-faint hover:text-ink-secondary"
            }`}
            style={layer === "translation" ? { background: "var(--gold-glow)" } : undefined}
          >
            白话
          </button>
        )}
        {hasN && (
          <button
            type="button"
            onClick={() => toggle("niNote")}
            aria-pressed={layer === "niNote"}
            className={`${chipCls} ${
              layer === "niNote" ? "text-gold" : "text-ink-faint hover:text-ink-secondary"
            }`}
            style={layer === "niNote" ? { background: "var(--gold-glow)" } : undefined}
          >
            倪注
          </button>
        )}

        {onToggleBookmark && (
          <button
            type="button"
            onClick={() => onToggleBookmark(p)}
            aria-pressed={bookmarked}
            aria-label={bookmarked ? "取消书签" : "加书签"}
            title={bookmarked ? "取消书签" : "加书签"}
            className={[
              chipCls,
              "ml-auto transition-opacity",
              bookmarked
                ? "text-gold opacity-100"
                : "text-ink-faint opacity-45 hover:text-gold group-hover:opacity-100 focus-visible:opacity-100",
            ].join(" ")}
          >
            {bookmarked ? "已收藏" : "书签"}
          </button>
        )}
      </div>

      {layer === "translation" && hasT && (
        <p className="mt-2 rounded-[6px] border-l-2 border-gold-dim bg-bg-raised px-3 py-2 text-[0.85em] leading-relaxed text-ink-secondary">
          {p.translation}
        </p>
      )}
      {layer === "niNote" && hasN && (
        <p className="mt-2 rounded-[6px] border-l-2 border-gold-dim bg-bg-raised px-3 py-2 text-[0.85em] leading-relaxed text-ink-secondary">
          <span className="mr-1.5 align-[0.08em] text-[11px] tracking-[0.08em] text-gold-dim">倪注</span>
          {p.niNote}
        </p>
      )}
    </div>
  );
}
