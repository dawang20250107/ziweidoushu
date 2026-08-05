import { Hero } from "@/components/landing/Hero";
import { Capabilities } from "@/components/landing/Capabilities";
import { TrustRow } from "@/components/landing/TrustRow";
import { ClosingCta } from "@/components/landing/ClosingCta";
import { StarVoyage } from "@/components/landing/StarVoyage";

/**
 * 观星台落地页 · 穿行星海:一屏一景(scroll-snap 轻吸附),
 * 滚动即向前飞行——星海画布随滚动加速、场景自深处驶来/自身旁掠过。
 * Hero(真实迷你星盘)→ 三大能力(各一景)→ 信任事实 + 尾部 CTA(终景)。
 */
export default function HomePage() {
  return (
    <>
      <StarVoyage />
      <Hero />
      <Capabilities />
      <section className="scene flex min-h-[92svh] flex-col justify-center">
        <div className="scene-body">
          <TrustRow />
          <ClosingCta />
        </div>
      </section>
    </>
  );
}
