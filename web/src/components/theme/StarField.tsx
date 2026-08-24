"use client";

import { useEffect, useRef } from "react";

/**
 * 全局动态星空(玄穹主题的呼吸底):
 *   - 双层视差:远层细星缓慢漂移,近层亮星呼吸闪烁;偶发流星划过;
 *   - 性能:devicePixelRatio 封顶 2、星数随面积自适应、标签页隐藏即暂停;
 *   - 可及性:prefers-reduced-motion 时只静态绘制一帧,无任何动画;
 *   - 主题:宣纸(light)下不渲染,监听 data-theme 切换即时启停。
 * 固定于视口底层(z-index:-1),不参与命中测试,滚动零成本。
 */

interface Star {
  x: number; // 0-1 归一化
  y: number;
  r: number;
  base: number; // 基础亮度 0-1
  phase: number; // 闪烁相位
  speed: number; // 闪烁速率
  drift: number; // 水平漂移速率(归一化/秒)
  gold: boolean; // 少数星曜着金
}

interface Meteor {
  x: number;
  y: number;
  vx: number;
  vy: number;
  life: number; // 剩余寿命(秒)
  total: number;
}

function makeStars(count: number, rand: () => number): Star[] {
  return Array.from({ length: count }, (_, i) => {
    const far = i < count * 0.62; // 远层占六成
    return {
      x: rand(),
      y: rand(),
      r: far ? 0.4 + rand() * 0.5 : 0.8 + rand() * 1.0,
      base: far ? 0.18 + rand() * 0.3 : 0.35 + rand() * 0.5,
      phase: rand() * Math.PI * 2,
      speed: 0.3 + rand() * 0.9,
      drift: (far ? 0.0016 : 0.004) * (0.5 + rand()),
      gold: rand() < 0.12,
    };
  });
}

export function StarField() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    let stars: Star[] = [];
    let meteor: Meteor | null = null;
    let nextMeteorAt = 0;
    let raf = 0;
    let running = false;
    let last = performance.now();

    const isDark = () => document.documentElement.getAttribute("data-theme") !== "light";

    function resize() {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas!.width = Math.round(window.innerWidth * dpr);
      canvas!.height = Math.round(window.innerHeight * dpr);
      ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
      const count = Math.min(170, Math.round((window.innerWidth * window.innerHeight) / 11000));
      stars = makeStars(count, Math.random);
    }

    function draw(now: number) {
      const w = window.innerWidth;
      const h = window.innerHeight;
      const dt = Math.min((now - last) / 1000, 0.1);
      last = now;
      ctx!.clearRect(0, 0, w, h);

      // 滚动视差:远层随滚动轻移、近层稍多,全站滚动即有星空纵深
      const sy = reduced ? 0 : window.scrollY;
      const t = now / 1000;
      for (const s of stars) {
        if (!reduced) {
          s.x += s.drift * dt;
          if (s.x > 1.002) s.x = -0.002;
        }
        const far = s.r < 0.9;
        const py = (((s.y * h - sy * (far ? 0.05 : 0.11)) % (h + 8)) + h + 8) % (h + 8) - 4;
        const tw = reduced ? 1 : 0.72 + 0.28 * Math.sin(t * s.speed + s.phase);
        ctx!.globalAlpha = s.base * tw;
        ctx!.fillStyle = s.gold ? "#d9b36c" : "#cdd6f0";
        ctx!.beginPath();
        ctx!.arc(s.x * w, py, s.r, 0, Math.PI * 2);
        ctx!.fill();
      }

      // 流星:低频点缀(约 9-17s 一颗)
      if (!reduced) {
        if (!meteor && now >= nextMeteorAt) {
          const fromX = 0.15 + Math.random() * 0.7;
          meteor = {
            x: fromX * w,
            y: -20,
            vx: (0.25 + Math.random() * 0.2) * w,
            vy: (0.35 + Math.random() * 0.15) * h,
            life: 0.9,
            total: 0.9,
          };
          nextMeteorAt = now + 9000 + Math.random() * 8000;
        }
        if (meteor) {
          meteor.life -= dt;
          if (meteor.life <= 0) {
            meteor = null;
          } else {
            meteor.x += meteor.vx * dt;
            meteor.y += meteor.vy * dt;
            const k = meteor.life / meteor.total;
            const tail = 90;
            const nx = meteor.vx / Math.hypot(meteor.vx, meteor.vy);
            const ny = meteor.vy / Math.hypot(meteor.vx, meteor.vy);
            const grad = ctx!.createLinearGradient(
              meteor.x, meteor.y, meteor.x - nx * tail, meteor.y - ny * tail,
            );
            grad.addColorStop(0, `rgba(242, 206, 133, ${0.85 * k})`);
            grad.addColorStop(1, "rgba(242, 206, 133, 0)");
            ctx!.globalAlpha = 1;
            ctx!.strokeStyle = grad;
            ctx!.lineWidth = 1.4;
            ctx!.beginPath();
            ctx!.moveTo(meteor.x, meteor.y);
            ctx!.lineTo(meteor.x - nx * tail, meteor.y - ny * tail);
            ctx!.stroke();
          }
        }
      }
      ctx!.globalAlpha = 1;
    }

    function loop(now: number) {
      draw(now);
      if (reduced) {
        running = false;
        return; // 静态一帧即止
      }
      raf = requestAnimationFrame(loop);
    }

    function start() {
      if (running || !isDark() || document.hidden) return;
      running = true;
      last = performance.now();
      raf = requestAnimationFrame(loop);
    }
    function stop(clear = false) {
      running = false;
      cancelAnimationFrame(raf);
      if (clear) ctx!.clearRect(0, 0, window.innerWidth, window.innerHeight);
    }

    resize();
    start();

    const onResize = () => {
      resize();
      if (reduced && isDark()) draw(performance.now());
    };
    const onVisibility = () => (document.hidden ? stop() : start());
    // data-theme 切换即时启停(宣纸不渲染星空)
    const mo = new MutationObserver(() => {
      if (isDark()) {
        if (reduced) draw(performance.now());
        else start();
      } else {
        stop(true);
      }
    });
    mo.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
    window.addEventListener("resize", onResize);
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      stop();
      mo.disconnect();
      window.removeEventListener("resize", onResize);
      document.removeEventListener("visibilitychange", onVisibility);
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
