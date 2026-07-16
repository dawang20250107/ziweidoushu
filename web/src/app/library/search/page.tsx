"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { searchClassics, ApiError } from "@/lib/api";
import type { SearchHit } from "@/lib/types";
import { SearchBox } from "@/components/library/SearchBox";
import { SkeletonLines } from "@/components/library/Skeleton";

function SearchInner() {
  const router = useRouter();
  const params = useSearchParams();
  const [q, setQ] = useState(params.get("q") ?? "");
  const [hits, setHits] = useState<SearchHit[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [searched, setSearched] = useState(false);

  // 输入即时检索(去抖 250ms),并把查询回写到 URL(不堆历史)。
  useEffect(() => {
    const query = q.trim();
    if (!query) {
      setHits([]);
      setCount(0);
      setError("");
      setSearched(false);
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    const handle = window.setTimeout(() => {
      searchClassics(query)
        .then((r) => {
          if (!active) return;
          setHits(r.hits);
          setCount(r.count);
          setError("");
          setSearched(true);
        })
        .catch((e) => {
          if (!active) return;
          setError(e instanceof ApiError ? e.message : "检索失败");
          setHits([]);
          setSearched(true);
        })
        .finally(() => {
          if (active) setLoading(false);
        });
      router.replace(`/library/search?q=${encodeURIComponent(query)}`);
    }, 250);
    return () => {
      active = false;
      window.clearTimeout(handle);
    };
  }, [q, router]);

  return (
    <div className="mx-auto max-w-3xl px-5 py-8 md:py-12">
      <nav className="mb-4 flex items-center gap-1.5 text-[13px] text-ink-faint" aria-label="面包屑">
        <Link href="/library" className="transition-colors hover:text-gold">
          书架
        </Link>
        <span aria-hidden>/</span>
        <span className="text-ink-secondary">检索</span>
      </nav>

      <SearchBox value={q} onChange={setQ} autoFocus placeholder="检索全文…书名、星曜、格局" />

      <p className="mt-3 text-[13px] text-ink-faint" aria-live="polite">
        {loading
          ? "检索中…"
          : q.trim() && searched && !error
            ? `找到 ${count} 条`
            : "输入关键词开始检索"}
      </p>

      {error && (
        <p className="mt-4 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {loading && <SkeletonLines lines={6} className="mt-5" />}

      {!loading && searched && !error && hits.length === 0 && (
        <p className="py-16 text-center text-ink-faint">未找到匹配的段落。</p>
      )}

      {!loading && hits.length > 0 && (
        <ul className="mt-4 divide-y divide-line">
          {hits.map((h, i) => (
            <li key={`${h.bookSlug}-${h.paragraphId}-${i}`}>
              <Link
                href={`/library/${h.bookSlug}/${h.chapterIdx}#p-${h.paragraphId}`}
                className="block rounded-[6px] px-2 py-4 transition-colors hover:bg-bg-raised"
              >
                <div className="flex items-baseline gap-2 text-[12px] text-ink-faint">
                  <span className="font-display text-gold-dim">{h.bookTitle}</span>
                  <span aria-hidden>·</span>
                  <span className="min-w-0 truncate">{h.chapterTitle}</span>
                </div>
                <p
                  className="mt-1.5 font-reading text-[15px] leading-relaxed text-ink-secondary"
                  dangerouslySetInnerHTML={{ __html: h.snippet }}
                />
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

/** 检索页:?q= 即时全文检索,命中 snippet 含 <mark>,点击跳阅读器并定位段落。 */
export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-3xl px-5 py-12">
          <SkeletonLines lines={5} />
        </div>
      }
    >
      <SearchInner />
    </Suspense>
  );
}
