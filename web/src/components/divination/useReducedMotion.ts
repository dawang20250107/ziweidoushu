"use client";

import { useEffect, useState } from "react";

/**
 * 是否偏好减少动效。SSR 首帧按「不减少」渲染,挂载后依系统偏好校正。
 * 需要在事件回调里同步判断时,直接读 window.matchMedia 而非此 hook。
 */
export function useReducedMotion(): boolean {
  const [reduced, setReduced] = useState(false);
  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    const sync = () => setReduced(mq.matches);
    sync();
    mq.addEventListener("change", sync);
    return () => mq.removeEventListener("change", sync);
  }, []);
  return reduced;
}

/** 事件回调内的同步判断(起卦计时用)。 */
export function prefersReducedMotion(): boolean {
  return typeof window !== "undefined" && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}
