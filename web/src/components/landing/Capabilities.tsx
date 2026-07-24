import type { ReactNode } from "react";

/* ────────────────────────────────────────────────────────────
 * 例 1:静态宫格样例(排盘块)——含四化徽章,复刻 PalaceCell 形制
 * ──────────────────────────────────────────────────────────── */

function SihuaBadge({ mark, color }: { mark: string; color: string }) {
  return (
    <sup
      className="ml-px rounded-[2px] px-[3px] text-[10px] font-semibold not-italic"
      style={{ background: color, color: "var(--bg)" }}
    >
      {mark}
    </sup>
  );
}

function MajorStar({ name, brightness, badge }: { name: string; brightness: string; badge?: ReactNode }) {
  return (
    <span className="font-display text-[17px] font-semibold leading-tight" style={{ color: brightness }}>
      {name}
      {badge}
    </span>
  );
}

function PalaceSample() {
  return (
    <div className="mx-auto w-full max-w-[240px]">
      <div
        className="relative flex min-h-[168px] flex-col rounded-[6px] bg-bg-raised p-3"
        style={{
          boxShadow: "inset 0 0 0 1px var(--gold-dim)",
          backgroundImage: "linear-gradient(var(--gold-glow), transparent 60%)",
        }}
      >
        <span className="absolute right-2 top-2 rounded-[2px] border border-gold-dim px-1 text-[10px] leading-4 text-gold">
          身
        </span>

        {/* 主星行 */}
        <div className="flex flex-wrap gap-x-2.5 gap-y-0.5">
          <MajorStar
            name="紫微"
            brightness="var(--brightness-miao)"
            badge={<SihuaBadge mark="权" color="var(--sihua-quan)" />}
          />
          <MajorStar name="天相" brightness="var(--brightness-mid)" />
        </div>

        {/* 辅星行 */}
        <div className="mt-1 flex flex-wrap gap-x-2 gap-y-0.5">
          <span className="text-[13px] leading-tight" style={{ color: "var(--brightness-mid)" }}>
            文曲
            <SihuaBadge mark="忌" color="var(--sihua-ji)" />
          </span>
          <span className="text-[13px] leading-tight" style={{ color: "var(--brightness-mid)" }}>
            右弼
          </span>
        </div>

        <div className="flex-1" />

        {/* 底栏:宫名 · 干支 · 大限 */}
        <div className="border-t border-line pt-1.5">
          <div className="flex items-baseline justify-between gap-1">
            <span className="font-display text-[13px] font-semibold text-ink">命宫</span>
            <span className="text-[11px] text-ink-secondary">
              甲寅
              <span className="tnum ml-1 text-ink-faint">6 - 15</span>
            </span>
          </div>
        </div>
      </div>
      <p className="mt-2.5 text-center text-[12px] text-ink-faint">
        庙旺利陷按亮度上色,四化以徽章挂于星侧
      </p>
    </div>
  );
}

/* ────────────────────────────────────────────────────────────
 * 例 2:古籍引文卡(古籍块)——《骨髓赋》,font-reading
 * ──────────────────────────────────────────────────────────── */

function ClassicQuote() {
  return (
    <div
      className="w-full max-w-[420px] rounded-[10px] bg-bg-raised p-5"
      style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}
    >
      <div className="mb-4 flex items-center gap-2">
        <span className="flex-1 rounded-[6px] px-3 py-1.5 text-[13px] text-ink-secondary" style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}>
          搜「紫微」
        </span>
        <span className="rounded-[6px] px-2.5 py-1.5 text-[12px] text-ink-faint" style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}>
          全文检索
        </span>
      </div>

      <p className="text-[12px] tracking-[0.08em] text-ink-faint">《紫微斗数全书 · 骨髓赋》</p>
      <blockquote className="mt-2 font-reading text-[22px] leading-[1.9] text-ink">
        <mark>紫微</mark>帝座,为众星之尊;天府令星,为万宫之主。
      </blockquote>
      <p className="mt-3 font-reading text-[15px] leading-[1.9] text-ink-secondary">
        紫微如帝王坐镇中天,统御群曜;天府似府库之主,主一身财禄。二者所临,格局自高。
      </p>
    </div>
  );
}

