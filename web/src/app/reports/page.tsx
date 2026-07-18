"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import {
  listProfiles,
  fetchEntitlements,
  generateReport,
  readRecentBirth,
  formatBirthSummary,
  ProfileError,
  type Profile,
  type BirthRequest,
} from "@/lib/profiles";
import { ReportText } from "@/components/profiles/ReportText";

const REPORT_TOPICS: { key: string; label: string }[] = [
  { key: "overview", label: "命盘总览" },
  { key: "career", label: "事业官禄" },
  { key: "wealth", label: "财帛理财" },
  { key: "marriage", label: "婚姻情感" },
  { key: "health", label: "健康养生" },
  { key: "children", label: "子女缘分" },
];

const RECENT = "recent"; // birthSource 的特殊取值:最近排盘

type ErrKind = "no_credits" | "ai_unavailable" | "network";
interface GenError {
  kind: ErrKind;
  message: string;
}

/** 深度报告:选生辰来源 + 主题,消耗 1 次深度报告生成长文解读。 */
export default function ReportsPage() {
  const [signedIn, setSignedIn] = useState<boolean | null>(null);
  const [initLoading, setInitLoading] = useState(true);

  const [recentBirth, setRecentBirth] = useState<BirthRequest | null>(null);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [credits, setCredits] = useState<number | null>(null); // null=未知(权益获取失败)

  const [source, setSource] = useState<string>(""); // RECENT 或 profile.id
  const [topic, setTopic] = useState<string>("overview");

  const [generating, setGenerating] = useState(false);
  const [report, setReport] = useState<string | null>(null);
  const [degraded, setDegraded] = useState(false);
  const [error, setError] = useState<GenError | null>(null);

  const loadInit = useCallback(async () => {
    setInitLoading(true);
    const recent = readRecentBirth();
    setRecentBirth(recent);

    const [profRes, entRes] = await Promise.allSettled([listProfiles(), fetchEntitlements()]);

    let list: Profile[] = [];
    if (profRes.status === "fulfilled") {
      list = profRes.value.profiles;
      setProfiles(list);
    }
    if (entRes.status === "fulfilled") {
      setCredits(entRes.value.credits?.deep_report ?? 0);
    } else {
      setCredits(null);
    }

    // 默认来源:优先最近排盘,否则默认档案 / 第一份档案
    if (recent) {
      setSource(RECENT);
    } else if (list.length > 0) {
      setSource((list.find((p) => p.isDefault) ?? list[0]).id);
    } else {
      setSource("");
    }
    setInitLoading(false);
  }, []);

  useEffect(() => {
    const sync = () => {
      const ok = !!currentUser();
      setSignedIn(ok);
      if (ok) void loadInit();
      else setInitLoading(false);
    };
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, [loadInit]);

  // 解析当前选中的生辰
  const resolvedBirth = useMemo<BirthRequest | null>(() => {
    if (source === RECENT) return recentBirth;
    return profiles.find((p) => p.id === source)?.birthInput ?? null;
  }, [source, recentBirth, profiles]);

  const noSource = !recentBirth && profiles.length === 0;
  const outOfCredits = credits === 0;

  const generate = useCallback(async () => {
    if (!resolvedBirth || !topic || generating) return;
    setGenerating(true);
    setError(null);
    setReport(null);
    try {
      const r = await generateReport({ ...resolvedBirth, topic });
      setReport(r.report.text);
      setDegraded(r.report.degraded);
      setCredits(r.remainingCredits);
    } catch (e) {
      if (e instanceof ProfileError) {
        if (e.status === 402 || e.code === "no_credits") {
          setError({ kind: "no_credits", message: e.message });
          setCredits(0);
        } else if (e.status === 503 || e.code === "ai_unavailable") {
          setError({ kind: "ai_unavailable", message: e.message });
        } else {
          setError({ kind: "network", message: e.message });
        }
      } else {
        setError({ kind: "network", message: "网络异常,请重试" });
      }
    } finally {
      setGenerating(false);
    }
  }, [resolvedBirth, topic, generating]);

  // ── 未登录 ──
  if (signedIn === false) {
    return (
      <div className="mx-auto max-w-md px-4 py-24 md:py-32">
        <div className="rounded-[10px] bg-bg-raised px-6 py-16 text-center shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-xl font-semibold text-ink">登录后生成深度报告</p>
          <p className="mt-3 text-[14px] leading-relaxed text-ink-secondary">
            深度报告以命盘为据,由 AI 撰写长文解读。请先登录。
          </p>
          <Link
            href="/login?next=/reports"
            className="glow-gold mt-8 inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-6 py-2.5 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去登录
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-5 py-14 md:py-24">
      <header className="mb-12 md:mb-14">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">AI · 深度报告</p>
        <div className="mt-3 flex flex-wrap items-baseline justify-between gap-3">
          <h1 className="font-display text-[31px] font-semibold text-ink sm:text-[39px]">深度报告</h1>
          <CreditsBadge credits={credits} />
        </div>
        <p className="mt-3 text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">选定命主与主题,生成一份长文命理解读。</p>
      </header>

      {initLoading ? (
        <ReportSkeleton label="加载中…" />
      ) : noSource ? (
        <div className="rounded-[10px] bg-bg-raised px-6 py-20 text-center shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-xl font-semibold text-ink">先排一张命盘</p>
          <p className="mt-3 text-[14px] leading-relaxed text-ink-secondary">
            深度报告需要命主生辰。去排盘后,系统会记住你的生辰,或保存为档案供随时选用。
          </p>
          <Link
            href="/chart"
            className="glow-gold mt-8 inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-6 py-2.5 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去排盘
          </Link>
        </div>
      ) : (
        <div className="flex flex-col gap-10 md:gap-12">
          {/* 生辰来源 */}
          <section>
            <h2 className="mb-4 text-[12px] font-medium tracking-[0.08em] text-gold">生辰来源</h2>
            <div className="flex flex-col gap-2">
              {recentBirth && (
                <SourceRow
                  active={source === RECENT}
                  onClick={() => setSource(RECENT)}
                  title="最近排盘"
                  summary={formatBirthSummary(recentBirth)}
                />
              )}
              {profiles.map((p) => (
                <SourceRow
                  key={p.id}
                  active={source === p.id}
                  onClick={() => setSource(p.id)}
                  title={p.label}
                  summary={formatBirthSummary(p.birthInput)}
                />
              ))}
            </div>
          </section>

          {/* 主题 */}
          <section>
            <h2 className="mb-4 text-[12px] font-medium tracking-[0.08em] text-gold">报告主题</h2>
            <div className="flex flex-wrap gap-2">
              {REPORT_TOPICS.map((t) => {
                const active = topic === t.key;
                return (
                  <button
                    key={t.key}
                    type="button"
                    aria-pressed={active}
                    onClick={() => setTopic(t.key)}
                    className={[
                      "min-h-[44px] rounded-[6px] px-4 py-2 text-[14px] transition-colors",
                      active
                        ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                        : "bg-bg-raised text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] hover:text-gold hover:shadow-[inset_0_0_0_1px_var(--gold-dim)]",
                    ].join(" ")}
                  >
                    {t.label}
                  </button>
                );
              })}
            </div>
          </section>

          {/* 生成 */}
          <section>
            {outOfCredits ? (
              <div className="flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[inset_0_0_0_1px_var(--gold-dim)]">
                <p className="text-[14px] text-ink-secondary">深度报告次数不足,购买次卡后即可生成。</p>
                <Link
                  href="/pricing"
                  className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
                >
                  去购买
                </Link>
              </div>
            ) : (
              <div className="flex flex-col items-start gap-2.5">
                <button
                  type="button"
                  onClick={generate}
                  disabled={generating || !resolvedBirth}
                  className="glow-gold inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-7 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-50 disabled:shadow-none"
                >
                  {generating ? "生成中…" : "生成深度报告"}
                </button>
                <p className="text-[12px] text-ink-faint">将消耗 1 次深度报告。生成较慢,通常需要 30 秒以上。</p>
              </div>
            )}
          </section>

          {/* 结果区:生成中 / 错误 / 报告 */}
          {generating && <ReportSkeleton label="AI 正在撰写深度报告,请稍候(通常 30 秒以上,请勿离开)…" />}

          {!generating && error && (
            <ErrorPanel error={error} onRetry={generate} />
          )}

          {!generating && !error && report != null && (
            <article className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)] md:px-10 md:py-10">
              {degraded && (
                <p className="mb-4 rounded-[4px] bg-bg px-3 py-2 text-[12px] text-warn shadow-[inset_0_0_0_1px_var(--line)]">
                  当前为降级解读(备用模型),内容仅供参考。
                </p>
              )}
              <ReportText text={report} />
              <p className="mt-6 border-t border-line pt-4 text-[11px] leading-relaxed text-ink-faint">
                内容为传统文化与娱乐参考,不构成医疗投资建议。
              </p>
            </article>
          )}
        </div>
      )}
    </div>
  );
}

