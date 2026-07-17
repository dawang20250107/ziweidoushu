"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AUTH_EVENT, currentUser, logout, type AuthUser } from "@/lib/auth";

const TIER_LABEL: Record<string, string> = { free: "", pro: "Pro", master: "大师" };

/** 顶栏用户区:未登录显示「登录」,已登录显示昵称与登出。 */
export function UserMenu() {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const sync = () => setUser(currentUser());
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, []);

  if (!user) {
    return (
      <Link
        href="/login"
        className="rounded-[6px] px-3 py-1.5 text-[15px] text-ink-secondary transition-colors hover:bg-bg-raised hover:text-ink"
      >
        登录
      </Link>
    );
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        className="flex items-center gap-1.5 rounded-[6px] px-3 py-1.5 text-[15px] text-ink transition-colors hover:bg-bg-raised"
      >
        {user.nickname}
        {TIER_LABEL[user.tier] && (
          <span className="rounded-[2px] px-1 text-[10px] tracking-[0.08em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
            {TIER_LABEL[user.tier]}
          </span>
        )}
      </button>
      {open && (
        <div className="absolute right-0 top-full z-50 mt-1 w-36 rounded-[6px] bg-bg-overlay p-1 shadow-[0_0_0_1px_var(--line-strong),0_8px_24px_rgba(0,0,0,0.3)]">
          {[
            { href: "/account", label: "我的账户" },
            { href: "/profiles", label: "命盘档案" },
            { href: "/reports", label: "深度报告" },
          ].map((item) => (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setOpen(false)}
              className="block rounded-[4px] px-3 py-1.5 text-[14px] text-ink-secondary transition-colors hover:bg-bg-raised hover:text-ink"
            >
              {item.label}
            </Link>
          ))}
          <div className="mx-2 my-1 h-px bg-line" aria-hidden />
          <button
            type="button"
            onClick={async () => {
              setOpen(false);
              await logout();
            }}
            className="w-full rounded-[4px] px-3 py-1.5 text-left text-[14px] text-ink-secondary transition-colors hover:bg-bg-raised hover:text-ink"
          >
            退出登录
          </button>
        </div>
      )}
    </div>
  );
}
