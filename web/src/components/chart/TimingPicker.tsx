"use client";

import { useEffect, useState } from "react";
import type { BirthInfo, EventCatalogItem, EventTiming } from "@/lib/types";
import { fetchTimingEvents, fetchEventTiming, ApiError } from "@/lib/api";

/**
 * 事项择吉:选所问之事(考试/签约/求财/婚嫁/求职/搬迁/出行),
 * 后端按对应宫逐层择时——利年 → 利月 → 利日(确定性,叠大限共振)。
 */
export function TimingPicker({ birth }: { birth: BirthInfo }) {
  const [events, setEvents] = useState<EventCatalogItem[]>([]);
  const [active, setActive] = useState<string>("");
  const [result, setResult] = useState<EventTiming | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchTimingEvents().then((r) => setEvents(r.events)).catch(() => {});
  }, []);

  // 切换命主或事项后清空旧结果。
  useEffect(() => {
    setResult(null);
    setActive("");
  }, [birth]);

  async function pick(key: string) {
    setActive(key);
    setLoading(true);
    setError("");
    try {
      const r = await fetchEventTiming(birth, key);
      setResult(r.timing);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "择吉计算失败");
      setResult(null);
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="rounded-[10px] bg-bg-raised p-5 shadow-[0_0_0_1px_var(--line)] md:p-6">
      <div className="flex items-baseline justify-between">
        <span className="text-[12px] tracking-[0.24em] text-gold">事项择吉</span>
        <span className="text-[11px] text-ink-faint">利年 · 利月 · 利日</span>
      </div>
      <p className="mt-1.5 text-[12px] text-ink-faint">选所问之事,按对应宫逐层择时(叠本命底色与大限共振,确定性)。</p>

      {(["auspicious", "avoid"] as const).map((kind) => {
        const items = events.filter((e) => e.kind === kind);
        if (items.length === 0) return null;
        return (
          <div key={kind} className="mt-3">
            <p className="mb-1.5 text-[11px] tracking-[0.16em] text-ink-faint">
              {kind === "auspicious" ? "择吉 · 宜择良时" : "避忌 · 宜避凶时"}
            </p>
            <div className="flex flex-wrap gap-1.5">
              {items.map((e) => (
                <button
                  key={e.key}
                  type="button"
                  aria-pressed={active === e.key}
                  onClick={() => pick(e.key)}
                  className={[
                    "min-h-[36px] rounded-full px-3.5 py-1.5 text-[13px] transition-colors",
                    active === e.key
                      ? "bg-[var(--gold-glow)] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                      : "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] hover:text-ink",
                  ].join(" ")}
                >
                  {e.label}
                </button>
              ))}
            </div>
          </div>
        );
      })}

      {loading && <p className="mt-4 text-[13px] text-ink-faint">推算中…</p>}
      {error && <p className="mt-4 text-[13px] text-danger">{error}</p>}

      {result && !loading && (
        <div className="mt-4 flex flex-col gap-3 border-t border-line pt-4">
          <p className="text-[13.5px] leading-[1.85] text-ink-secondary">{result.summary}</p>
          {result.baseNote && (
            <p className="rounded-[6px] bg-bg px-3 py-2 text-[12.5px] leading-[1.8] text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]">
              本命底色 · {result.baseQuality}：{result.baseNote}
            </p>
          )}

          {result.years.length > 0 && (
            <div>
              <p className="mb-1.5 text-[11px] tracking-[0.16em] text-ink-faint">
                {result.kind === "avoid" ? "忌年(未来十年宜避)" : "利年(未来十年)"}
              </p>
              <ul className="flex flex-col gap-1.5">
                {result.years.map((y) => (
                  <li key={y.year} className="rounded-[6px] bg-bg px-3 py-2 shadow-[inset_0_0_0_1px_var(--line)]">
                    <span className="font-display text-gold">{y.year}</span>
                    <span className="ml-1 text-[12px] text-ink-faint">{y.ganZhi}</span>
                    <span className="ml-2 text-[13px] text-ink-secondary">{y.note}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className="grid gap-3 sm:grid-cols-2">
            {result.bestMonth && (
              <div className="rounded-[6px] bg-bg px-3 py-2 shadow-[inset_0_0_0_1px_var(--line)]">
                <p className="text-[11px] tracking-[0.16em] text-ink-faint">{result.kind === "avoid" ? "忌月" : "利月"}</p>
                <p className="mt-0.5 text-[13.5px] text-ink-secondary">{result.bestMonth}</p>
              </div>
            )}
            {result.bestDays && result.bestDays.length > 0 && (
              <div className="rounded-[6px] bg-bg px-3 py-2 shadow-[inset_0_0_0_1px_var(--line)]">
                <p className="text-[11px] tracking-[0.16em] text-ink-faint">{result.kind === "avoid" ? "忌日" : "利日"}</p>
                <ul className="mt-0.5 text-[13.5px] text-ink-secondary">
                  {result.bestDays.map((d) => (
                    <li key={d}>{d}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          <p className="text-[12.5px] leading-[1.8] text-ink-faint">{result.advice}</p>
        </div>
      )}
    </section>
  );
}
