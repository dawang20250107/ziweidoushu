"use client";

import Link from "next/link";
import { TOSS_OPTIONS } from "@/lib/divination";

// 占卜工作台 · 输入侧小部件(自 Workbench.tsx 拆出,纯展示,零业务状态)。

/** 剩余解卦次数徽标。 */
export function CreditsBadge({ credits }: { credits: number | null }) {
  if (credits == null) return null;
  if (credits <= 0) {
    return (
      <Link
        href="/pricing"
        className="tnum inline-flex items-center rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-colors hover:bg-bg-raised"
      >
        解卦次数不足 · 去购买
      </Link>
    );
  }
  return (
    <span className="tnum inline-flex items-center rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
      解卦剩余 {credits} 次
    </span>
  );
}

/** 起卦方式分段项。 */
export function MethodTab({
  active,
  onClick,
  title,
  hint,
}: {
  active: boolean;
  onClick: () => void;
  title: string;
  hint: string;
}) {
  return (
    <button
      type="button"
      aria-pressed={active}
      onClick={onClick}
      className={[
        "flex min-h-[44px] flex-col items-start gap-0.5 rounded-[6px] px-4 py-2.5 text-left transition-shadow",
        active
          ? "bg-[var(--gold-glow)] shadow-[inset_0_0_0_1px_var(--gold-dim)]"
          : "bg-bg shadow-[inset_0_0_0_1px_var(--line)] hover:shadow-[inset_0_0_0_1px_var(--line-strong)]",
      ].join(" ")}
    >
      <span className={`text-[14px] font-medium ${active ? "text-gold" : "text-ink"}`}>{title}</span>
      <span className="text-[11px] text-ink-faint">{hint}</span>
    </button>
  );
}

/** 报数输入(1-999,仅数字)。 */
export function NumField({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[11px] text-ink-faint">{label}</span>
      <input
        type="text"
        inputMode="numeric"
        value={value}
        onChange={(e) => onChange(e.target.value.replace(/\D/g, "").slice(0, 3))}
        placeholder="1-999"
        className="tnum w-24 rounded-[6px] bg-bg px-3 py-2.5 text-center text-[16px] text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow placeholder:text-ink-faint focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
      />
    </label>
  );
}

const YAO_NAMES = ["初爻", "二爻", "三爻", "四爻", "五爻", "上爻"];

/** 六爻报爻录入:初爻在上(先摇先录),每爻四选一(背面数)。 */
export function TossEntry({ tosses, onChange }: { tosses: (number | null)[]; onChange: (t: (number | null)[]) => void }) {
  const set = (i: number, backs: number) => {
    const next = [...tosses];
    next[i] = backs;
    onChange(next);
  };
  return (
    <div className="mt-4 flex flex-col gap-2">
      <p className="text-[12px] leading-relaxed text-ink-faint">
        以三枚铜钱自摇六次,自初爻起逐次录入每掷的背面枚数(字面朝上不计)。
      </p>
      {YAO_NAMES.map((name, i) => (
        <div key={name} className="flex items-center gap-2.5">
          <span className="w-9 shrink-0 text-[12px] text-ink-secondary">{name}</span>
          <div className="flex flex-1 flex-wrap gap-1.5">
            {TOSS_OPTIONS.map((opt) => {
              const active = tosses[i] === opt.backs;
              return (
                <button
                  key={opt.backs}
                  type="button"
                  aria-pressed={active}
                  onClick={() => set(i, opt.backs)}
                  className={[
                    "flex min-h-[38px] flex-col items-center justify-center rounded-[4px] px-2.5 py-1 transition-shadow",
                    active
                      ? "bg-[var(--gold-glow)] shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                      : "bg-bg shadow-[inset_0_0_0_1px_var(--line)] hover:shadow-[inset_0_0_0_1px_var(--line-strong)]",
                  ].join(" ")}
                >
                  <span className={`text-[12px] leading-tight ${active ? "text-gold" : "text-ink"}`}>{opt.label}</span>
                  <span className="text-[10px] leading-tight text-ink-faint">{opt.hint}</span>
                </button>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}
