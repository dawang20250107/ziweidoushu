import Link from "next/link";

/** 尾部 CTA 条:再次引导开始排盘。 */
export function ClosingCta() {
  return (
    <section className="mx-auto max-w-6xl px-4 pb-24 md:pb-32">
      <div
        className="relative overflow-hidden rounded-[10px] bg-bg-raised px-8 py-14 md:px-14 md:py-20"
        style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}
      >
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 -z-0"
          style={{ background: "radial-gradient(70% 120% at 100% 0%, var(--gold-glow), transparent 60%)" }}
        />
        <div className="aurora" aria-hidden />
        <div className="relative flex flex-col items-start gap-8 md:flex-row md:items-center md:justify-between md:gap-12">
          <div>
            <h2 className="text-balance font-display text-[31px] font-semibold leading-snug text-ink md:text-[39px]">
              输入生辰,开始排盘
            </h2>
            <p className="mt-3 max-w-lg text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">
              一张盘,几秒钟;之后的运限、古籍与解读,都从这里展开。
            </p>
          </div>
          <Link
            href="/chart"
            className="cta-breathe inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-8 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            开始排盘
          </Link>
        </div>
      </div>
    </section>
  );
}
