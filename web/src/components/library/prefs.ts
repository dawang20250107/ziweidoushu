"use client";

/** 古籍阅读器的本地偏好与进度存取(localStorage)。 */
import { useEffect, useState } from "react";

// ── 字号(三档,记忆)──────────────────────────────────
export const FONT_SIZES = [18, 20, 23] as const;
export type FontSize = (typeof FONT_SIZES)[number];
const FONT_KEY = "ziwei-font-size";

/** 字号 hook:默认 20px,挂载后读取本地记忆,更新即持久化。 */
export function useFontSize(): [FontSize, (s: FontSize) => void] {
  const [size, setSize] = useState<FontSize>(20);

  useEffect(() => {
    try {
      const v = Number(localStorage.getItem(FONT_KEY));
      if ((FONT_SIZES as readonly number[]).includes(v)) setSize(v as FontSize);
    } catch {
      /* 隐私模式等 storage 不可用时静默降级 */
    }
  }, []);

  function update(s: FontSize) {
    setSize(s);
    try {
      localStorage.setItem(FONT_KEY, String(s));
    } catch {
      /* ignore */
    }
  }

  return [size, update];
}

// ── 阅读进度 ──────────────────────────────────────────
export interface ReadingProgress {
  chapterIdx: number;
  paragraphId: string;
}

function progressKey(slug: string): string {
  return `ziwei-reading-${slug}`;
}

export function loadProgress(slug: string): ReadingProgress | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(progressKey(slug));
    if (!raw) return null;
    const p = JSON.parse(raw) as Partial<ReadingProgress>;
    if (typeof p.chapterIdx === "number" && typeof p.paragraphId === "string") {
      return { chapterIdx: p.chapterIdx, paragraphId: p.paragraphId };
    }
  } catch {
    /* ignore */
  }
  return null;
}

export function saveProgress(slug: string, p: ReadingProgress): void {
  try {
    localStorage.setItem(progressKey(slug), JSON.stringify(p));
  } catch {
    /* ignore */
  }
}
