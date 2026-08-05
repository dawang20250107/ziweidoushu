"use client";

import { useState } from "react";
import type { BirthInfo, Gender } from "@/lib/types";
import { HOUR_NAMES } from "@/lib/types";
import { DateSelect } from "@/components/ui/DateSelect";
import { TrueSolarPicker, emptyTrueSolar, type TrueSolarValue } from "@/components/chart/TrueSolarPicker";

const fieldCls =
  "rounded-[6px] bg-bg px-3 py-2 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow";

/** 排盘输入表单(公历生辰 + 时辰 + 性别)。 */
export function BirthForm({
  initial, loading, onSubmit,
}: {
  initial?: BirthInfo;
  loading?: boolean;
  onSubmit: (b: BirthInfo) => void;
}) {
  const [name, setName] = useState(initial?.name ?? "");
  const [date, setDate] = useState(
    initial ? `${initial.year}-${String(initial.month).padStart(2, "0")}-${String(initial.day).padStart(2, "0")}` : "1990-06-15",
  );
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

  function submit(e: React.FormEvent) {
    e.preventDefault();
    const [y, m, d] = date.split("-").map(Number);
    if (!y || !m || !d || y < 1900 || y > 2100) {
      setError("请输入 1900-2100 之间的有效公历日期");
      return;
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
        <span className="text-[12px] text-ink-faint">公历生日</span>
        <DateSelect value={date} onChange={setDate} />
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
        disabled={loading}
        className="glow-gold w-full min-h-[44px] rounded-[6px] bg-gold px-6 py-2 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-50 disabled:shadow-none sm:w-auto"
      >
        {loading ? "排盘中…" : "排盘"}
      </button>
      <TrueSolarPicker value={solar} onChange={setSolar} />
      {error && <p className="basis-full text-[13px] text-danger">{error}</p>}
    </form>
  );
}
