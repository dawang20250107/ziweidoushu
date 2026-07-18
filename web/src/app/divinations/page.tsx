"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import { listDivinations, deleteDivination, DivinationError, type DivinationRecord } from "@/lib/divination";
import { DIVINATION_KIND_LABEL as KIND_LABEL, formatDivinationTime } from "@/components/divination/format";

/** 卦档:历次问卦的存档列表(登录起卦自动留档,AI 解卦一并归档)。 */

export default function DivinationsPage() {
  const [signedIn, setSignedIn] = useState<boolean | null>(null);
  const [records, setRecords] = useState<DivinationRecord[] | null>(null);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [loadingMore, setLoadingMore] = useState(false);

  const PAGE = 100;

  const load = useCallback(() => {
    if (!currentUser()) {
      setSignedIn(false);
      setRecords(null);
      return;
    }
    setSignedIn(true);
    listDivinations(PAGE, 0)
      .then((r) => {
        setRecords(r.records);
        setTotal(r.total);
        setError(null);
      })
      .catch((e) => setError(e instanceof DivinationError ? e.message : "卦档加载失败,请刷新重试"));
  }, []);

  const loadMore = async () => {
    if (!records || loadingMore) return;
    setLoadingMore(true);
    try {
      const r = await listDivinations(PAGE, records.length);
      setRecords((prev) => [...(prev ?? []), ...r.records]);
      setTotal(r.total);
    } catch {
      setError("加载失败,请重试");
    } finally {
      setLoadingMore(false);
    }
  };

  useEffect(() => {
    load();
    window.addEventListener(AUTH_EVENT, load);
    return () => window.removeEventListener(AUTH_EVENT, load);
  }, [load]);

  const remove = async (id: string) => {
    if (deleting) return;
    setDeleting(id);
    try {
      await deleteDivination(id);
      setRecords((prev) => (prev ? prev.filter((r) => r.id !== id) : prev));
      setTotal((t) => Math.max(0, t - 1));
    } catch {
      setError("删除失败,请重试");
    } finally {
      setDeleting(null);
    }
  };

  return (
    <div className="mx-auto max-w-3xl px-5 py-14 md:py-20">
      <header>
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">问卦 · 卦档</p>
        <div className="mt-3 flex flex-wrap items-baseline justify-between gap-3">
          <h1 className="font-display text-[39px] font-semibold text-ink sm:text-[49px]">卦档</h1>
          {records != null && total > 0 && (
            <span className="tnum text-[12px] tracking-[0.08em] text-ink-faint">共 {total} 卦</span>
          )}
        </div>
        <p className="mt-3 text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">
          登录后起卦自动留档,AI 解卦一并归档;一事一占,回看应期。
        </p>
      </header>

      <div className="mt-10">
        {signedIn === false && (
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[inset_0_0_0_1px_var(--gold-dim)]">
            <p className="text-[14px] text-ink-secondary">登录后即可查看你的卦档。</p>
            <Link
              href="/login?next=/divinations"
              className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
            >
              去登录
            </Link>
          </div>
        )}

        {signedIn && error && <p className="text-[13px] text-danger">{error}</p>}

        {signedIn && records == null && !error && (
          <div className="flex flex-col gap-3" role="status" aria-label="卦档加载中">
            {[0, 1, 2].map((i) => (
              <div key={i} className="h-20 animate-pulse rounded-[10px] bg-bg-raised" aria-hidden />
            ))}
          </div>
        )}

        {signedIn && records != null && records.length === 0 && (
          <div className="flex flex-col items-start gap-4 rounded-[10px] bg-bg-raised px-6 py-10 shadow-[0_0_0_1px_var(--line)]">
            <p className="text-[15px] text-ink-secondary">卦档还空着。心念既定,去起一卦。</p>
            <Link
              href="/divination"
              className="glow-gold inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-5 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
            >
              去问卦
            </Link>
          </div>
        )}

        {signedIn && records != null && records.length > 0 && (
          <ul className="flex flex-col gap-3">
            {records.map((r) => (
              <li key={r.id}>
                <div className="group relative rounded-[10px] bg-bg-raised px-5 py-4 shadow-[0_0_0_1px_var(--line)] transition-shadow hover:shadow-[0_0_0_1px_var(--line-strong)]">
                  <Link href={`/divinations/${r.id}`} className="block">
                    <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1.5 pr-14">
                      <span className="rounded-[2px] px-1.5 py-0.5 text-[11px] leading-none tracking-[0.06em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
                        {KIND_LABEL[r.kind] ?? r.kind}
                      </span>
                      <span className="font-display text-[16px] font-medium text-ink">{r.summary}</span>
                      {r.hasReading && (
                        <span className="rounded-[2px] px-1.5 py-0.5 text-[11px] leading-none tracking-[0.06em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
                          已解卦
                        </span>
                      )}
                    </div>
                    <div className="mt-1.5 flex flex-wrap items-baseline gap-x-3 gap-y-1 pr-14">
                      {r.question && (
                        <span className="font-reading text-[14px] leading-relaxed text-ink-secondary">
                          {r.question}
                        </span>
                      )}
                      <span className="tnum text-[12px] tracking-[0.06em] text-ink-faint">
                        {formatDivinationTime(r.castAt)}
                      </span>
                    </div>
                  </Link>
                  <button
                    type="button"
                    onClick={() => remove(r.id)}
                    disabled={deleting === r.id}
                    aria-label={`删除卦档:${r.summary}`}
                    className="absolute right-3 top-3 rounded-[4px] px-2 py-1 text-[12px] text-ink-faint transition-colors hover:text-danger disabled:opacity-50"
                  >
                    {deleting === r.id ? "删除中" : "删除"}
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}

        {signedIn && records != null && records.length < total && (
          <div className="mt-5 flex justify-center">
            <button
              type="button"
              onClick={loadMore}
              disabled={loadingMore}
              className="min-h-[44px] rounded-[6px] bg-bg-raised px-5 py-2 text-[13px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink disabled:opacity-50"
            >
              {loadingMore ? "加载中…" : `加载更多(还有 ${total - records.length} 卦)`}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