/** 剩余次数徽标 */
function CreditsBadge({ credits }: { credits: number | null }) {
  if (credits == null) return null;
  if (credits <= 0) {
    return (
      <Link
        href="/pricing"
        className="tnum inline-flex items-center gap-1.5 rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-colors hover:bg-bg-raised"
      >
        次数不足 · 去购买
      </Link>
    );
  }
  return (
    <span className="tnum inline-flex items-center gap-1.5 rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
      剩余 {credits} 次
    </span>
  );
}

/** 生辰来源可选行 */
function SourceRow({
  active,
  onClick,
  title,
  summary,
}: {
  active: boolean;
  onClick: () => void;
  title: string;
  summary: string;
}) {
  return (
    <button
      type="button"
      aria-pressed={active}
      onClick={onClick}
      className={[
        "flex min-h-[44px] flex-col items-start gap-0.5 rounded-[6px] px-4 py-3 text-left transition-shadow",
        active
          ? "bg-bg-raised shadow-[inset_0_0_0_1px_var(--gold-dim)]"
          : "bg-bg-raised shadow-[inset_0_0_0_1px_var(--line)] hover:shadow-[inset_0_0_0_1px_var(--line-strong)]",
      ].join(" ")}
    >
      <span className={`text-[14px] ${active ? "text-gold" : "text-ink"}`}>{title}</span>
      <span className="tnum text-[12px] text-ink-secondary">{summary}</span>
    </button>
  );
}

