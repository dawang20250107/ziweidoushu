"use client";

import { useEffect, useRef, useState } from "react";
import { prefersReducedMotion } from "@/components/divination/useReducedMotion";

/**
 * 逐爻摇卦:三枚铜钱六掷,每掷翻转后揭示正反面(字/背),爻画自下而上逐爻落定。
 * 随机源为浏览器 crypto(起卦免费环节);六掷齐后把 tosses 交回页面,
 * 由服务端按记录装卦——AI 解卦的同一卦契约(tosses+castAt 回传)不受影响。
 */

const YAO_NAMES = ["初爻", "二爻", "三爻", "四爻", "五爻", "上爻"];
const FLIP_MS = 950;

/** 单掷:3 枚铜钱各随机字/背,返回背面枚数 0-3。 */
function tossCoins(): { backs: number; faces: boolean[] } {
  const buf = new Uint8Array(3);
  crypto.getRandomValues(buf);
  const faces = Array.from(buf, (b) => b % 2 === 0); // true=背
  return { backs: faces.filter(Boolean).length, faces };
}

function tossLabel(backs: number): string {
  return ["老阴 · 动", "少阳", "少阴", "老阳 · 动"][backs];
}

/** 已落一爻的迷你爻画行。 */
function YaoRow({ index, backs, faces }: { index: number; backs: number; faces: boolean[] }) {
  const yang = backs === 1 || backs === 3;
  const moving = backs === 0 || backs === 3;
  const bar = {
    height: 7,
    background: moving ? "var(--gold)" : "var(--ink-secondary)",
    borderRadius: 1,
  };
  return (
    <div className="flex items-center gap-3">
      <span className="w-9 shrink-0 text-[11px] text-ink-faint">{YAO_NAMES[index]}</span>
      <span className="flex w-24 shrink-0 items-center justify-center" style={{ gap: 8 }}>
        {yang ? (
          <span className="w-full" style={bar} />
        ) : (
          <>
            <span className="flex-1" style={bar} />
            <span className="flex-1" style={bar} />
          </>
        )}
      </span>
      <span className="w-3 text-center text-[11px] leading-none text-gold" aria-hidden>
        {moving ? (yang ? "○" : "×") : ""}
      </span>
      <span className="text-[11px] text-ink-secondary">{tossLabel(backs)}</span>
      <span className="tnum text-[11px] tracking-[0.1em] text-ink-faint">
        {faces.map((b) => (b ? "背" : "字")).join(" ")}
      </span>
    </div>
  );
}

/** 一枚铜钱(方孔圆钱):face true=背(素面),false=字(「通宝」示意点)。 */
function Coin({ face, flipping, delay }: { face: boolean | null; flipping: boolean; delay: number }) {
  return (
    <span
      className={`sw-coin flex h-11 w-11 items-center justify-center rounded-full ${flipping ? "sw-flipping" : ""}`}
      style={{
        animationDelay: `${delay}s`,
        background: "radial-gradient(circle at 35% 30%, var(--gold-glow), transparent 72%)",
        boxShadow: "inset 0 0 0 1.5px var(--gold-dim)",
      }}
      aria-hidden
    >
      {/* 方孔;定面后「字」面四点示意钱文 */}
      <span className="relative flex h-3 w-3 items-center justify-center rounded-[1px]" style={{ boxShadow: "inset 0 0 0 1.5px var(--gold-dim)" }}>
        {face === false && !flipping && (
          <>
            <span className="absolute -top-[7px] h-[3px] w-[3px] rounded-full bg-gold-dim" />
            <span className="absolute -bottom-[7px] h-[3px] w-[3px] rounded-full bg-gold-dim" />
            <span className="absolute -left-[7px] h-[3px] w-[3px] rounded-full bg-gold-dim" />
            <span className="absolute -right-[7px] h-[3px] w-[3px] rounded-full bg-gold-dim" />
          </>
        )}
      </span>
    </span>
  );
}

