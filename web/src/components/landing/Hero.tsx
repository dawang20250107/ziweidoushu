import Link from "next/link";
import { HeroChart } from "./HeroChart";

/** 落地页第一屏:左标题区 + 右真实迷你星盘。动效预算集中在星盘与极光底。 */
export function Hero() {
  return (
    <section className="relative mx-auto max-w-6xl px-4 pt-20 pb-16 md:pt-32 md:pb-20">
      {/* 极光星云底:紫金双斑缓慢漂移(宣纸主题自动隐藏) */}
      <div className="aurora -z-10" aria-hidden />
      {/* 下滑退焦:镜头移开首屏,焦点交给能力区(scroll-driven,渐进增强) */}
      <div className="hero-recede grid items-center gap-14 lg:grid-cols-[1.05fr_0.95fr] lg:gap-20">
        {/* 标题区 */}
        <div className="max-w-xl">
          <p className="flex items-center gap-3 text-[12px] font-medium tracking-[0.24em] text-gold">
            <span className="h-px w-6 bg-gold-dim" aria-hidden />
            观星台 · 紫微斗数
          </p>
          <h1 className="mt-6 text-balance font-display text-[39px] font-bold leading-[1.12] text-ink sm:text-[49px] md:text-[61px]">
            把紫微斗数,
            <br className="hidden sm:block" />
            <span className="text-gold-gradient">排到毫厘</span>
          </h1>
          <p className="mt-6 max-w-lg text-[16px] leading-[1.75] text-ink-secondary md:text-[17px]">
            完整安星与格局判定,运限逐层下钻;古籍全文可查,解读依盘而言。以倪海厦《天纪》体系为口径,逐字段对齐经典。
          </p>
          <div className="mt-10 flex flex-wrap items-center gap-4">
            <Link
              href="/chart"
              className="cta-breathe inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-7 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
            >
              开始排盘
            </Link>
            <Link
              href="/library"
              className="inline-flex min-h-[44px] items-center rounded-[6px] border border-line-strong px-7 py-3 text-[15px] text-ink transition-all hover:border-gold-dim hover:text-gold hover:shadow-[0_0_18px_rgba(217,179,108,0.12)]"
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

      {/* 下滑指引:金线下探,一经滚动即隐 */}
      <div className="hero-hint mt-14 flex flex-col items-center gap-2 md:mt-20" aria-hidden>
        <span className="text-[11px] tracking-[0.3em] text-ink-faint">下滑探索</span>
        <span className="hint-bob h-7 w-px bg-gradient-to-b from-transparent via-gold-dim to-gold" />
      </div>
    </section>
  );
}
