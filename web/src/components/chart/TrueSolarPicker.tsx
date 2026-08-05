"use client";

import { useEffect, useMemo, useState } from "react";
import type { WorldCity, ProvinceCities } from "@/lib/types";
import { fetchWorldCities, fetchChinaCities } from "@/lib/api";

const fieldCls =
  "rounded-[6px] bg-bg px-3 py-2 text-[14px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow";

export interface TrueSolarValue {
  enabled: boolean;
  region: "cn" | "intl";
  province: string;
  city: string;
  worldCity: string;
}

export const emptyTrueSolar: TrueSolarValue = {
  enabled: false,
  region: "cn",
  province: "",
  city: "",
  worldCity: "",
};

/**
 * 真太阳时出生地选择:开关 + 国内省市 / 海外城市二选一。
 * 数据来自后端 /api/v1/cities(省市经度)与 /api/v1/world-cities(世界城市经度+时区)。
 * 国内以北京时(东经 120°)为基准;海外以当地时区标准经线为基准,均叠加均时差。
 */
export function TrueSolarPicker({
  value,
  onChange,
}: {
  value: TrueSolarValue;
  onChange: (v: TrueSolarValue) => void;
}) {
  const [provinces, setProvinces] = useState<ProvinceCities[]>([]);
  const [worldCities, setWorldCities] = useState<WorldCity[]>([]);

  // 首次开启时懒加载两份地理数据。
  useEffect(() => {
    if (!value.enabled) return;
    if (provinces.length === 0) {
      fetchChinaCities().then((r) => setProvinces(r.provinces)).catch(() => {});
    }
    if (worldCities.length === 0) {
      fetchWorldCities().then((r) => setWorldCities(r.cities)).catch(() => {});
    }
  }, [value.enabled, provinces.length, worldCities.length]);

  const cities = useMemo(
    () => provinces.find((p) => p.name === value.province)?.cities ?? [],
    [provinces, value.province],
  );

  // 海外城市按国家分组。
  const worldByCountry = useMemo(() => {
    const groups = new Map<string, WorldCity[]>();
    for (const c of worldCities) {
      const arr = groups.get(c.country) ?? [];
      arr.push(c);
      groups.set(c.country, arr);
    }
    return [...groups.entries()];
  }, [worldCities]);

  const selectedWorld = worldCities.find((c) => c.name === value.worldCity);

  return (
    <div className="basis-full">
      <label className="inline-flex cursor-pointer items-center gap-2 text-[13px] text-ink-secondary">
        <input
          type="checkbox"
          checked={value.enabled}
          onChange={(e) => onChange({ ...value, enabled: e.target.checked })}
          className="h-4 w-4 accent-[var(--gold)]"
        />
        真太阳时校正(按出生地经度与均时差还原时辰)
      </label>

      {value.enabled && (
        <div className="mt-3 flex flex-wrap items-end gap-3 rounded-[8px] bg-[var(--surface-1)] p-3 shadow-[inset_0_0_0_1px_var(--line)]">
          <div className="flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]">
            {(["cn", "intl"] as const).map((r) => (
              <button
                key={r}
                type="button"
                aria-pressed={value.region === r}
                onClick={() => onChange({ ...value, region: r })}
                className={[
                  "px-4 py-2 text-[13px] transition-colors",
                  value.region === r
                    ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                    : "bg-bg text-ink-secondary hover:text-ink",
                ].join(" ")}
              >
                {r === "cn" ? "中国" : "海外"}
              </button>
            ))}
          </div>

          {value.region === "cn" ? (
            <>
              <label className="flex flex-col gap-1">
                <span className="text-[12px] text-ink-faint">省 / 直辖市</span>
                <select
                  value={value.province}
                  onChange={(e) => onChange({ ...value, province: e.target.value, city: "" })}
                  className={fieldCls}
                >
                  <option value="">请选择</option>
                  {provinces.map((p) => (
                    <option key={p.name} value={p.name}>{p.name}</option>
                  ))}
                </select>
              </label>
              <label className="flex flex-col gap-1">
                <span className="text-[12px] text-ink-faint">城市</span>
                <select
                  value={value.city}
                  onChange={(e) => onChange({ ...value, city: e.target.value })}
                  disabled={!value.province}
                  className={`${fieldCls} disabled:opacity-50`}
                >
                  <option value="">请选择</option>
                  {cities.map((c) => (
                    <option key={c.name} value={c.name}>{c.name}</option>
                  ))}
                </select>
              </label>
            </>
          ) : (
            <label className="flex flex-col gap-1">
              <span className="text-[12px] text-ink-faint">出生城市</span>
              <select
                value={value.worldCity}
                onChange={(e) => onChange({ ...value, worldCity: e.target.value })}
                className={fieldCls}
              >
                <option value="">请选择</option>
                {worldByCountry.map(([country, list]) => (
                  <optgroup key={country} label={country}>
                    {list.map((c) => (
                      <option key={c.name} value={c.name}>{c.name}</option>
                    ))}
                  </optgroup>
                ))}
              </select>
            </label>
          )}

          {value.region === "intl" && selectedWorld?.dst && (
            <p className="basis-full text-[12px] text-ink-faint">
              注:{selectedWorld.country}实行夏令时(数据按标准时)。若出生于夏令时期间,
              请将出生钟点提前 1 小时后再对应时辰。
            </p>
          )}
        </div>
      )}
    </div>
  );
}
