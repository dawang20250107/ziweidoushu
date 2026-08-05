"use client";

import { useRef, useState } from "react";
import { castDaLiuRen, divineDaLiuRenAI, DivinationError, type DaLiuRenResult } from "@/lib/divination";
import { currentUser } from "@/lib/auth";
import { ReportText } from "@/components/profiles/ReportText";
import { XiaoLiuRen } from "@/components/divination/XiaoLiuRen";
import { LiurenPan } from "@/components/divination/LiurenPan";
import { LiurenCast } from "@/components/divination/LiurenCast";
import { prefersReducedMotion } from "@/components/divination/useReducedMotion";

const KE_NAMES = ["一课", "二课", "三课", "四课"];
// 起课仪式总时长(月将加时→天将布位→课成)
const CAST_ANIM_MS = 3200;
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
  const [shiMode, setShiMode] = useState<"bao" | "zheng">("bao");
  const [baoInput, setBaoInput] = useState("");
  const [casting, setCasting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [r, setR] = useState<DaLiuRenResult | null>(null);
  const [castAt, setCastAt] = useState<number | null>(null);
  const [recordId, setRecordId] = useState<string | null>(null);
  const [aiLoading, setAiLoading] = useState(false);
  const [reading, setReading] = useState<string | null>(null);
  const [aiError, setAiError] = useState<string | null>(null);
  // 起课问辞快照(解读须扣起课之问,非点击解课时输入框的最新文本)
  const [castQuestion, setCastQuestion] = useState("");
  // 起课序号:解课途中若重新起课,迟到的解读不得错挂到新课上
  const castSeq = useRef(0);

  const divine = async () => {
    if (aiLoading || castAt == null) return;
    const seq = castSeq.current;
    setAiLoading(true);
    setAiError(null);
    setReading(null);
    try {
      const { reading: rd } = await divineDaLiuRenAI({
        castAt,
        question: castQuestion || "断大势",
        recordId: recordId ?? undefined,
        // 活时课须以同一报数重推同一课(服务端代摇之数已随课回传)
        baoShu: r?.baoShu || undefined,
      });
      if (seq === castSeq.current) setReading(rd.text);
    } catch (e) {
      if (seq !== castSeq.current) return;
      if (e instanceof DivinationError && e.status === 401) setAiError("登录后即可 AI 深度解课。");
      else if (e instanceof DivinationError && (e.status === 402 || e.code === "no_credits")) setAiError("解卦次数不足,请先购买次卡。");
      else if (e instanceof DivinationError && (e.status === 503 || e.code === "ai_unavailable")) setAiError("AI 服务暂不可用,本次未扣次数。");
      else setAiError(e instanceof DivinationError ? e.message : "解课失败,请重试");
    } finally {
      setAiLoading(false);
    }
  };

  const cast = async () => {
    if (casting) return;
    // 活时:报数定占时(留空由服务端代摇);正时:同一时辰之课人人相同,古以年命分断
    let bao: number | undefined;
    if (shiMode === "bao") {
      const t = baoInput.trim();
      if (t === "") {
        bao = 0; // 代摇
      } else {
        const n = Number(t);
        if (!Number.isInteger(n) || n < 1 || n > 99) {
          setError("报数请输入 1-99 的整数,或留空由天心代摇");
          return;
        }
        bao = n;
      }
    }
    setCasting(true);
    setError(null);
    const startedAt = Date.now();
    try {
      const resp = await castDaLiuRen({ question: question.trim() || undefined, baoShu: bao });
      // 仪式演满再揭课(reduced-motion 直出)
      const wait = (prefersReducedMotion() ? 0 : CAST_ANIM_MS) - (Date.now() - startedAt);
      if (wait > 0) await new Promise((res) => setTimeout(res, wait));
      castSeq.current += 1;
      setR(resp.result);
      setCastAt(resp.castAt);
      setRecordId(resp.recordId ?? null);
      setCastQuestion(question.trim());
      setReading(null);
      setAiError(null);
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
        {/* 占时方式:活时报数(众人同刻各课)/正时(同一时辰课同,古以年命分断) */}
        <p className="mt-5 text-[13px] font-medium tracking-[0.06em] text-gold">占时</p>
        <div className="mt-2.5 flex flex-wrap items-stretch gap-2">
          <button
            type="button"
            aria-pressed={shiMode === "bao"}
            onClick={() => setShiMode("bao")}
            className={[
              "flex min-h-[48px] flex-col justify-center rounded-[6px] px-4 py-2 text-left transition-shadow",
              shiMode === "bao"
                ? "bg-[var(--gold-glow)] shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                : "bg-bg shadow-[inset_0_0_0_1px_var(--line)]",
            ].join(" ")}
          >
            <span className={`text-[14px] ${shiMode === "bao" ? "font-medium text-gold" : "text-ink"}`}>报数活时</span>
            <span className="text-[11px] text-ink-faint">心动报一数定占时 · 主推</span>
          </button>
          <button
            type="button"
            aria-pressed={shiMode === "zheng"}
            onClick={() => setShiMode("zheng")}
            className={[
              "flex min-h-[48px] flex-col justify-center rounded-[6px] px-4 py-2 text-left transition-shadow",
              shiMode === "zheng"
                ? "bg-[var(--gold-glow)] shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                : "bg-bg shadow-[inset_0_0_0_1px_var(--line)]",
            ].join(" ")}
          >
            <span className={`text-[14px] ${shiMode === "zheng" ? "font-medium text-gold" : "text-ink"}`}>正时起课</span>
            <span className="text-[11px] text-ink-faint">以当下时辰 · 同辰课同</span>
          </button>
          {shiMode === "bao" && (
            <input
              type="number"
              inputMode="numeric"
              min={1}
              max={99}
              value={baoInput}
              onChange={(e) => setBaoInput(e.target.value.slice(0, 2))}
              placeholder="报数,留空代摇"
              aria-label="活时报数(1-99,留空由服务端代摇)"
              className="tnum min-h-[48px] w-36 rounded-[6px] bg-bg px-3.5 text-[14px] text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow placeholder:text-ink-faint focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
            />
          )}
        </div>
        <p className="mt-2 text-[11px] leading-relaxed text-ink-faint">
          {shiMode === "bao"
            ? "自子顺数至所报之数定占时,众人同刻各得其课;留空则由天心代摇一数。"
            : "正时之课同一时辰人人相同,古以问者年命分断;欲各得其课请用报数活时。"}
        </p>

        <div className="mt-5 flex flex-col items-start gap-2.5">
          <button
            type="button"
            onClick={cast}
            disabled={casting}
            className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-8 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-45 sm:w-auto"
          >
            {casting ? "起课中…" : shiMode === "bao" ? "报数起课" : "以此时起课"}
          </button>
          {error && <p className="text-[13px] text-danger">{error}</p>}
        </div>
      </section>

      {/* 起课仪式 */}
      {casting && <LiurenCast />}

      {r && !casting && (
        <div className="page-enter mt-10">
          {/* 式盘:天地盘/天将/三传/课骨一体呈现;各块与全站同套滚动聚焦节奏 */}
          <div className="reveal">
            <LiurenPan result={r} />
          </div>

          {/* 四课 / 三传 */}
          <div className="reveal mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
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
                    <span className="font-display text-[25px] font-semibold text-ink">
                      {r.chuanDunGan?.[i] ? (
                        <span className="mr-0.5 align-middle text-[14px] font-normal text-ink-faint">
                          {r.chuanDunGan[i]}
                        </span>
                      ) : (
                        r.xunKong?.length === 2 && (
                          <span className="mr-0.5 align-middle text-[11px] font-normal text-danger">空</span>
                        )
                      )}
                      {c}
                    </span>
                    {r.chuanJiang?.[i] && (
                      <span className="rounded-[3px] px-1.5 py-0.5 text-[10px] leading-none text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
                        {r.chuanJiang[i]}
                      </span>
                    )}
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
            <div className="reveal mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
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

          {/* ── AI 深度解课 ── */}
          <div className="reveal mt-8">
            {reading != null && !aiLoading ? (
              <article className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)] md:px-10 md:py-10">
                <p className="mb-5 text-[12px] font-medium tracking-[0.24em] text-gold">AI 深度解课</p>
                <ReportText text={reading} />
                <p className="mt-6 border-t border-line pt-4 text-[11px] leading-relaxed text-ink-faint">
                  占卜为传统文化参考,不构成决策建议。
                </p>
              </article>
            ) : aiLoading ? (
              <div className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)]" role="status">
                <div className="flex flex-col gap-3">
                  {[94, 100, 86, 72].map((w, i) => (
                    <div key={i} className="h-4 animate-pulse rounded-[2px] bg-line" style={{ width: `${w}%` }} aria-hidden />
                  ))}
                </div>
                <p className="mt-5 text-[12px] text-ink-faint">AI 正在依课象逐层解读,通常需 20 秒以上…</p>
              </div>
            ) : (
              <div className="flex flex-col items-start gap-2.5">
                <button
                  type="button"
                  onClick={divine}
                  className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-7 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright sm:w-auto"
                >
                  {currentUser() ? "AI 深度解课(消耗 1 次)" : "登录后 AI 深度解课"}
                </button>
                <p className="text-[12px] text-ink-faint">依课体三传天将与断语骨架逐层解读。</p>
                {aiError && <p className="text-[13px] text-danger">{aiError}</p>}
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── 小六壬快占 ── */}
      <div className="reveal">
        <XiaoLiuRen />
      </div>
    </div>
  );
}
