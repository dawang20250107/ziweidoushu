import Link from "next/link";

/** 尾部 CTA 条:再次引导开始排盘。 */
export function ClosingCta() {
  return (
    <section className="mx-auto max-w-6xl px-4 pb-20 md:pb-28">
      <div
        className="relative overflow-hidden rounded-[10px] bg-bg-raised px-6 py-12 md:px-12 md:py-16"
        style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}
      >
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 -z-0"
          style={{ background: "radial-gradient(70% 120% at 100% 0%, var(--gold-glow), transparent 60%)" }}
        />
        <div className="relative flex flex-col items-start gap-6 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-balance font-display text-[27px] font-semibold leading-snug text-ink md:text-[33px]">
              输入生辰,开始排盘
            </h2>
            <p className="mt-2 text-[15px] text-ink-secondary">一张盘,几秒钟;之后的运限、古籍与解读,都从这里展开。</p>
          </div>
          <Link
            href="/chart"
            className="shrink-0 rounded-[6px] bg-gold px-7 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            开始排盘
          </Link>
        </div>
      </div>
    </section>
  );
}
