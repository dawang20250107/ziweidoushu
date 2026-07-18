"use client";

import { useState } from "react";
import { formatDate, type MeUser } from "@/lib/billing";
import { TierBadge } from "@/components/billing/badges";

/**
 * 账户用户卡:头像字符 + 昵称 + 层级徽标 + 到期时间 + 退出登录。
 */
export function AccountUserCard({
  user,
  onLogout,
}: {
  user: MeUser;
  onLogout: () => void | Promise<void>;
}) {
  const [busy, setBusy] = useState(false);
  const monogram = Array.from(user.nickname.trim())[0] ?? "星";

  async function handleLogout() {
    if (busy) return;
    setBusy(true);
    try {
      await onLogout();
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-5 rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)] md:p-8">
      <span
        className="grid h-14 w-14 shrink-0 place-items-center rounded-full font-display text-xl text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
        aria-hidden
      >
        {monogram}
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-display text-lg font-semibold text-ink">{user.nickname}</span>
          <TierBadge tier={user.tier} />
        </div>
        <p className="mt-1 text-[13px] text-ink-secondary">
          {user.tier === "free"
            ? "体验版账户 · 升级解锁全部能力"
            : user.tierExpiresAt
              ? `会员有效期至 ${formatDate(user.tierExpiresAt)}`
              : "会员生效中"}
        </p>
      </div>

      <button
        type="button"
        onClick={handleLogout}
        disabled={busy}
        className="inline-flex min-h-[44px] items-center justify-center rounded-[6px] px-4 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink disabled:opacity-40"
      >
        {busy ? "退出中…" : "退出登录"}
      </button>
    </div>
  );
}
