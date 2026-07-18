/** 骨架文本:加载态占位,弱化的 --line 条,不喧宾夺主。 */
export function SkeletonLines({
  lines = 4,
  className = "",
}: {
  lines?: number;
  className?: string;
}) {
  return (
    <div className={`flex flex-col gap-3 ${className}`} role="status" aria-label="加载中">
      {Array.from({ length: lines }).map((_, i) => (
        <div
          key={i}
          className="h-4 animate-pulse rounded-[2px] bg-line"
          style={{ width: `${92 - (i % 3) * 16}%` }}
          aria-hidden
        />
      ))}
      <span className="sr-only">加载中…</span>
    </div>
  );
}

/** 书架星牌骨架(与立式星卡同形) */
export function SkeletonCard() {
  return (
    <div
      className="flex flex-col rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]"
      role="status"
      aria-label="加载中"
    >
      <div className="mb-5 h-12 animate-pulse rounded-[6px] bg-line" aria-hidden />
      <div className="h-3 w-1/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-3 h-6 w-3/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-5 flex flex-col gap-2">
        <div className="h-3 w-full animate-pulse rounded-[2px] bg-line" aria-hidden />
        <div className="h-3 w-11/12 animate-pulse rounded-[2px] bg-line" aria-hidden />
        <div className="h-3 w-2/3 animate-pulse rounded-[2px] bg-line" aria-hidden />
      </div>
      <div className="mt-6 h-3 w-2/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <span className="sr-only">加载中…</span>
    </div>
  );
}
