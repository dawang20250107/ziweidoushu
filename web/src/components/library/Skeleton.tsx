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

/** 书架卡片骨架 */
export function SkeletonCard() {
  return (
    <div className="rounded-[6px] bg-bg-raised p-5 shadow-[0_0_0_1px_var(--line)]" role="status" aria-label="加载中">
      <div className="h-6 w-2/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-3 h-3 w-1/4 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-4 flex flex-col gap-2">
        <div className="h-3 w-full animate-pulse rounded-[2px] bg-line" aria-hidden />
        <div className="h-3 w-11/12 animate-pulse rounded-[2px] bg-line" aria-hidden />
        <div className="h-3 w-3/4 animate-pulse rounded-[2px] bg-line" aria-hidden />
      </div>
      <span className="sr-only">加载中…</span>
    </div>
  );
}
