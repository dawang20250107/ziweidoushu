"use client";

import { useEffect, useRef, useState } from "react";
import { DateSelect } from "./DateSelect";
import { fetchLunarYear, lunarToSolar, type LunarMonthMeta } from "@/lib/api";

const fieldCls =
  "rounded-[6px] bg-bg px-3 py-2 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow";

const LUNAR_MONTH_NAMES = ["正月", "二月", "三月", "四月", "五月", "六月", "七月", "八月", "九月", "十月", "冬月", "腊月"];
const LUNAR_DAY_NAMES = [
  "初一", "初二", "初三", "初四", "初五", "初六", "初七", "初八", "初九", "初十",
  "十一", "十二", "十三", "十四", "十五", "十六", "十七", "十八", "十九", "二十",
  "廿一", "廿二", "廿三", "廿四", "廿五", "廿六", "廿七", "廿八", "廿九", "三十",
];

/** 月 key:闰月加 L 前缀(如 "L3"),区分同名平/闰月。 */
const monthKey = (m: LunarMonthMeta) => (m.leap ? `L${m.month}` : `${m.month}`);
const monthLabel = (m: LunarMonthMeta) => (m.leap ? "闰" : "") + LUNAR_MONTH_NAMES[m.month - 1];
const pad = (n: number) => String(n).padStart(2, "0");

/**
 * 生日选择(公历/农历双模式,共享组件:排盘表单与合盘双方生辰共用)。
 * 对外唯一口径是公历 yyyy-mm-dd:农历模式下月/日选项按后端历表渲染
 * (哪年闰几月、大小月以 lunar-go 为准),每次选择即换算回写公历并展示预览;
 * 外部整体改值(档案填入/恢复)时农历选择失义,自动退回公历显示。
 */
export function CalendarDateField({
  value, onChange, onPendingChange,
}: {
  value: string; // 公历 yyyy-mm-dd
  onChange: (solarDate: string) => void;
  /** 农历换算进行中(父级可借此禁用提交,防止竞态提交旧值) */
  onPendingChange?: (pending: boolean) => void;
}) {
  const [calendar, setCalendar] = useState<"solar" | "lunar">("solar");
  const [lunarYear, setLunarYear] = useState(() => Number(value.split("-")[0]) || 1990);
  const [months, setMonths] = useState<LunarMonthMeta[]>([]);
  const [monthK, setMonthK] = useState("1");
  const [day, setDay] = useState(15);
  const [preview, setPreview] = useState("");
  const lastEmitted = useRef<string | null>(null);

  // 外部改值(非本组件回写)→ 农历选择已失义,退回公历显示
  useEffect(() => {
    if (calendar === "lunar" && value !== lastEmitted.current) {
      setCalendar("solar");
      setPreview("");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  // 农历年变化 → 拉当年月表;选中月不存在(如换到无此闰月的年)则回退首月
  useEffect(() => {
    if (calendar !== "lunar") return;
    let cancelled = false;
    fetchLunarYear(lunarYear)
      .then(({ months: ms }) => {
        if (cancelled) return;
        setMonths(ms);
        if (!ms.some((m) => monthKey(m) === monthK)) setMonthK(monthKey(ms[0]));
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [calendar, lunarYear]);

  const sel = months.find((m) => monthKey(m) === monthK);
  const dayMax = sel?.days ?? 30;

  // 农历选择齐备 → 换算回写公历(与提交同一端点,无第二套口径)
  useEffect(() => {
    if (calendar !== "lunar" || !sel) return;
    let cancelled = false;
    onPendingChange?.(true);
    lunarToSolar({ year: lunarYear, month: sel.month, leap: sel.leap, day: Math.min(day, dayMax) })
      .then((s) => {
        if (cancelled) return;
        const str = `${s.year}-${pad(s.month)}-${pad(s.day)}`;
        setPreview(str);
        lastEmitted.current = str;
        onChange(str);
      })
      .catch(() => {
        if (!cancelled) setPreview("");
      })
      .finally(() => {
        if (!cancelled) onPendingChange?.(false);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [calendar, lunarYear, monthK, day, months]);

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center gap-2">
        <span className="text-[12px] text-ink-faint">生日</span>
        <div className="flex overflow-hidden rounded-[4px] shadow-[inset_0_0_0_1px_var(--line)]" role="radiogroup" aria-label="历法">
          {(["solar", "lunar"] as const).map((c) => (
            <button
              key={c}
              type="button"
              role="radio"
              aria-checked={calendar === c}
              onClick={() => setCalendar(c)}
              className={[
                "px-2 py-0.5 text-[11px] transition-colors",
                calendar === c
                  ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                  : "bg-bg text-ink-secondary hover:text-ink",
              ].join(" ")}
            >
              {c === "solar" ? "公历" : "农历"}
            </button>
          ))}
        </div>
      </div>
      {calendar === "solar" ? (
        <DateSelect value={value} onChange={onChange} />
      ) : (
        <div className="flex flex-wrap items-center gap-1.5">
          <select
            value={lunarYear}
            onChange={(e) => setLunarYear(Number(e.target.value))}
            className={fieldCls}
            aria-label="农历年"
          >
            {Array.from({ length: 201 }, (_, i) => 2100 - i).map((y) => (
              <option key={y} value={y}>
                {y} 年
              </option>
            ))}
          </select>
          <select
            value={monthK}
            onChange={(e) => setMonthK(e.target.value)}
            className={fieldCls}
            aria-label="农历月"
          >
            {months.map((m) => (
              <option key={monthKey(m)} value={monthKey(m)}>
                {monthLabel(m)}
              </option>
            ))}
          </select>
          <select
            value={Math.min(day, dayMax)}
            onChange={(e) => setDay(Number(e.target.value))}
            className={fieldCls}
            aria-label="农历日"
          >
            {LUNAR_DAY_NAMES.slice(0, dayMax).map((label, idx) => (
              <option key={idx} value={idx + 1}>
                {label}
              </option>
            ))}
          </select>
        </div>
      )}
      {calendar === "lunar" && preview && (
        <span className="tnum text-[11px] text-ink-faint">≈ 公历 {preview}</span>
      )}
    </div>
  );
}
