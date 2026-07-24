"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import {
  getDivination,
  deleteDivination,
  luckTone,
  RELATION_TONE,
  DivinationError,
  type DivinationRecord,
  type LiuYaoResult,
  type MeihuaResult,
  type XiaoLiuRenResult,
} from "@/lib/divination";
import { DIVINATION_KIND_LABEL, formatDivinationTime } from "@/components/divination/format";
import { LiuYaoPan } from "@/components/divination/LiuYaoPan";
import { HexagramView } from "@/components/divination/HexagramView";
import { toneBadgeClass } from "@/components/divination/tone";
import { ReportText } from "@/components/profiles/ReportText";

/** 卦档详情:重现当时的完整卦象与 AI 解卦全文。 */

type LoadState =
  | { phase: "loading" }
  | { phase: "unauth" }
  | { phase: "error"; message: string }
  | { phase: "ready"; record: DivinationRecord };

export default function DivinationDetailPage() {
  const params = useParams<{ id: string }>();
  const [state, setState] = useState<LoadState>({ phase: "loading" });

  useEffect(() => {
    const load = () => {
      if (!currentUser()) {
        setState({ phase: "unauth" });
        return;
      }
      setState({ phase: "loading" });
      getDivination(params.id)
        .then((record) => setState({ phase: "ready", record }))
        .catch((e) =>
          setState({
            phase: "error",
            message: e instanceof DivinationError && e.status === 404 ? "卦档不存在或已删除。" : "卦档加载失败,请刷新重试。",
          }),
        );
    };
    load();
    window.addEventListener(AUTH_EVENT, load);
    return () => window.removeEventListener(AUTH_EVENT, load);
  }, [params.id]);

  return (
    <div className="mx-auto max-w-4xl px-5 py-14 md:py-20">
      <Link
        href="/divinations"
        className="inline-flex min-h-[44px] items-center gap-1.5 text-[13px] tracking-[0.06em] text-ink-faint transition-colors hover:text-ink"
      >
        <span aria-hidden>←</span> 卦档
      </Link>

      {state.phase === "unauth" && (
        <div className="mt-8 flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[inset_0_0_0_1px_var(--gold-dim)]">
          <p className="text-[14px] text-ink-secondary">登录后即可查看卦档。</p>
          <Link
            href={`/login?next=/divinations/${params.id}`}
            className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去登录
          </Link>
        </div>
      )}

      {state.phase === "loading" && (
        <div className="mt-8 flex flex-col gap-4" role="status" aria-label="卦档加载中">
          <div className="h-8 w-56 animate-pulse rounded-[4px] bg-bg-raised" aria-hidden />
          <div className="h-72 animate-pulse rounded-[10px] bg-bg-raised" aria-hidden />
        </div>
      )}

      {state.phase === "error" && <p className="mt-8 text-[14px] text-danger">{state.message}</p>}

      {state.phase === "ready" && <RecordView record={state.record} />}
    </div>
  );
}

