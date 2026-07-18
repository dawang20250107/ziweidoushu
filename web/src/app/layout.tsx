import type { Metadata, Viewport } from "next";
import Link from "next/link";
import "./globals.css";
import { ThemeProvider, themeInitScript } from "@/components/theme/ThemeProvider";
import { StarField } from "@/components/theme/StarField";
import { DesktopNav, MobileNav } from "@/components/theme/NavLinks";

export const metadata: Metadata = {
  title: { default: "观星台 · 紫微斗数", template: "%s · 观星台" },
  description:
    "基于倪海厦《天纪》体系的紫微斗数排盘与古籍查阅平台:完整安星算法、格局判定、运限下钻、古籍全文检索、AI 解读。",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  themeColor: "#090c17", // 玄穹为产品默认
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body className="bg-bg text-ink">
        <StarField />
        <ThemeProvider>
          <header className="sticky top-0 z-40 border-b border-line bg-bg/90 backdrop-blur">
            <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
              <Link href="/" className="whitespace-nowrap font-display text-lg font-semibold tracking-wide text-ink">
                观星台<span className="ml-2 hidden text-[12px] font-normal tracking-[0.24em] text-gold sm:inline">紫微斗数</span>
              </Link>
              <DesktopNav />
              <MobileNav />
            </div>
          </header>
          <main>{children}</main>
          <footer className="mt-16 border-t border-line py-8 text-center text-[13px] text-ink-faint">
            <p>命理内容为传统文化与娱乐参考,不构成医疗、投资或重大决策建议。</p>
          </footer>
        </ThemeProvider>
      </body>
    </html>
  );
}
