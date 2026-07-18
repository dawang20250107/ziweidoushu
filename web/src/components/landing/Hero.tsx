import Link from "next/link";
import { HeroChart } from "./HeroChart";

/** 落地页第一屏:左标题区 + 右真实迷你星盘。动效预算集中在星盘。 */
export function Hero() {
  return (
    <section className="mx-auto max-w-6xl px-4 pt-20 pb-24 md:pt-32 md:pb-32">
      <div className="grid items-center gap-14 lg:grid-cols-[1.05fr_0.95fr] lg:gap-20">
        {/* 标题区 */}
        <div className="max-w-xl">
          <p className="flex items-center gap-3 text-[12px] font-medium tracking-[0.24em] text-gold">
            <span className="h-px w-6 bg-gold-dim" aria-hidden />
            观星台 · 紫微斗数
          </p>
          <h1 className="mt-6 text-balance font-display text-[39px] font-bold leading-[1.08] text-ink sm:text-[49px] md:text-[61px]">
            把紫微斗数,排到毫厘
          </h1>
          <p className="mt-6 max-w-lg text-[16px] leading-[1.75] text-ink-secondary md:text-[17px]">
            完整安星与格局判定,运限逐层下钻;古籍全文可查,解读依盘而言。以倪海厦《天纪》体系为口径,逐字段对齐经典。
          </p>
          <div className="mt-10 flex flex-wrap items-center gap-4">
            <Link
              href="/chart"
              className="glow-gold inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-7 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
            >
              开始排盘
            </Link>
            <Link
              href="/library"
              className="inline-flex min-h-[44px] items-center rounded-[6px] border border-line-strong px-7 py-3 text-[15px] text-ink transition-colors hover:border-gold-dim hover:text-gold"
            >
              查阅古籍
            </Link>
          </div>
        </div>

        {/* 真实迷你星盘 */}
        <div className="mx-auto w-full max-w-[440px] lg:max-w-none">
          <HeroChart />
        </div>
      </div>
    </section>
  );
}
