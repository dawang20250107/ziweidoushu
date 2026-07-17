"use client";

/**
 * 公历日期选择:年/月/日三段 select,替换原生 date input(系统日历弹层无法主题化,
 * 深空盘面里突兀;select 收起态可完全按设计系统描边,移动端弹出原生滚轮体验更好)。
 * 值契约与原 input 相同:yyyy-mm-dd 字符串,非法输入不可能产生。
 */

const fieldCls =
  "rounded-[6px] bg-bg px-2 py-2 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow";

function daysIn(y: number, m: number): number {
  return new Date(y, m, 0).getDate();
}

export function DateSelect({
  value,
  onChange,
  minYear = 1900,
  maxYear = 2100,
  ariaLabel = "公历生日",
}: {
  value: string; // yyyy-mm-dd
  onChange: (date: string) => void;
  minYear?: number;
  maxYear?: number;
  ariaLabel?: string;
}) {
  const [ys, ms, ds] = value.split("-");
  const y = Number(ys) || 1990;
  const m = Number(ms) || 1;
  const d = Number(ds) || 1;

  const set = (ny: number, nm: number, nd: number) => {
    const clamped = Math.min(nd, daysIn(ny, nm));
    onChange(`${ny}-${String(nm).padStart(2, "0")}-${String(clamped).padStart(2, "0")}`);
  };

  const years: number[] = [];
  for (let i = maxYear; i >= minYear; i--) years.push(i);

  return (
    <div className="flex items-center gap-1.5" role="group" aria-label={ariaLabel}>
      <select
        value={y}
        onChange={(e) => set(Number(e.target.value), m, d)}
        className={`${fieldCls} tnum`}
        aria-label="年"
      >
        {years.map((yy) => (
          <option key={yy} value={yy}>
            {yy} 年
          </option>
        ))}
      </select>
      <select
        value={m}
        onChange={(e) => set(y, Number(e.target.value), d)}
        className={`${fieldCls} tnum`}
        aria-label="月"
      >
        {Array.from({ length: 12 }, (_, i) => i + 1).map((mm) => (
          <option key={mm} value={mm}>
            {mm} 月
          </option>
        ))}
      </select>
      <select
        value={d}
        onChange={(e) => set(y, m, Number(e.target.value))}
        className={`${fieldCls} tnum`}
        aria-label="日"
      >
        {Array.from({ length: daysIn(y, m) }, (_, i) => i + 1).map((dd) => (
          <option key={dd} value={dd}>
            {dd} 日
          </option>
        ))}
      </select>
    </div>
  );
}
