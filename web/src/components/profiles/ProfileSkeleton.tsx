/** 档案卡片骨架:加载态占位。 */
export function ProfileSkeleton() {
  return (
    <div className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]" role="status" aria-label="加载中">
      <div className="h-5 w-2/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-3 h-4 w-1/4 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-4 h-3 w-3/5 animate-pulse rounded-[2px] bg-line" aria-hidden />
      <div className="mt-5 flex gap-2 border-t border-line pt-4">
        <div className="h-9 w-24 animate-pulse rounded-[6px] bg-line" aria-hidden />
        <div className="h-9 w-20 animate-pulse rounded-[6px] bg-line" aria-hidden />
      </div>
      <span className="sr-only">加载中…</span>
    </div>
  );
}
