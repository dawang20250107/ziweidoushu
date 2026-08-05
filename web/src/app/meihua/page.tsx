import type { Metadata } from "next";
import { DivinationWorkbench } from "@/components/divination/Workbench";

export const metadata: Metadata = {
  title: "梅花易数 · 观星台",
  description: "邵康节心易:时间/报数起卦,体用生克为纲、卦气旺衰定力度,互卦断过程、变卦断结局,AI 深度解卦。",
};

/** 梅花易数独立板块(自问卦页拆分,先攻纵深)。 */
export default function MeihuaPage() {
  return <DivinationWorkbench kind="meihua" homePath="/meihua" />;
}
