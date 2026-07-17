import Link from "next/link";
import {
  formatDate,
  TIER_LABEL,
  type EntitlementsResponse,
  type Tier,
} from "@/lib/billing";

/**
 * 权益区:有效订阅时段 + 次数余额。两者皆空时引导去定价页。
 */
export function EntitlementsPanel({ data }: { data: EntitlementsResponse }) {
  const subs = data.entitlements ?? [];
  const deepReport = data.credits?.deep_report ?? 0;
  const empty = subs.length === 0 && deepReport === 0;

  if (empty) {
    return (
      <div className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]">
        <h2 className="font-display text-lg font-semibold text-ink">我的权益</h2>
        <div className="mt-4 rounded-[6px] px-4 py-8 text-center shadow-[inset_0_0_0_1px_var(--line)]">
          <p className="text-[14px] text-ink-secondary">还没有任何有效权益。</p>
          <Link
            href="/pricing"
            className="mt-4 inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-6 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去选购方案
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]">
      <h2 className="font-display text-lg font-semibold text-ink">我的权益</h2>

      {/* 订阅时段 */}
      <div className="mt-4">
        <p className="text-[12px] tracking-[0.08em] text-ink-faint">订阅时段</p>
        {subs.length > 0 ? (
          <ul className="mt-2 flex flex-col gap-2">
            {subs.map((e, i) => (
              <li
                key={`${e.tier}-${e.startsAt}-${i}`}
                className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1 rounded-[6px] px-3 py-2.5 shadow-[inset_0_0_0_1px_var(--line)]"
              >
                <span className="text-[14px] font-medium text-gold">
                  {TIER_LABEL[e.tier as Tier] ?? e.tier}
                </span>
                <span className="text-[13px] text-ink-secondary tnum">
                  {formatDate(e.startsAt)} — {formatDate(e.endsAt)}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="mt-2 text-[13px] text-ink-faint">暂无有效订阅。</p>
        )}
      </div>

      {/* 次数余额 */}
      <div className="mt-5">
        <p className="text-[12px] tracking-[0.08em] text-ink-faint">次数余额</p>
        <div className="mt-2 flex items-center justify-between rounded-[6px] px-3 py-2.5 shadow-[inset_0_0_0_1px_var(--line)]">
          <span className="text-[14px] text-ink-secondary">深度报告</span>
          <span className="font-display text-[15px] text-ink tnum">剩余 {deepReport} 次</span>
        </div>
      </div>
    </div>
  );
}
