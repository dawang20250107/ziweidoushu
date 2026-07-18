/** 计费页骨架屏:加载态占位,弱化 --line 条,不喧宾夺主。 */

/** 定价卡骨架 */
export function PricingCardSkeleton() {
  return (
    <div
      className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]"
      role="status"
      aria-label="加载中"
    >
      <div className="h-5 w-1/3 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-5 h-9 w-2/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-6 flex flex-col gap-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-3.5 w-4/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
        ))}
      </div>
      <div className="mt-7 h-11 w-full animate-pulse rounded-[6px] bg-line" aria-hidden />
      <span className="sr-only">加载中…</span>
    </div>
  );
}

/** 账户区块骨架(用户卡 / 权益卡 / 订单行通用) */
export function AccountBlockSkeleton({ lines = 3 }: { lines?: number }) {
  return (
    <div
      className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]"
      role="status"
      aria-label="加载中"
    >
      <div className="h-5 w-1/4 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-4 flex flex-col gap-3">
        {Array.from({ length: lines }).map((_, i) => (
          <div
            key={i}
            className="h-4 animate-pulse rounded-[2px] bg-line"
            style={{ width: `${88 - (i % 3) * 14}%` }}
            aria-hidden
          />
        ))}
      </div>
      <span className="sr-only">加载中…</span>
    </div>
  );
}
