"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { ThemeToggle } from "./ThemeProvider";
import { UserMenu } from "./UserMenu";

export const NAV_ITEMS = [
  { href: "/chart", label: "排盘" },
  { href: "/meihua", label: "梅花" },
  { href: "/liuyao", label: "六爻" },
  { href: "/liuren", label: "六壬" },
  { href: "/heming", label: "合盘" },
  { href: "/library", label: "古籍" },
  { href: "/chat", label: "问星" },
  { href: "/pricing", label: "定价" },
];

function isActive(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(href + "/");
}

/** 桌面导航:当前板块金色短下划线高亮。 */
export function DesktopNav() {
  const pathname = usePathname();
  return (
    <nav className="hidden items-center gap-1 md:flex">
      {NAV_ITEMS.map((item) => {
        const active = isActive(pathname, item.href);
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={active ? "page" : undefined}
            className={[
              "relative whitespace-nowrap rounded-[6px] px-3 py-1.5 text-[15px] transition-colors",
              active ? "text-ink" : "text-ink-secondary hover:bg-bg-raised hover:text-ink",
            ].join(" ")}
          >
            {item.label}
            {active && (
              <span aria-hidden className="absolute inset-x-3 -bottom-px h-[2px] rounded-full bg-gold" />
            )}
          </Link>
        );
      })}
      <span className="mx-2 h-4 w-px bg-line" aria-hidden />
      <ThemeToggle />
      <UserMenu />
    </nav>
  );
}

/** 移动端导航:汉堡按钮 + 全宽抽屉(路由变化自动收起)。 */
export function MobileNav() {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  // 导航后自动收起
  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  return (
    <div className="md:hidden">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label={open ? "关闭菜单" : "打开菜单"}
        className="flex min-h-[44px] min-w-[44px] items-center justify-center rounded-[6px] text-ink-secondary transition-colors hover:text-ink"
      >
        {open ? (
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
            <path d="M6 6l12 12M18 6L6 18" strokeLinecap="round" />
          </svg>
        ) : (
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
            <path d="M4 7h16M4 12h16M4 17h16" strokeLinecap="round" />
          </svg>
        )}
      </button>

      {open && (
        <>
          {/* 遮罩 */}
          <div
            className="fixed inset-0 top-14 z-40 bg-[rgba(0,0,0,0.5)]"
            onClick={() => setOpen(false)}
            aria-hidden
          />
          {/* 抽屉面板 */}
          <div className="fixed inset-x-0 top-14 z-50 border-b border-line bg-bg-overlay px-4 pb-4 pt-2 shadow-[0_16px_40px_rgba(0,0,0,0.5)]">
            <nav className="flex flex-col">
              {NAV_ITEMS.map((item) => {
                const active = isActive(pathname, item.href);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    aria-current={active ? "page" : undefined}
                    className={[
                      "flex min-h-[48px] items-center whitespace-nowrap rounded-[6px] px-3 text-[16px] transition-colors",
                      active ? "text-gold" : "text-ink-secondary hover:bg-bg-raised hover:text-ink",
                    ].join(" ")}
                  >
                    {item.label}
                    {active && <span aria-hidden className="ml-2 h-1 w-1 rounded-full bg-gold" />}
                  </Link>
                );
              })}
            </nav>
            <div className="mt-2 flex items-center justify-between border-t border-line pt-3">
              <ThemeToggle />
              <UserMenu />
            </div>
          </div>
        </>
      )}
    </div>
  );
}
