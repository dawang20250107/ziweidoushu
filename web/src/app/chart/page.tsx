"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { fetchChart, fetchHoroscope, ApiError } from "@/lib/api";
import type { BirthInfo, ChartResponse, Horoscope, HoroscopeReading } from "@/lib/types";
import { DENSITY_LABELS, type Density } from "@/lib/chart-helpers";
import { BirthForm } from "@/components/chart/BirthForm";
import { ChartBoard } from "@/components/chart/ChartBoard";
import { PalaceDrawer } from "@/components/chart/PalaceDrawer";
import { PatternsOverview } from "@/components/chart/PatternsOverview";
import { TimelineBar, type TimelineSelection } from "@/components/chart/TimelineBar";
import { SiZhuPanel } from "@/components/chart/SiZhuPanel";
import { ReadingPanel } from "@/components/chart/ReadingPanel";
import { HoroscopeReadingPanel } from "@/components/chart/HoroscopeReadingPanel";
import { TimingPicker } from "@/components/chart/TimingPicker";
import { LuopanCast } from "@/components/chart/LuopanCast";
import { SaveProfileButton } from "@/components/profiles/SaveProfileButton";

/** 排盘工作台:盘面 + 运限时间轴 + 宫位详情。 */
export default function ChartPage() {
  const [birth, setBirth] = useState<BirthInfo | null>(null);
  const [data, setData] = useState<ChartResponse | null>(null);
  const [horoscope, setHoroscope] = useState<Horoscope | null>(null);
  const [horoReading, setHoroReading] = useState<HoroscopeReading | null>(null);
  const [timeline, setTimeline] = useState<TimelineSelection>({ year: null });
  const [density, setDensity] = useState<Density>("pro");
  const [selectedBranch, setSelectedBranch] = useState<number | null>(null);
  const [hlPalaces, setHlPalaces] = useState<string[] | null>(null); // 格局悬停联动点亮的宫名
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [casting, setCasting] = useState(false); // 罗盘起盘仪式中
  const [castLeaving, setCastLeaving] = useState(false); // 仪式收场淡出中
  const [boardCue, setBoardCue] = useState(0); // 仪式收束→盘面聚焦重放(级联/命宫点睛)
  const [spot, setSpot] = useState(false); // 收束光圈:幕布拉开时聚拢盘面再散开
  const boardWrapRef = useRef<HTMLDivElement>(null);

  const [initialBirth, setInitialBirth] = useState<BirthInfo | null>(null);

  const runChart = useCallback(async (b: BirthInfo): Promise<boolean> => {
    setLoading(true);
    setError("");
    setHoroscope(null);
    setHoroReading(null);
    setTimeline({ year: null });
    setSelectedBranch(null);
    try {
      const resp = await fetchChart(b);
      setBirth(b);
      setData(resp);
      // 跨页契约:问星页(/chat)读取最近一次排盘的生辰
      localStorage.setItem("ziwei-birth", JSON.stringify(b));
      return true;
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "排盘失败,请稍后重试");
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  // 手动排盘:星光击罗盘全屏仪式(≥3.6s 全序列 + 0.5s 淡出收场),
  // 收场时镜头交接:光圈聚拢盘面、十二宫级联重放、盘面滚至视口中心。
  // 恢复路径与 reduced-motion 直出。
  const castChart = useCallback(async (b: BirthInfo) => {
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduced) {
      await runChart(b);
      return;
    }
    setCasting(true);
    const minShow = new Promise((r) => setTimeout(r, 3600));
    const [ok] = await Promise.all([runChart(b), minShow]);
    if (ok) {
      setBoardCue((c) => c + 1);
      setSpot(true);
    }
    setCastLeaving(true);
    await new Promise((r) => setTimeout(r, 500));
    setCasting(false);
    setCastLeaving(false);
    if (ok) setTimeout(() => setSpot(false), 650);
  }, [runChart]);

  // 仪式期间锁页面滚动;幕布开始拉开即解锁,让聚焦滚动接管
  useEffect(() => {
    const lock = casting && !castLeaving;
    document.body.style.overflow = lock ? "hidden" : "";
    return () => {
      document.body.style.overflow = "";
    };
  }, [casting, castLeaving]);

  // 仪式收束:盘面滚至视口中心(与幕布淡出、级联点亮同步进行)
  useEffect(() => {
    if (boardCue === 0) return;
    const raf = requestAnimationFrame(() => {
      boardWrapRef.current?.scrollIntoView({ behavior: "smooth", block: "center" });
    });
    return () => cancelAnimationFrame(raf);
  }, [boardCue]);

  // 挂载时恢复最近一次排盘生辰并自动出盘(档案「载入排盘」/刷新续排共用 ziwei-birth 契约)
  useEffect(() => {
    try {
      const raw = localStorage.getItem("ziwei-birth");
      if (!raw) return;
      const b = JSON.parse(raw) as BirthInfo;
      if (typeof b?.year === "number" && b.year >= 1900 && b.year <= 2100 && b.gender) {
        setInitialBirth(b);
        void runChart(b);
      }
    } catch {
      // 本地数据损坏则忽略,走空白表单
    }
  }, [runChart]);

  // 时间轴选择 → 拉运限。未下钻的层取稳定默认(7 月 15 日午时,避开农历年界)。
  useEffect(() => {
    if (!birth || timeline.year == null) {
      setHoroscope(null);
      setHoroReading(null);
      return;
    }
    let cancelled = false;
    fetchHoroscope(birth, {
      year: timeline.year,
      month: timeline.month ?? 7,
      day: timeline.day ?? 15,
      hour: timeline.hour ?? 6,
    })
      .then((resp) => {
        if (!cancelled) {
          setHoroscope(resp.horoscope);
          setHoroReading(resp.reading ?? null);
        }
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "运限计算失败");
      });
    return () => {
      cancelled = true;
    };
  }, [birth, timeline.year, timeline.month, timeline.day, timeline.hour]);

  // 下钻深度决定盘面叠加层
  const overlayScopes = useMemo(() => {
    const scopes: Array<"decadal" | "yearly" | "monthly" | "daily" | "hourly"> = ["decadal", "yearly"];
    if (timeline.month != null) scopes.push("monthly");
    if (timeline.day != null) scopes.push("daily");
    if (timeline.hour != null) scopes.push("hourly");
    return scopes;
  }, [timeline.month, timeline.day, timeline.hour]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-6">
      <h1 className="sr-only">紫微斗数排盘工作台</h1>
      <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
        <BirthForm key={initialBirth ? "restored" : "blank"} initial={initialBirth ?? undefined} loading={loading} onSubmit={castChart} />
        {data && birth && (
          <div className="flex flex-wrap items-center gap-2">
            <Link
              href="/famous"
              className="inline-flex min-h-[44px] items-center rounded-[6px] bg-bg-raised px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-gold hover:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
            >
              名人盘库
            </Link>
            <SaveProfileButton birth={birth} />
            <div className="flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]" role="radiogroup" aria-label="显示密度">
              {(Object.keys(DENSITY_LABELS) as Density[]).map((d) => (
                <button
                  key={d}
                  type="button"
                  role="radio"
                  aria-checked={density === d}
                  onClick={() => setDensity(d)}
                  className={[
                    "px-3 py-1.5 text-[13px] transition-colors",
                    density === d
                      ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                      : "bg-bg-raised text-ink-secondary hover:text-ink",
                  ].join(" ")}
                >
                  {DENSITY_LABELS[d]}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {error && (
        <p className="mb-4 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {!data && !loading && (
        <div className="rounded-[6px] bg-bg-raised px-6 py-16 text-center text-ink-faint shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-lg text-ink-secondary">输入生辰,开始排盘</p>
          <p className="mt-2 text-[13px]">安星算法与 iztro 逐字段对齐,1500+ 黄金基准回归保障。</p>
        </div>
      )}

      {data && (
        <div className="flex flex-col gap-4">
          <TimelineBar chart={data.chart} selection={timeline} horoscope={horoscope} onChange={setTimeline} />
          {/* 盘面全宽:格局下沉为独立分区、宫位详情浮出为抽屉,盘面不再与长侧栏比高。
              选宫时桌面端右侧让位抽屉(padding 缓动平移,连线由 ResizeObserver 追踪) */}
          <div>
            <div
              className={[
                "relative transition-[padding] duration-500",
                selectedBranch != null ? "lg:pr-[376px]" : "",
              ].join(" ")}
              style={{ transitionTimingFunction: "var(--ease-out)" }}
            >
              <div className="overflow-x-auto">
                {/* key=boardCue:仪式收束时整盘重挂,十二宫级联与命宫点睛随之重放 */}
                <div
                  ref={boardWrapRef}
                  key={boardCue}
                  className={["mx-auto min-w-[640px] max-w-[1120px]", boardCue > 0 ? "board-focus" : ""].join(" ")}
                >
                  <ChartBoard
                    chart={data.chart}
                    density={density}
                    selectedBranch={selectedBranch}
                    onSelectBranch={setSelectedBranch}
                    horoscope={horoscope}
                    overlayScopes={overlayScopes}
                    patterns={data.patterns ?? []}
                    highlightNames={hlPalaces}
                    onPatternHover={setHlPalaces}
                  />
                </div>
              </div>
              {/* 移动端:右缘渐隐提示盘面可横向滑动 */}
              <div
                aria-hidden
                className="pointer-events-none absolute inset-y-0 right-0 w-10 bg-gradient-to-l from-bg to-transparent md:hidden"
              />
            </div>
            <p className="mt-1.5 text-center text-[11px] text-ink-faint md:hidden">左右滑动查看全盘</p>
            <p className="mt-1.5 hidden text-center text-[11px] text-ink-faint md:block">
              点击宫位查看三方四正与星曜细目
            </p>
          </div>

          {/* 格局总览:全宽卡片墙(长文多列铺开,悬停点亮盘上关联宫位) */}
          <PatternsOverview patterns={data.patterns ?? []} onPatternHover={setHlPalaces} />

          {/* 运限断语:随时间轴选择的目标日期逐层生成(大限→流年→流月→流日→流时) */}
          {horoReading && (
            <div className="reveal">
              <HoroscopeReadingPanel reading={horoReading} />
            </div>
          )}

          {/* 多维断语:逐宫断语骨架(随盘而异,确定性);
              下方各分区与落地页同一套滚动聚焦节奏(reveal 渐进增强) */}
          {data.reading && (
            <div className="reveal">
              <ReadingPanel reading={data.reading} />
            </div>
          )}

          {/* 事项择吉:选事项 → 利年/利月/利日(确定性) */}
          {birth && (
            <div className="reveal">
              <TimingPicker birth={birth} />
            </div>
          )}

          {/* 四柱视角:八字附加层(可折叠) */}
          {data.chart.siZhu && (
            <div className="reveal">
              <SiZhuPanel siZhu={data.chart.siZhu} />
            </div>
          )}
        </div>
      )}

      {/* 宫位详情抽屉:桌面右缘滑入、移动端底部上滑 */}
      {data && (
        <PalaceDrawer
          chart={data.chart}
          patterns={data.patterns ?? []}
          selectedBranch={selectedBranch}
          onClose={() => setSelectedBranch(null)}
        />
      )}

      {casting && <LuopanCast leaving={castLeaving} />}
      {/* 收束光圈:幕布拉开瞬间光聚盘面,再徐徐散开(镜头交接) */}
      {spot && <div className="cast-spot" aria-hidden />}
    </div>
  );
}
