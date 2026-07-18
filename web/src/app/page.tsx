import { Hero } from "@/components/landing/Hero";
import { Capabilities } from "@/components/landing/Capabilities";
import { TrustRow } from "@/components/landing/TrustRow";
import { ClosingCta } from "@/components/landing/ClosingCta";

/** 观星台落地页:Hero(真实迷你星盘)→ 三大能力 → 信任事实 → 尾部 CTA。 */
export default function HomePage() {
  return (
    <>
      <Hero />
      <Capabilities />
      <TrustRow />
      <ClosingCta />
    </>
  );
}
