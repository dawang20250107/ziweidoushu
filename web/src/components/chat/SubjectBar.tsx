"use client";

import Link from "next/link";
import type { BirthInfo } from "@/lib/types";
import { HOUR_NAMES } from "@/lib/types";

const pad = (n: number) => String(n).padStart(2, "0");

/** 命主摘要条:姓名/生辰/时辰/性别 + 换盘 + 清空对话。 */
export function SubjectBar({
  birth,
  canClear,
  onClear,
}: {
  birth: BirthInfo;
  canClear: boolean;
  onClear: () => void;
}) {
  const dateText = `${birth.year}-${pad(birth.month)}-${pad(birth.day)}`;
  const hourText = HOUR_NAMES[birth.hour] ?? "";
  const genderText = birth.gender === "male" ? "男" : "女";

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-2.5 shadow-[0_0_0_1px_var(--line)]">
      <div className="flex flex-wrap items-baseline gap-x-2.5 gap-y-1 text-[14px]">
        <span className="font-display text-[15px] font-semibold text-ink">
          {birth.name || "命主"}
        </span>
        <span className="tnum text-ink-secondary">{dateText}</span>
        <span className="text-ink-secondary">{hourText}</span>
        <span className="text-ink-secondary">{genderText}</span>
      </div>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onClear}
          disabled={!canClear}
          className="rounded-[2px] px-2.5 py-1 text-[13px] text-ink-faint transition-colors hover:text-ink disabled:cursor-not-allowed disabled:opacity-40"
        >
          清空对话
        </button>
        <Link
          href="/chart"
          className="rounded-[2px] border border-line-strong px-3 py-1 text-[13px] text-ink-secondary transition-colors hover:border-gold-dim hover:text-gold"
        >
          换盘
        </Link>
      </div>
    </div>
  );
}
