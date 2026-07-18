"use client";

import { useState } from "react";
import {
  formatBirthSummary,
  relationLabel,
  type Profile,
} from "@/lib/profiles";

/**
 * 档案卡片:label + 关系徽标 + 生辰摘要 + 默认金色徽标。
 * 操作:载入排盘(主)、设为默认、删除(带确认层防误触)。
 */
export function ProfileCard({
  profile,
  busy,
  onLoad,
  onSetDefault,
  onDelete,
}: {
  profile: Profile;
  busy?: boolean;
  onLoad: (p: Profile) => void;
  onSetDefault: (p: Profile) => void;
  onDelete: (p: Profile) => void;
}) {
  const [confirming, setConfirming] = useState(false);

  return (
    <div className="lift flex flex-col rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="truncate font-display text-[20px] font-semibold text-ink">{profile.label}</h2>
          <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
            <span className="rounded-[2px] px-1.5 py-0.5 text-[11px] tracking-[0.08em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
              {relationLabel(profile.relation)}
            </span>
            {profile.isDefault && (
              <span className="inline-flex items-center gap-1 rounded-[2px] px-1.5 py-0.5 text-[11px] tracking-[0.08em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
                <StarIcon />
                默认
              </span>
            )}
          </div>
        </div>
      </div>

      <p className="tnum mt-4 text-[13px] leading-relaxed text-ink-secondary">{formatBirthSummary(profile.birthInput)}</p>

      <div className="mt-5 flex flex-wrap items-center gap-2 border-t border-line pt-4">
        <button
          type="button"
          onClick={() => onLoad(profile)}
          className="inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
        >
          载入排盘
        </button>
        {!profile.isDefault && (
          <button
            type="button"
            disabled={busy}
            onClick={() => onSetDefault(profile)}
            className="inline-flex min-h-[44px] items-center rounded-[6px] bg-bg px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-gold hover:shadow-[inset_0_0_0_1px_var(--gold-dim)] disabled:opacity-50"
          >
            设为默认
          </button>
        )}
        {confirming ? (
          <span className="ml-auto inline-flex items-center gap-2">
            <span className="text-[13px] text-ink-secondary">确认删除?</span>
            <button
              type="button"
              disabled={busy}
              onClick={() => onDelete(profile)}
              className="inline-flex min-h-[44px] items-center rounded-[6px] px-3 py-2 text-[14px] text-danger shadow-[inset_0_0_0_1px_var(--danger)] transition-colors hover:bg-bg disabled:opacity-50"
            >
              删除
            </button>
            <button
              type="button"
              onClick={() => setConfirming(false)}
              className="inline-flex min-h-[44px] items-center rounded-[6px] px-3 py-2 text-[14px] text-ink-secondary transition-colors hover:text-ink"
            >
              取消
            </button>
          </span>
        ) : (
          <button
            type="button"
            onClick={() => setConfirming(true)}
            aria-label="删除档案"
            className="ml-auto inline-flex min-h-[44px] items-center rounded-[6px] px-3 py-2 text-[14px] text-ink-faint transition-colors hover:text-danger"
          >
            <TrashIcon />
          </button>
        )}
      </div>
    </div>
  );
}

function StarIcon() {
  return (
    <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <path d="M12 2l2.9 6.3 6.9.8-5.1 4.7 1.4 6.8L12 17.8 5.9 21.4l1.4-6.8L2.2 9.9l6.9-.8L12 2z" />
    </svg>
  );
}

function TrashIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden>
      <path d="M4 7h16M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2m2 0v12a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
