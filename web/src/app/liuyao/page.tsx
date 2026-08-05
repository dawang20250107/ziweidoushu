import type { Metadata } from "next";
import { DivinationWorkbench } from "@/components/divination/Workbench";

export const metadata: Metadata = {
  title: "六爻纳甲 · 观星台",
  description: "火珠林法:铜钱摇卦装卦,用神旺衰、动变生克、世应应期,依《增删卜易》断法 AI 深度解卦。",
};

/** 六爻纳甲独立板块(自问卦页拆分)。 */
export default function LiuYaoPage() {
  return <DivinationWorkbench kind="liuyao" homePath="/liuyao" />;
}