export function StepShake({
  canThrow,
  disabledHint,
  onComplete,
}: {
  canThrow: boolean; // 所问之事已填写且未在起卦中
  disabledHint?: string;
  onComplete: (tosses: number[]) => void;
}) {
  const [tosses, setTosses] = useState<{ backs: number; faces: boolean[] }[]>([]);
  const [flipping, setFlipping] = useState(false);
  const [pendingFaces, setPendingFaces] = useState<boolean[] | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
  }, []);

  const done = tosses.length >= 6;

  const throwOnce = () => {
    if (flipping || done || !canThrow) return;
    const t = tossCoins();
    if (prefersReducedMotion()) {
      const next = [...tosses, t];
      setTosses(next);
      if (next.length === 6) onComplete(next.map((x) => x.backs));
      return;
    }
    setFlipping(true);
    setPendingFaces(t.faces);
    timer.current = setTimeout(() => {
      setFlipping(false);
      setTosses((prev) => {
        const next = [...prev, t];
        if (next.length === 6) {
          // 稍作停顿再装卦,让第六爻先落定
          timer.current = setTimeout(() => onComplete(next.map((x) => x.backs)), 500);
        }
        return next;
      });
    }, FLIP_MS);
  };

  const reset = () => {
    if (timer.current) clearTimeout(timer.current);
    setFlipping(false);
    setPendingFaces(null);
    setTosses([]);
  };

  const current = flipping ? pendingFaces : tosses.length > 0 ? tosses[tosses.length - 1].faces : null;

  return (
    <div className="mt-4 flex flex-col gap-4">
      <style>{`
        @keyframes sw-flip {
          0%   { transform: translateY(0) rotateY(0deg); }
          40%  { transform: translateY(-18px) rotateY(540deg); }
          80%  { transform: translateY(0) rotateY(900deg); }
          100% { transform: translateY(0) rotateY(1080deg); }
        }
        .sw-flipping { animation: sw-flip ${FLIP_MS / 1000}s var(--ease-inout) both; transform-style: preserve-3d; }
        @media (prefers-reduced-motion: reduce) { .sw-flipping { animation: none; } }
      `}</style>

      <p className="text-[12px] leading-relaxed text-ink-faint">
        意守所问,逐爻掷钱;三背为老阳、无背为老阴,动爻由此而生。
      </p>

      {/* 铜钱 + 掷爻按钮 */}
      <div className="flex flex-wrap items-center gap-5">
        <div className="flex items-center gap-3" style={{ perspective: 420 }} role="img" aria-label="三枚铜钱">
          {[0, 1, 2].map((i) => (
            <Coin key={`${tosses.length}-${flipping ? "f" : "s"}-${i}`} face={current ? current[i] : null} flipping={flipping} delay={i * 0.1} />
          ))}
        </div>
        {!done && (
          <button
            type="button"
            onClick={throwOnce}
            disabled={!canThrow || flipping}
            className="glow-gold inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-6 py-2.5 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-45 disabled:shadow-none"
          >
            {flipping ? "铜钱落定中…" : `掷${YAO_NAMES[tosses.length]}`}
          </button>
        )}
        {done && <span className="text-[13px] text-gold">六爻已齐,装卦中…</span>}
        {tosses.length > 0 && !done && (
          <button
            type="button"
            onClick={reset}
            className="min-h-[44px] rounded-[6px] px-3 text-[12px] text-ink-faint transition-colors hover:text-ink"
          >
            重摇
          </button>
        )}
      </div>
      {!canThrow && disabledHint && <p className="text-[12px] text-ink-faint">{disabledHint}</p>}

      {/* 已落之爻(自下而上,新爻在上方追加则违背卦序——列表倒序展示) */}
      {tosses.length > 0 && (
        <div className="flex flex-col-reverse gap-1.5" aria-label="已落之爻">
          {tosses.map((t, i) => (
            <YaoRow key={i} index={i} backs={t.backs} faces={t.faces} />
          ))}
        </div>
      )}
    </div>
  );
}
