"use client";

import Link from "next/link";
import { ReportText } from "@/components/profiles/ReportText";

// 占卜工作台 · AI 深度解卦区(自 Workbench.tsx 拆出)。

export type AiErr =
  | { kind: "unauth" }
  | { kind: "no_credits" }
  | { kind: "ai_unavailable" }
  | { kind: "network"; message: string };

/** AI 深度解卦区:登录/次数/加载/错误/结果分态。 */
export function AiSection({
  signedIn,
  credits,
  loading,
  reading,
  error,
  hint,
  archived,
  homePath,
  onDivine,
}: {
  signedIn: boolean | null;
  credits: number | null;
  loading: boolean;
  reading: string | null;
  error: AiErr | null;
  hint: string;
  archived?: boolean;
  homePath: string;
  onDivine: () => void;
}) {
  // 已出结果
  if (reading != null && !loading) {
    return (
      <article className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)] md:px-10 md:py-10">
        <p className="mb-5 text-[12px] font-medium tracking-[0.24em] text-gold">AI 深度解卦</p>
        <ReportText text={reading} />
        <div className="mt-6 flex flex-wrap items-center justify-between gap-2 border-t border-line pt-4">
          <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
            {credits != null && <span className="tnum text-[12px] text-ink-faint">解卦剩余 {credits} 次</span>}
            {archived && (
              <Link href="/divinations" className="text-[12px] text-gold transition-opacity hover:opacity-80">
                已存入卦档 · 查看
              </Link>
            )}
          </span>
          <span className="text-[11px] leading-relaxed text-ink-faint">占卜为传统文化参考,不构成决策建议。</span>
        </div>
      </article>
    );
  }

  // 加载骨架(解卦较慢,>20s 常见)
  if (loading) {
    return (
      <div
        className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)]"
        role="status"
        aria-label="AI 正在解卦"
      >
        <div className="h-4 w-24 animate-pulse rounded-[2px] bg-line" aria-hidden />
        <div className="mt-5 flex flex-col gap-3">
          {[94, 100, 86, 96, 72].map((w, i) => (
            <div key={i} className="h-4 animate-pulse rounded-[2px] bg-line" style={{ width: `${w}%` }} aria-hidden />
          ))}
        </div>
        <p className="mt-6 text-[12px] text-ink-faint">AI 正在依卦象逐层解读,通常需 20 秒以上,请勿离开…</p>
      </div>
    );
  }

  // 错误分态
  if (error) {
    if (error.kind === "unauth") {
      return (
        <PromptBar tone="gold" text="登录后即可 AI 深度解卦。" action={{ href: `/login?next=${homePath}`, label: "去登录" }} />
      );
    }
    if (error.kind === "no_credits") {
      return (
        <PromptBar tone="gold" text="解卦次数不足,购买次卡后再试。" action={{ href: "/pricing", label: "去购买" }} />
      );
    }
    if (error.kind === "ai_unavailable") {
      return <PromptBar tone="warn" text="AI 服务暂不可用,本次未扣次数,请稍后再试。" retry={onDivine} />;
    }
    return <PromptBar tone="danger" text={error.message || "解卦失败,请重试。"} retry={onDivine} />;
  }

  // 初始:登录 / 次数 / 解卦按钮(本区唯一金色辉光主 CTA)
  if (signedIn === false) {
    return (
      <Link
        href={`/login?next=${homePath}`}
        className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-7 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright sm:w-auto"
      >
        登录后 AI 深度解卦
      </Link>
    );
  }
  if (signedIn && credits === 0) {
    return <PromptBar tone="gold" text="解卦次数不足,购买次卡后即可深度解卦。" action={{ href: "/pricing", label: "去购买" }} />;
  }
  return (
    <div className="flex flex-col items-start gap-2.5">
      <button
        type="button"
        onClick={onDivine}
        className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-7 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright sm:w-auto"
      >
        AI 深度解卦(消耗 1 次)
      </button>
      <p className="text-[12px] text-ink-faint">
        {hint}
        {credits != null ? `,当前剩余 ${credits} 次` : ""}。
      </p>
    </div>
  );
}

/** 提示条:登录/购买/重试。 */
function PromptBar({
  tone,
  text,
  action,
  retry,
}: {
  tone: "gold" | "warn" | "danger";
  text: string;
  action?: { href: string; label: string };
  retry?: () => void;
}) {
  const ring =
    tone === "gold"
      ? "shadow-[inset_0_0_0_1px_var(--gold-dim)]"
      : tone === "warn"
        ? "shadow-[inset_0_0_0_1px_var(--warn)]"
        : "shadow-[inset_0_0_0_1px_var(--danger)]";
  const textColor = tone === "gold" ? "text-ink-secondary" : tone === "warn" ? "text-warn" : "text-danger";
  return (
    <div className={`flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 ${ring}`}>
      <p className={`text-[14px] ${textColor}`}>{text}</p>
      {action && (
        <Link
          href={action.href}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
        >
          {action.label}
        </Link>
      )}
      {retry && (
        <button
          type="button"
          onClick={retry}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-bg px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
        >
          重试
        </button>
      )}
    </div>
  );
}
