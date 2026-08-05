"use client";

import { useEffect, useState } from "react";
import type { BirthInfo, Gender } from "@/lib/types";
import { HOUR_NAMES } from "@/lib/types";
import { DateSelect } from "@/components/ui/DateSelect";
import { TrueSolarPicker, emptyTrueSolar, type TrueSolarValue } from "@/components/chart/TrueSolarPicker";
import { fetchLunarYear, lunarToSolar, type LunarMonthMeta, ApiError } from "@/lib/api";

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
const monthLabel = (m: LunarMonthMeta) =>
  (m.leap ? "闰" : "") + LUNAR_MONTH_NAMES[m.month - 1];

/**
 * 排盘输入表单:公历/农历双模式 + 时辰 + 性别(+ 真太阳时)。
 * 农历模式按后端历表渲染月/日(哪年有闰几月、大小月以 lunar-go 为准),
 * 提交前换算为公历——下游(排盘/档案/问星)仍以公历为唯一事实源。
 */
export function BirthForm({
  initial, loading, onSubmit,
}: {
  initial?: BirthInfo;
  loading?: boolean;
  onSubmit: (b: BirthInfo) => void;
}) {
  const [name, setName] = useState(initial?.name ?? "");
  const [calendar, setCalendar] = useState<"solar" | "lunar">("solar");
  const [date, setDate] = useState(
    initial ? `${initial.year}-${String(initial.month).padStart(2, "0")}-${String(initial.day).padStart(2, "0")}` : "1990-06-15",
  );
  // 农历模式选择
  const [lunarYear, setLunarYear] = useState(initial?.year ?? 1990);
  const [lunarMonths, setLunarMonths] = useState<LunarMonthMeta[]>([]);
  const [lunarMonth, setLunarMonth] = useState("5");
  const [lunarDay, setLunarDay] = useState(1);
  const [solarPreview, setSolarPreview] = useState("");

  const [hour, setHour] = useState(initial?.hour ?? 6);
  const [gender, setGender] = useState<Gender>(initial?.gender ?? "male");
  const [solar, setSolar] = useState<TrueSolarValue>(() =>
    initial?.trueSolarTime
      ? {
          enabled: true,
          region: initial.worldCity ? "intl" : "cn",
          province: initial.province ?? "",
          city: initial.city ?? "",
          worldCity: initial.worldCity ?? "",
        }
      : emptyTrueSolar,
  );
  const [error, setError] = useState("");
  const [converting, setConverting] = useState(false);

  // 农历年变化 → 拉当年月表(含闰月位置与大小月);选中月不存在则回退首月
  useEffect(() => {
    if (calendar !== "lunar") return;
    let cancelled = false;
    fetchLunarYear(lunarYear)
      .then(({ months }) => {
        if (cancelled) return;
        setLunarMonths(months);
        if (!months.some((m) => monthKey(m) === lunarMonth)) {
          setLunarMonth(monthKey(months[0]));
        }
      })
      .catch(() => {
        if (!cancelled) setError("农历历表加载失败,请稍后重试");
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [calendar, lunarYear]);

  const selMonth = lunarMonths.find((m) => monthKey(m) === lunarMonth);
  const dayMax = selMonth?.days ?? 30;

  // 选满年月日 → 预览对应公历(与提交同一换算端点,不另算)
  useEffect(() => {
    if (calendar !== "lunar" || !selMonth) return;
    const d = Math.min(lunarDay, dayMax);
    let cancelled = false;
    lunarToSolar({ year: lunarYear, month: selMonth.month, leap: selMonth.leap, day: d })
      .then((s) => {
        if (!cancelled) setSolarPreview(`${s.year}-${String(s.month).padStart(2, "0")}-${String(s.day).padStart(2, "0")}`);
      })
      .catch(() => {
        if (!cancelled) setSolarPreview("");
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [calendar, lunarYear, lunarMonth, lunarDay, lunarMonths]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    let y: number, m: number, d: number;
    if (calendar === "lunar") {
      if (!selMonth) {
        setError("请先选择农历月份");
        return;
      }
      setConverting(true);
      try {
        const s = await lunarToSolar({
          year: lunarYear,
          month: selMonth.month,
          leap: selMonth.leap,
          day: Math.min(lunarDay, dayMax),
        });
        y = s.year; m = s.month; d = s.day;
      } catch (err) {
        setError(err instanceof ApiError ? err.message : "农历换算失败,请重试");
        setConverting(false);
        return;
      }
      setConverting(false);
    } else {
      [y, m, d] = date.split("-").map(Number);
      if (!y || !m || !d || y < 1900 || y > 2100) {
        setError("请输入 1900-2100 之间的有效公历日期");
        return;
      }
    }
    if (solar.enabled) {
      const hasPlace =
        solar.region === "cn" ? solar.province && solar.city : solar.worldCity;
      if (!hasPlace) {
        setError("已开启真太阳时,请选择出生地");
        return;
      }
    }
    setError("");
    const b: BirthInfo = { year: y, month: m, day: d, hour, gender, name: name || undefined };
    if (solar.enabled) {
      b.trueSolarTime = true;
      if (solar.region === "cn") {
        b.province = solar.province;
        b.city = solar.city;
      } else {
        b.worldCity = solar.worldCity;
      }
    }
    onSubmit(b);
  }

  return (
    <form onSubmit={submit} className="flex flex-wrap items-end gap-3" aria-label="排盘信息">
      <label className="flex flex-col gap-1">
        <span className="text-[12px] text-ink-faint">姓名(可选)</span>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="命主"
          maxLength={12}
          className={`${fieldCls} w-28`}
        />
      </label>

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
          <DateSelect value={date} onChange={setDate} />
        ) : (
          <div className="flex items-center gap-1.5">
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
              value={lunarMonth}
              onChange={(e) => setLunarMonth(e.target.value)}
              className={fieldCls}
              aria-label="农历月"
            >
              {lunarMonths.map((m) => (
                <option key={monthKey(m)} value={monthKey(m)}>
                  {monthLabel(m)}
                </option>
              ))}
            </select>
            <select
              value={Math.min(lunarDay, dayMax)}
              onChange={(e) => setLunarDay(Number(e.target.value))}
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
        {calendar === "lunar" && solarPreview && (
          <span className="tnum text-[11px] text-ink-faint">≈ 公历 {solarPreview}</span>
        )}
      </div>

      <label className="flex flex-col gap-1">
        <span className="text-[12px] text-ink-faint">时辰</span>
        <select value={hour} onChange={(e) => setHour(Number(e.target.value))} className={fieldCls}>
          {HOUR_NAMES.map((label, idx) => (
            <option key={idx} value={idx}>
              {label}
            </option>
          ))}
        </select>
      </label>
      <div className="flex flex-col gap-1">
        <span className="text-[12px] text-ink-faint">性别</span>
        <div className="flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]">
          {(["male", "female"] as const).map((g) => (
            <button
              key={g}
              type="button"
              aria-pressed={gender === g}
              onClick={() => setGender(g)}
              className={[
                "px-4 py-2 text-[14px] transition-colors",
                gender === g
                  ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                  : "bg-bg text-ink-secondary hover:text-ink",
              ].join(" ")}
            >
              {g === "male" ? "男" : "女"}
            </button>
          ))}
        </div>
      </div>
      <button
        type="submit"
        disabled={loading || converting}
        className="glow-gold w-full min-h-[44px] rounded-[6px] bg-gold px-6 py-2 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-50 disabled:shadow-none sm:w-auto"
      >
        {loading ? "排盘中…" : converting ? "换算中…" : "排盘"}
      </button>
      <TrueSolarPicker value={solar} onChange={setSolar} />
      {error && <p className="basis-full text-[13px] text-danger">{error}</p>}
    </form>
  );
}
