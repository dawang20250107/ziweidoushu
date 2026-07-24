import { redirect } from "next/navigation";

/** 问卦页已拆分为独立板块:梅花易数(/meihua)、六爻(/liuyao)、六壬(/liuren)。
 * 旧链接与书签统一落到梅花板块。 */
export default function DivinationRedirect() {
  redirect("/meihua");
}
