"use client";

import { use, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { fetchChapter, ApiError } from "@/lib/api";
import type { Chapter } from "@/lib/types";
import { ReaderToolbar } from "@/components/library/ReaderToolbar";
import { ReaderParagraph } from "@/components/library/ReaderParagraph";
import { SkeletonLines } from "@/components/library/Skeleton";
import { useFontSize, saveProgress } from "@/components/library/prefs";

interface ChapterData {
  book: { title: string; slug: string };
  index: number;
  total: number;
  chapter: Chapter;
}

/** 命中锚点段落时,以 --gold-glow 背景 2 秒淡出的短暂金晕。 */
function glow(el: HTMLElement) {
  if (typeof el.animate !== "function") return;
  const color = getComputedStyle(document.documentElement).getPropertyValue("--gold-glow").trim();
  if (!color) return;
  el.animate(
    [{ backgroundColor: color }, { backgroundColor: "transparent" }],
    { duration: 2000, easing: "cubic-bezier(0.22, 1, 0.36, 1)" },
  );
}

/** 阅读器(核心):三档字号、锚点定位金晕、三层文本、章节导航、进度记忆。 */
export default function ReaderPage({
  params,
}: {
  params: Promise<{ slug: string; chapter: string }>;
}) {
  const { slug, chapter } = use(params);
  const idx = Number(chapter);
  const router = useRouter();
  const [size, setSize] = useFontSize();
  const [data, setData] = useState<ChapterData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // 拉取章节
  useEffect(() => {
    if (!Number.isInteger(idx) || idx < 0) {
      setError("章节不存在");
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError("");
    setData(null);
    fetchChapter(slug, idx)
      .then((r) => {
        if (!cancelled) setData(r);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "章节加载失败");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [slug, idx]);

  const total = data?.total ?? 0;
  const hasPrev = idx > 0;
  const hasNext = data != null && idx < total - 1;
  const prevHref = `/library/${slug}/${idx - 1}`;
  const nextHref = `/library/${slug}/${idx + 1}`;

  // 键盘 ←/→ 翻章(忽略输入态与修饰键)
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.metaKey || e.ctrlKey || e.altKey || e.shiftKey) return;
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (e.key === "ArrowLeft" && hasPrev) {
        e.preventDefault();
        router.push(prevHref);
      } else if (e.key === "ArrowRight" && hasNext) {
        e.preventDefault();
        router.push(nextHref);
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [hasPrev, hasNext, prevHref, nextHref, router]);

  // 锚点定位金晕 + 进度写入(首个可见段)
  useEffect(() => {
    if (!data) return;
    const els = Array.from(document.querySelectorAll<HTMLElement>("[data-pid]"));

    // 进入章节即记录进度
    saveProgress(slug, { chapterIdx: idx, paragraphId: els[0]?.dataset.pid ?? "" });

    // URL 带 #p-xxx:滚动定位并金晕高亮;否则回到顶部
    const hash = window.location.hash;
    if (hash.startsWith("#p-")) {
      const el = document.getElementById(hash.slice(1));
      if (el) {
        requestAnimationFrame(() => {
          el.scrollIntoView({ behavior: "smooth", block: "start" });
          glow(el);
        });
      }
    } else {
      window.scrollTo({ top: 0 });
    }

    // 首个可见段 → 进度(去抖)
    const visible = new Set<string>();
    let timer: number | undefined;
    const persist = () => {
      const first = els.find((el) => el.dataset.pid && visible.has(el.dataset.pid)) ?? els[0];
      if (first?.dataset.pid != null) {
        saveProgress(slug, { chapterIdx: idx, paragraphId: first.dataset.pid });
      }
    };
    const io = new IntersectionObserver(
      (entries) => {
        for (const en of entries) {
          const pid = (en.target as HTMLElement).dataset.pid;
          if (!pid) continue;
          if (en.isIntersecting) visible.add(pid);
          else visible.delete(pid);
        }
        window.clearTimeout(timer);
        timer = window.setTimeout(persist, 300);
      },
      { rootMargin: "-140px 0px -55% 0px" },
    );
    els.forEach((el) => io.observe(el));
    return () => {
      io.disconnect();
      window.clearTimeout(timer);
    };
  }, [data, slug, idx]);

  return (
    <div>
      <ReaderToolbar
        slug={slug}
        bookTitle={data?.book.title ?? ""}
        chapterTitle={data?.chapter.title ?? ""}
        size={size}
        onSize={setSize}
      />

      <article
        className="mx-auto max-w-[68ch] px-5 py-8 md:py-12"
        style={{ fontSize: `${size}px`, lineHeight: 1.9 }}
      >
        {error && (
          <p className="rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
            {error}
          </p>
        )}

        {loading && !data && <SkeletonLines lines={10} />}

        {data && (
          <>
            <header className="mb-8 text-center">
              <p className="tnum text-[12px] tracking-[0.24em] text-gold">
                第 {idx + 1} / {total} 章
              </p>
              <h1 className="mt-2 font-display text-2xl font-semibold text-ink">
                {data.chapter.title}
              </h1>
              {data.chapter.subtitle && (
                <p className="mt-1 font-reading text-[14px] text-ink-secondary">
                  {data.chapter.subtitle}
                </p>
              )}
            </header>

            <div className="font-reading">
              {data.chapter.paragraphs.map((p) => (
                <ReaderParagraph key={p.id} p={p} />
              ))}
            </div>

            <nav
              className="mt-12 flex items-center justify-between gap-3 border-t border-line pt-6 font-body text-[14px]"
              aria-label="章节导航"
            >
              {hasPrev ? (
                <Link
                  href={prevHref}
                  className="rounded-[6px] px-3 py-2 text-ink-secondary transition-colors hover:text-gold"
                >
                  ← 上一章
                </Link>
              ) : (
                <span aria-hidden />
              )}
              <Link
                href={`/library/${slug}`}
                className="rounded-[6px] px-3 py-2 text-ink-faint transition-colors hover:text-gold"
              >
                目录
              </Link>
              {hasNext ? (
                <Link
                  href={nextHref}
                  className="rounded-[6px] px-3 py-2 text-ink-secondary transition-colors hover:text-gold"
                >
                  下一章 →
                </Link>
              ) : (
                <span aria-hidden />
              )}
            </nav>
          </>
        )}
      </article>
    </div>
  );
}
