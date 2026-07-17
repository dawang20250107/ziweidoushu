"use client";

import { useCallback, useState } from "react";
import { castXiaoLiuRen, luckTone, DivinationError, type XiaoLiuRenResult } from "@/lib/divination";
import { toneBadgeClass, toneTextClass } from "./tone";

/**
 * 小六壬快占(独立轻区块):一键掐指起算,三步路径依次点亮 + 结果断语。
 * 免费、可反复;不占用 AI 解卦次数。本区块 CTA 用金描边(非辉光),
 * 辉光主 CTA 留给起卦与 AI 解卦。
 */
export function XiaoLiuRen() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<XiaoLiuRenResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [nonce, setNonce] = useState(0); // 每次起算换 key,重放点亮动效

  const run = useCallback(async () => {
    if (loading) return;
    setLoading(true);
    setError(null);
    try {
      const { result: r } = await castXiaoLiuRen();
      setResult(r);
      setNonce((n) => n + 1);
    } catch (e) {
      setError(e instanceof DivinationError ? e.message : "起算失败,请重试");
    } finally {
      setLoading(false);
    }
  }, [loading]);

  return (
    <section className="mt-14 rounded-[10px] bg-bg-raised px-5 py-8 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <style>{`
        @keyframes dvn-light {
          0%   { opacity: 0; transform: translateY(6px) scale(0.96); }
          100% { opacity: 1; transform: none; }
        }
        .dvn-step { animation: dvn-light 0.34s var(--ease-out) both; }
        @media (prefers-reduced-motion: reduce) {
          .dvn-step { animation: none; opacity: 1; transform: none; }
        }
      `}</style>

      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-[12px] font-medium tracking-[0.24em] text-gold">小六壬 · 急事速占</p>
          <h2 className="mt-2 font-display text-[25px] font-semibold text-ink">掐指一算</h2>
          <p className="mt-2 text-[14px] leading-relaxed text-ink-secondary">
            急事当下起算,掐指三步定吉凶缓急。免费,可反复。
          </p>
        </div>
        <button
          type="button"
          onClick={run}
          disabled={loading}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-bg-raised px-5 py-2.5 text-[15px] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-shadow hover:shadow-[inset_0_0_0_1px_var(--gold)] disabled:opacity-50"
        >
          {loading ? "掐指中…" : result ? "再算一次" : "急事速占"}
        </button>
      </div>

      {error && (
        <p className="mt-6 rounded-[6px] bg-bg px-4 py-3 text-[14px] text-danger shadow-[inset_0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {result && (
        <div className="mt-7 flex flex-col gap-6" key={nonce}>
          <p className="tnum text-[12px] tracking-[0.06em] text-ink-faint">起算 · {result.lunarText}</p>

          {/* 掐指三步:依次点亮 */}
          <div className="flex items-stretch gap-2">
            {result.path.map((pos, i) => {
              const tone = luckTone(pos.luck);
              const last = i === result.path.length - 1;
              return (
                <div key={i} className="flex flex-1 items-stretch gap-2">
                  <div
                    className={`dvn-step flex flex-1 flex-col items-center justify-center gap-1.5 rounded-[6px] bg-bg px-2 py-4 text-center shadow-[inset_0_0_0_1px_var(--line)] ${
                      last ? "shadow-[inset_0_0_0_1px_var(--gold-dim)]" : ""
                    }`}
                    style={{ animationDelay: `${i * 0.28}s` }}
                  >
                    <span className={`font-display text-[18px] font-semibold ${last ? "text-gold" : "text-ink"}`}>
                      {pos.name}
                    </span>
                    <span className={`text-[11px] ${toneTextClass(tone)}`}>{pos.luck}</span>
                  </div>
                  {!last && (
                    <span className="flex items-center text-ink-faint" aria-hidden>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
                        <path d="M9 6l6 6-6 6" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                    </span>
                  )}
                </div>
              );
            })}
          </div>

          {/* 结果断语 */}
          <div className="flex flex-col gap-3 rounded-[6px] bg-bg px-4 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
            <div className="flex items-center gap-3">
              <span className="font-display text-[22px] font-semibold text-ink">{result.result.name}</span>
              <span
                className={`inline-flex items-center rounded-[2px] px-2 py-0.5 text-[12px] tracking-[0.08em] ${toneBadgeClass(
                  luckTone(result.result.luck),
                )}`}
              >
                {result.result.luck}
              </span>
            </div>
            <p className="text-[14px] leading-relaxed text-ink-secondary">{result.result.meaning}</p>
          </div>
        </div>
      )}
    </section>
  );
}