/* ────────────────────────────────────────────────────────────
 * 例 3:对话气泡(AI 块)
 * ──────────────────────────────────────────────────────────── */

function ChatSample() {
  return (
    <div className="w-full max-w-[420px] space-y-3">
      {/* 用户 */}
      <div className="flex justify-end">
        <div
          className="max-w-[80%] rounded-[10px] rounded-tr-[2px] px-3.5 py-2 text-[14px] leading-relaxed text-ink"
          style={{ background: "var(--gold-glow)", boxShadow: "inset 0 0 0 1px var(--gold-dim)" }}
        >
          紫微坐命,事业适合守成还是开创?
        </div>
      </div>

      {/* AI */}
      <div className="flex justify-start">
        <div
          className="max-w-[85%] rounded-[10px] rounded-tl-[2px] bg-bg-raised px-3.5 py-2.5"
          style={{ boxShadow: "inset 0 0 0 1px var(--line)" }}
        >
          <p className="text-[14px] leading-relaxed text-ink">
            命宫紫微化权、会天相,主见强而能担事,宜居主导之位；但三方文曲化忌,决策易受口舌牵动,开创之余须留一手核账。
          </p>
          <div className="mt-2 flex items-center gap-1.5 border-t border-line pt-2 text-[11px] text-ink-faint">
            <span className="inline-block size-1 rounded-full" style={{ background: "var(--gold)" }} />
            依据 · 命宫紫微·权 / 三方文曲·忌
          </div>
        </div>
      </div>
    </div>
  );
}

/* ────────────────────────────────────────────────────────────
 * 能力区通用行:左文右例,交替左右
 * ──────────────────────────────────────────────────────────── */

interface CapabilityRowProps {
  eyebrow: string;
  title: string;
  body: string;
  example: ReactNode;
  flip?: boolean;
}

function CapabilityRow({ eyebrow, title, body, example, flip = false }: CapabilityRowProps) {
  return (
    <div className="reveal grid items-center gap-12 lg:grid-cols-2 lg:gap-20">
      <div className={flip ? "lg:order-2" : ""}>
        <p className="text-[12px] font-medium tracking-[0.08em] text-gold">{eyebrow}</p>
        <h2 className="mt-4 text-balance font-display text-[26px] font-semibold leading-snug text-ink md:text-[31px]">
          {title}
        </h2>
        <p className="mt-5 max-w-md text-[15px] leading-[1.75] text-ink-secondary">{body}</p>
      </div>
      <div className={["flex justify-center", flip ? "lg:order-1 lg:justify-start" : "lg:justify-end"].join(" ")}>
        {example}
      </div>
    </div>
  );
}

/** 三大能力区:排盘做透 / 古籍丝滑 / AI 问星。 */
export function Capabilities() {
  return (
    <section className="mx-auto max-w-6xl px-4 py-16 md:py-24">
      <div className="space-y-24 md:space-y-32">
        <CapabilityRow
          eyebrow="排盘"
          title="从安星到运限,一张盘全交代"
          body="十二宫、四化、三方四正一次算全;星曜按庙旺利陷上色,强弱一眼可读。格局与大限流年逐层展开,每一步都对齐经典口径。"
          example={<PalaceSample />}
        />
        <CapabilityRow
          flip
          eyebrow="古籍"
          title="古籍全文,查得到也读得下"
          body="《紫微斗数全书》等经典逐句录入,全文检索命中即达。正文以楷体排版、注译随文对照,古人的话就在纸上,不必再东翻西找。"
          example={<ClassicQuote />}
        />
        <CapabilityRow
          eyebrow="问星"
          title="解读依盘而言,不作空谈"
          body="把你的命盘交给 AI,它引盘中的星曜与格局作答,每一句都标注依据、可回到盘上核对。要的是有出处的判断,而非泛泛之词。"
          example={<ChatSample />}
        />
      </div>
    </section>
  );
}
