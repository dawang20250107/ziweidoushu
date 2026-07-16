"use client";

import { useEffect, useState } from "react";
import { fetchChart } from "@/lib/api";
import { BRANCHES } from "@/lib/types";
import type { Chart, Palace } from "@/lib/types";
import {
  BOARD_GRID_TEMPLATE, ENTER_ORDER_BY_BRANCH, GRID_AREA_BY_BRANCH, brightnessVar,
} from "@/lib/chart-helpers";

/** Hero 示例命盘的固定生辰(1964 甲辰 · 巳时 · 男)。 */
const DEMO_BIRTH = { year: 1964, month: 9, day: 10, hour: 5, gender: "male" } as const;

type Status = "loading" | "ready" | "error";

/** 迷你宫格:仅主星名 + 亮度上色 + 宫名/地支底标。 */
function MiniCell({ branch, palace }: { branch: number; palace?: Palace }) {
  const major = palace?.stars.filter((s) => s.type === "major") ?? [];
  const borrowed = palace?.borrowedStars ?? [];
  const isMing = palace?.isMingGong;

  return (
    <div
      className="palace-enter relative flex min-h-0 flex-col rounded-[6px] bg-bg-raised p-1.5"
      style={{
        gridArea: GRID_AREA_BY_BRANCH[branch],
        animationDelay: `${ENTER_ORDER_BY_BRANCH[branch] * 34}ms`,
        boxShadow: isMing ? "inset 0 0 0 1px var(--gold-dim)" : "inset 0 0 0 1px var(--line)",
        backgroundImage: isMing ? "linear-gradient(var(--gold-glow), transparent 60%)" : undefined,
      }}
    >
      <div className="flex flex-1 flex-wrap content-start gap-x-1.5 gap-y-0.5 overflow-hidden">
        {major.map((s) => (
          <span
            key={s.name}
            className="font-display text-[12px] font-semibold leading-tight"
            style={{ color: brightnessVar(s.brightness) }}
          >
            {s.name}
          </span>
        ))}
        {major.length === 0 && borrowed.length > 0 && (
          <span className="text-[11px] leading-tight text-ink-faint">借{borrowed[0]}</span>
        )}
      </div>
      <div className="mt-1 flex items-baseline justify-between">
        <span className="font-display text-[11px] leading-none text-ink-secondary">
          {palace?.name ?? "　"}
        </span>
        <span className="text-[10px] leading-none text-ink-faint">{BRANCHES[branch]}</span>
      </div>
    </div>
  );
}

/** 中宫:命主概要;数据未就绪时显示品牌印记。 */
function MiniCenter({ chart }: { chart?: Chart }) {
  return (
    <div
      className="flex flex-col items-center justify-center rounded-[6px] bg-bg px-3 text-center"
      style={{ gridArea: "center", boxShadow: "inset 0 0 0 1px var(--line)" }}
    >
      {chart ? (
        <>
          <div className="font-display text-[15px] font-semibold text-ink">
            命主<span className="ml-1.5 text-[12px] font-normal text-ink-secondary">男命</span>
          </div>
          <div className="tnum mt-1 text-[11px] text-ink-faint">
            {chart.birthInfo.year}-{chart.birthInfo.month}-{chart.birthInfo.day} {chart.timeName}
          </div>
          <div className="mt-1.5 text-[11px] text-ink-secondary">
            {chart.wuxingJuName} · {BRANCHES[chart.mingGongBranch]}宫命
          </div>
        </>
      ) : (
        <>
          <div className="font-display text-[15px] font-semibold text-ink-secondary">紫微垣</div>
          <div className="mt-1 text-[11px] text-ink-faint">十二宫 · 待安星</div>
        </>
      )}
    </div>
  );
}

/** Hero 右侧:客户端拉取真实排盘,渲染 4×4 简化盘;加载时 12 空宫骨架。 */
export function HeroChart() {
  const [status, setStatus] = useState<Status>("loading");
  const [chart, setChart] = useState<Chart | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchChart(DEMO_BIRTH)
      .then((resp) => {
        if (cancelled) return;
        setChart(resp.chart);
        setStatus("ready");
      })
      .catch(() => {
        if (!cancelled) setStatus("error");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const palaceByBranch = new Map<number, Palace>();
  chart?.palaces.forEach((p) => palaceByBranch.set(p.branch, p));

  return (
    <figure className="relative m-0">
      {/* 星象微光:单色柔光,非渐变强调 */}
      <div
        aria-hidden
        className="pointer-events-none absolute -inset-6 -z-10"
        style={{
          background: "radial-gradient(60% 55% at 65% 30%, var(--gold-glow), transparent 70%)",
        }}
      />
      <div className="rounded-[10px] bg-bg-raised/40 p-3" style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}>
        <figcaption className="mb-2.5 flex items-center justify-between px-0.5">
          <span className="flex items-center gap-1.5 text-[11px] tracking-[0.16em] text-ink-secondary">
            <span
              className="inline-block size-1 rounded-full"
              style={{ background: status === "loading" ? "var(--ink-faint)" : "var(--gold)" }}
            />
            示例命盘
          </span>
          <span className="tnum text-[11px] text-ink-faint">1964 · 甲辰</span>
        </figcaption>

        <div
          className={["grid aspect-square w-full gap-1", status === "loading" ? "animate-pulse" : ""].join(" ")}
          style={{
            gridTemplateAreas: BOARD_GRID_TEMPLATE,
            gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
            gridTemplateRows: "repeat(4, minmax(0, 1fr))",
          }}
          role="img"
          aria-label="紫微斗数示例命盘 · 1964 年生辰"
        >
          {Array.from({ length: 12 }, (_, branch) => (
            <MiniCell key={branch} branch={branch} palace={palaceByBranch.get(branch)} />
          ))}
          <MiniCenter chart={chart ?? undefined} />
        </div>
      </div>
    </figure>
  );
}
