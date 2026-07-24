"use client";

import { useState } from "react";
import { castDaLiuRen, DivinationError, type DaLiuRenResult } from "@/lib/divination";
import { XiaoLiuRen } from "@/components/divination/XiaoLiuRen";

const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];
const KE_NAMES = ["一课", "二课", "三课", "四课"];
const CHUAN_NAMES = ["初传", "中传", "末传"];
const LEVEL_CLS: Record<string, string> = {
  good: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]",
  neutral: "text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)]",
  caution: "text-danger shadow-[inset_0_0_0_1px_var(--danger)]",
};
const LEVEL_LABEL: Record<string, string> = { good: "吉", neutral: "平", caution: "慎" };

/**
 * 六壬板块:大六壬起课(月将加时·天地盘/四课/三传/课体 + 确定性断语)
 * + 小六壬快占(倪师课堂掐指法)。自问卦页拆分独立。
 */
export default function LiuRenPage() {
  const [question, setQuestion] = useState("");
  const [casting, setCasting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [r, setR] = useState<DaLiuRenResult | null>(null);

  const cast = async () => {
    if (casting) return;
    setCasting(true);
    setError(null);
    try {
      const { result } = await castDaLiuRen({ question: question.trim() || undefined });
      setR(result);
    } catch (e) {
      setR(null);
      setError(e instanceof DivinationError ? e.message : "起课失败,请重试");
    } finally {
      setCasting(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl px-5 py-14 md:py-20">
      <header>
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">占卜 · 三式</p>
        <h1 className="font-display text-[39px] font-semibold text-ink sm:text-[49px]">六壬</h1>
        <p className="mt-3 text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">
          大六壬以月将加时布天地盘,发四课、立三传,课体定格局;小六壬掐指速断当下缓急。
        </p>
      </header>

      {/* ── 大六壬起课 ── */}
      <section className="mt-10 rounded-[10px] bg-bg-raised px-5 py-7 shadow-[0_0_0_1px_var(--line)] md:px-8 md:py-8">
        <label htmlFor="dlr-q" className="text-[13px] font-medium tracking-[0.06em] text-gold">
          心中默念所问之事(可留空断大势)
        </label>
        <textarea
          id="dlr-q"
          value={question}
          onChange={(e) => setQuestion(e.target.value.slice(0, 200))}
          rows={2}
          placeholder="例如:此事近期可有转机?"
          className="mt-3 w-full resize-none rounded-[6px] bg-bg px-4 py-3 text-[15px] leading-relaxed text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow placeholder:text-ink-faint focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
        />
        <div className="mt-5 flex flex-col items-start gap-2.5">
          <button
            type="button"
            onClick={cast}
            disabled={casting}
            className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-8 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-45 sm:w-auto"
          >
            {casting ? "起课中…" : "以此时起课"}
          </button>
          {error && <p className="text-[13px] text-danger">{error}</p>}
        </div>
      </section>

      {r && (
        <div className="page-enter mt-10">
          {/* 课骨:日干支/占时/月将/课体 */}
          <div className="tnum flex flex-wrap items-center gap-x-3 gap-y-1 text-[13px] text-ink-secondary">
            <span>
              {r.dayStem}
              {r.dayBranch}日
            </span>
            <span>· {r.hourBranch}时占</span>
            <span>· 月将{r.monthGen}</span>
            <span className="rounded-[3px] px-1.5 py-0.5 text-[12px] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
              {r.keType}课
            </span>
          </div>

          {/* 天地盘 */}
          <div className="mt-6 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
            <p className="text-[12px] font-medium tracking-[0.24em] text-gold">天地盘</p>
            <div className="mt-4 grid grid-cols-6 gap-1.5 sm:grid-cols-12">
              {BRANCHES.map((b, i) => (
                <div
                  key={b}
                  className="flex flex-col items-center gap-1 rounded-[6px] bg-bg px-1 py-2.5 shadow-[inset_0_0_0_1px_var(--line)]"
                >
                  <span className="font-display text-[15px] text-gold">{r.tianPan[i]}</span>
                  <span className="text-[11px] text-ink-faint">{b}</span>
                </div>
              ))}
            </div>
            <p className="mt-2 text-[11px] text-ink-faint">上行天盘之神,下行地盘定位(月将加时顺布)。</p>
          </div>

          {/* 四课 / 三传 */}
          <div className="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)]">
              <p className="text-[12px] font-medium tracking-[0.24em] text-gold">四课</p>
              <div className="mt-4 grid grid-cols-4 gap-2">
                {r.ke.map((k, i) => (
                  <div
                    key={i}
                    className="flex flex-col items-center gap-1.5 rounded-[6px] bg-bg px-2 py-3 shadow-[inset_0_0_0_1px_var(--line)]"
                  >
                    <span className="font-display text-[19px] text-ink">{k.upper}</span>
                    <span className="h-px w-6 bg-line" aria-hidden />
                    <span className="text-[14px] text-ink-secondary">{k.lower}</span>
                    <span className="text-[10px] text-ink-faint">{KE_NAMES[i]}</span>
                  </div>
                ))}
              </div>
            </div>
            <div className="rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)]">
              <p className="text-[12px] font-medium tracking-[0.24em] text-gold">三传</p>
              <div className="mt-4 flex items-center justify-around">
                {r.chuan.map((c, i) => (
                  <div key={i} className="flex flex-col items-center gap-1.5">
                    <span className="font-display text-[25px] font-semibold text-ink">{c}</span>
                    <span className="text-[11px] text-ink-faint">{CHUAN_NAMES[i]}</span>
                  </div>
                ))}
              </div>
              {r.judgment && (
                <ul className="mt-4 flex flex-col gap-1.5 border-t border-line pt-3">
                  {r.judgment.sanChuan.map((s, i) => (
                    <li key={i} className="text-[13px] leading-relaxed text-ink-secondary">
                      {s}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>

          {/* 断语 */}
          {r.judgment && (
            <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-[12px] font-medium tracking-[0.24em] text-gold">断语 · 课体三传</p>
                <span
                  className={`rounded-[3px] px-1.5 py-0.5 text-[11px] leading-none ${LEVEL_CLS[r.judgment.level] ?? LEVEL_CLS.neutral}`}
                >
                  {LEVEL_LABEL[r.judgment.level] ?? "平"}
                </span>
              </div>
              <p className="mt-4 font-reading text-[16px] leading-[1.9] text-ink">{r.judgment.conclusion}</p>
              <p className="mt-3 text-[13px] leading-relaxed text-ink-secondary">{r.judgment.keTypeText}</p>
              <ul className="mt-3 flex flex-col gap-2 border-t border-line pt-3">
                {r.judgment.points.map((p, i) => (
                  <li key={i} className="flex gap-2 text-[14px] leading-relaxed text-ink-secondary">
                    <span aria-hidden className="mt-[9px] h-[3px] w-[3px] shrink-0 rounded-full bg-gold-dim" />
                    {p}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* ── 小六壬快占 ── */}
      <XiaoLiuRen />
    </div>
  );
}