/** 生成/加载骨架 */
function ReportSkeleton({ label }: { label: string }) {
  return (
    <div
      className="rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8 md:py-8"
      role="status"
      aria-label={label}
    >
      <div className="h-5 w-1/3 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-5 flex flex-col gap-3">
        {[92, 100, 88, 96, 70].map((w, i) => (
          <div key={i} className="h-4 animate-pulse rounded-[2px] bg-line" style={{ width: `${w}%` }} aria-hidden />
        ))}
      </div>
      <p className="mt-6 text-[12px] text-ink-faint">{label}</p>
    </div>
  );
}

/** 错误态 */
function ErrorPanel({ error, onRetry }: { error: GenError; onRetry: () => void }) {
  if (error.kind === "no_credits") {
    return (
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[0_0_0_1px_var(--gold-dim)]">
        <p className="text-[14px] text-ink-secondary">深度报告次数不足,购买次卡后再试。</p>
        <Link
          href="/pricing"
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
        >
          去购买
        </Link>
      </div>
    );
  }
  if (error.kind === "ai_unavailable") {
    return (
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[0_0_0_1px_var(--warn)]">
        <p className="text-[14px] text-warn">AI 服务暂不可用,本次未扣次数,请稍后再试。</p>
        <button
          type="button"
          onClick={onRetry}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-bg px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
        >
          重试
        </button>
      </div>
    );
  }
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[0_0_0_1px_var(--danger)]">
      <p className="text-[14px] text-danger">{error.message || "生成失败,请重试。"}</p>
      <button
        type="button"
        onClick={onRetry}
        className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-bg px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
      >
        重试
      </button>
    </div>
  );
}