function RecordView({ record }: { record: DivinationRecord }) {
  const router = useRouter();
  const [deleting, setDeleting] = useState(false);

  const remove = async () => {
    if (deleting) return;
    setDeleting(true);
    try {
      await deleteDivination(record.id);
      router.push("/divinations");
    } catch {
      setDeleting(false);
    }
  };

  return (
    <div className="mt-6">
      <header className="relative">
        <button
          type="button"
          onClick={remove}
          disabled={deleting}
          aria-label="删除此卦档"
          className="absolute right-0 top-0 min-h-[44px] rounded-[4px] px-2 py-1 text-[12px] text-ink-faint transition-colors hover:text-danger disabled:opacity-50"
        >
          {deleting ? "删除中…" : "删除"}
        </button>
        <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1.5 pr-14">
          <span className="rounded-[2px] px-1.5 py-0.5 text-[11px] leading-none tracking-[0.06em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
            {DIVINATION_KIND_LABEL[record.kind] ?? record.kind}
          </span>
          <span className="tnum text-[12px] tracking-[0.06em] text-ink-faint">{formatDivinationTime(record.castAt)}</span>
          {record.kind === "meihua" && (record.payload as MeihuaResult | undefined)?.lunarText && (
            <span className="tnum text-[12px] tracking-[0.06em] text-ink-faint">
              · 农历 {(record.payload as MeihuaResult).lunarText}
            </span>
          )}
        </div>
        <h1 className="mt-3 font-display text-[27px] font-semibold leading-snug text-ink sm:text-[31px]">
          {record.summary}
        </h1>
        {record.question && (
          <p className="mt-2 font-reading text-[16px] text-ink-secondary">所问:{record.question}</p>
        )}
      </header>

      {/* 卦象重现 */}
      <div className="mt-8">
        {record.kind === "liuyao" && record.payload && <LiuYaoPan result={record.payload as LiuYaoResult} />}
        {record.kind === "meihua" && record.payload && <MeihuaView result={record.payload as MeihuaResult} />}
        {record.kind === "xiaoliuren" && record.payload && (
          <XiaoLiuRenView result={record.payload as XiaoLiuRenResult} />
        )}
      </div>

      {/* AI 解卦全文(小六壬为快占,无解卦消费点) */}
      {record.kind !== "xiaoliuren" && (
      <div className="mt-8">
        {record.hasReading && record.reading ? (
          <article className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)] md:px-10 md:py-10">
            <p className="mb-5 text-[12px] font-medium tracking-[0.24em] text-gold">AI 深度解卦(归档)</p>
            <ReportText text={record.reading} />
            <p className="mt-6 border-t border-line pt-4 text-[11px] leading-relaxed text-ink-faint">
              占卜为传统文化参考,不构成决策建议。
            </p>
          </article>
        ) : (
          <p className="text-[13px] text-ink-faint">此卦未做 AI 解卦。解卦须在起卦当下进行,新问题可去对应板块重占。</p>
        )}
      </div>
      )}
    </div>
  );
}

/** 小六壬重现:三步掐指落位 + 断语。 */
function XiaoLiuRenView({ result }: { result: XiaoLiuRenResult }) {
  const steps = ["月", "日", "时"];
  return (
    <div className="rounded-[10px] bg-bg-raised px-5 py-7 shadow-[0_0_0_1px_var(--line)] sm:px-8">
      <p className="tnum text-[12px] tracking-[0.06em] text-ink-faint">起算 {result.lunarText}</p>
      <div className="mt-5 flex flex-wrap items-center gap-x-3 gap-y-2">
        {result.path.map((p, i) => (
          <span key={i} className="flex items-center gap-3">
            <span className="flex flex-col items-center gap-1">
              <span className="text-[11px] text-ink-faint">{steps[i]}</span>
              <span
                className={`rounded-[4px] px-2.5 py-1 font-display text-[17px] font-medium ${toneBadgeClass(luckTone(p.luck))}`}
              >
                {p.name}
              </span>
            </span>
            {i < result.path.length - 1 && (
              <span aria-hidden className="text-[13px] text-ink-faint">→</span>
            )}
          </span>
        ))}
      </div>
      <div className="mt-6 flex flex-wrap items-baseline gap-3">
        <span className="font-display text-[27px] font-semibold text-ink">{result.result.name}</span>
        <span className={`rounded-[2px] px-2 py-0.5 text-[13px] font-medium ${toneBadgeClass(luckTone(result.result.luck))}`}>
          {result.result.luck}
        </span>
      </div>
      <p className="mt-2 text-[14px] leading-relaxed text-ink-secondary">{result.result.meaning}</p>
    </div>
  );
}

/** 梅花卦象重现:本卦(大)+ 互/变 + 体用生克。 */
function MeihuaView({ result }: { result: MeihuaResult }) {
  const rel = RELATION_TONE[result.relation];
  return (
    <div>
      <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-8 shadow-[0_0_0_1px_var(--line)]">
        <HexagramView hexagram={result.ben} moving={result.moving} label="本卦" emphasis />
      </div>
      <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)]">
          <HexagramView hexagram={result.hu} moving={0} label="互卦" />
        </div>
        <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)]">
          <HexagramView hexagram={result.bian} moving={0} label="变卦" />
        </div>
      </div>
      {rel && (
        <div className="mt-4 flex flex-wrap items-center gap-3 rounded-[10px] bg-bg-raised px-5 py-5 shadow-[0_0_0_1px_var(--line)]">
          <span
            className={`inline-flex items-center rounded-[2px] px-2.5 py-1 text-[13px] font-medium tracking-[0.06em] ${toneBadgeClass(rel.tone)}`}
          >
            {rel.label}
          </span>
          <p className="flex-1 text-[14px] leading-relaxed text-ink-secondary">{result.verdict}</p>
        </div>
      )}
    </div>
  );
}
