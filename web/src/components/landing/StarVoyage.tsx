"use client";

import { useEffect, useRef } from "react";

/**
 * 穿行星海(落地页专属,叠于全局星空之上):
 *   - 三维星场透视投影:恒速缓行给「处于星辰之中」的漂浮感;
 *   - 滚动耦合:下滚即向前飞行,滚得越快星辰越快掠过,
 *     速度过阈值时星点拉成放射光迹(曲速感)——「进入下一页」的穿越瞬间;
 *   - 同时给 <html> 挂 voyage-snap(scroll-snap proximity),一屏一景轻吸附;
 *   - 性能:DPR 封顶 2、星数随面积自适应、标签页隐藏即停;
 *   - 可及性:prefers-reduced-motion 不渲染(退回全局静态星空);
 *   - 主题:宣纸(light)不渲染,监听 data-theme 即时启停。
 */

interface VStar {
  x: number; // 视空间 [-1,1](以消失点为原点)
  y: number;
  z: number; // 深度 (0,1],越小越近
  gold: boolean;
  px: number; // 上一帧投影位置(拉光迹用)
  py: number;
}

const BASE_SPEED = 0.045; // 恒速前行(z/秒)
const SCROLL_GAIN = 0.0011; // 滚动速度 → 前行速度增益
const MAX_SPEED = 1.6;
const STREAK_FROM = 0.18; // 速度过此阈值开始拉光迹

function spawn(rand: () => number, z: number): VStar {
  return {
    x: (rand() * 2 - 1) * 1.6,
    y: (rand() * 2 - 1) * 1.2,
    z,
    gold: rand() < 0.14,
    px: NaN,
    py: NaN,
  };
}

export function StarVoyage() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    // 一屏一景轻吸附(proximity 不劫持,自由滚动仍顺畅)
    document.documentElement.classList.add("voyage-snap");

    let stars: VStar[] = [];
    let raf = 0;
    let running = false;
    let last = performance.now();
    let speed = BASE_SPEED;
    let lastScrollY = window.scrollY;
    let scrollVel = 0; // 平滑后的滚动速度(px/s,带方向)

    const isDark = () => document.documentElement.getAttribute("data-theme") !== "light";

    function resize() {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas!.width = Math.round(window.innerWidth * dpr);
      canvas!.height = Math.round(window.innerHeight * dpr);
      ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
      const count = Math.min(260, Math.round((window.innerWidth * window.innerHeight) / 6200));
      stars = Array.from({ length: count }, () => spawn(Math.random, 0.08 + Math.random() * 0.92));
    }

    function frame(now: number) {
      const w = window.innerWidth;
      const h = window.innerHeight;
      const dt = Math.min((now - last) / 1000, 0.1);
      last = now;

      // 滚动速度平滑(指数衰减),驱动前行速度
      const sy = window.scrollY;
      const instVel = dt > 0 ? (sy - lastScrollY) / dt : 0;
      lastScrollY = sy;
      scrollVel += (instVel - scrollVel) * Math.min(1, dt * 9);
      const target = BASE_SPEED + scrollVel * SCROLL_GAIN; // 上滚略倒退,下滚加速
      speed += (Math.max(-0.5, Math.min(MAX_SPEED, target)) - speed) * Math.min(1, dt * 7);

      ctx!.clearRect(0, 0, w, h);
      const cx = w / 2;
      const cy = h * 0.46; // 消失点略高于屏心,与极光/构图一致
      const f = Math.min(w, h) * 0.5;
      const streak = Math.abs(speed) > STREAK_FROM;

      for (const s of stars) {
        s.z -= speed * dt;
        if (s.z <= 0.045 || s.z > 1.02) {
          // 穿过身旁(或倒退出景)即在远处重生
          const ns = spawn(Math.random, speed >= 0 ? 0.96 + Math.random() * 0.05 : 0.06);
          s.x = ns.x; s.y = ns.y; s.z = ns.z; s.gold = ns.gold;
          s.px = NaN; s.py = NaN;
        }
        const sx = cx + (s.x / s.z) * f * 0.42;
        const syp = cy + (s.y / s.z) * f * 0.42;
        if (sx < -40 || sx > w + 40 || syp < -40 || syp > h + 40) {
          s.px = sx; s.py = syp;
          continue;
        }
        const near = 1 - s.z; // 越近越大越亮
        const r = 0.35 + near * 1.75;
        const alpha = 0.14 + near * 0.6;
        const color = s.gold ? "217, 179, 108" : "205, 214, 240";

        if (streak && Number.isFinite(s.px)) {
          // 曲速光迹:上一帧投影 → 当前投影
          const grad = ctx!.createLinearGradient(s.px, s.py, sx, syp);
          grad.addColorStop(0, `rgba(${color}, 0)`);
          grad.addColorStop(1, `rgba(${color}, ${Math.min(0.85, alpha + 0.2)})`);
          ctx!.strokeStyle = grad;
          ctx!.lineWidth = Math.max(0.6, r * 0.9);
          ctx!.beginPath();
          ctx!.moveTo(s.px, s.py);
          ctx!.lineTo(sx, syp);
          ctx!.stroke();
        } else {
          ctx!.globalAlpha = alpha;
          ctx!.fillStyle = s.gold ? "#d9b36c" : "#cdd6f0";
          ctx!.beginPath();
          ctx!.arc(sx, syp, r, 0, Math.PI * 2);
          ctx!.fill();
          ctx!.globalAlpha = 1;
        }
        s.px = sx;
        s.py = syp;
      }
      raf = requestAnimationFrame(frame);
    }

    function start() {
      if (running || !isDark() || document.hidden) return;
      running = true;
      last = performance.now();
      lastScrollY = window.scrollY;
      raf = requestAnimationFrame(frame);
    }
    function stop(clear = false) {
      running = false;
      cancelAnimationFrame(raf);
      if (clear) ctx!.clearRect(0, 0, window.innerWidth, window.innerHeight);
    }

    resize();
    start();

    const onResize = () => resize();
    const onVisibility = () => (document.hidden ? stop() : start());
    const mo = new MutationObserver(() => (isDark() ? start() : stop(true)));
    mo.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
    window.addEventListener("resize", onResize);
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      stop();
      mo.disconnect();
      window.removeEventListener("resize", onResize);
      document.removeEventListener("visibilitychange", onVisibility);
      document.documentElement.classList.remove("voyage-snap");
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      aria-hidden
      className="pointer-events-none fixed inset-0 -z-10"
    />
  );
}
