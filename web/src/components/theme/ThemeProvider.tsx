"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";

type Theme = "dark" | "light" | "system";

const ThemeContext = createContext<{
  theme: Theme;
  setTheme: (t: Theme) => void;
}>({ theme: "system", setTheme: () => {} });

const STORAGE_KEY = "ziwei-theme";

/** 主题控制:system 跟随系统;dark/light 以 data-theme 强制(token 级覆盖)。 */
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>("system");

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY) as Theme | null;
    if (saved === "dark" || saved === "light") {
      setThemeState(saved);
      document.documentElement.setAttribute("data-theme", saved);
    }
  }, []);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    if (t === "system") {
      localStorage.removeItem(STORAGE_KEY);
      document.documentElement.removeAttribute("data-theme");
    } else {
      localStorage.setItem(STORAGE_KEY, t);
      document.documentElement.setAttribute("data-theme", t);
    }
  }, []);

  return <ThemeContext.Provider value={{ theme, setTheme }}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  return useContext(ThemeContext);
}

/** 无闪烁初始化脚本:SSR HTML 里内联执行,先于首帧应用存储的主题。 */
export const themeInitScript = `(function(){try{var t=localStorage.getItem("${STORAGE_KEY}");if(t==="dark"||t==="light"){document.documentElement.setAttribute("data-theme",t)}}catch(e){}})()`;

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const next = theme === "dark" ? "light" : theme === "light" ? "system" : "dark";
  const label = theme === "dark" ? "玄穹" : theme === "light" ? "宣纸" : "跟随系统";
  return (
    <button
      type="button"
      onClick={() => setTheme(next)}
      className="rounded-[2px] border border-line-strong px-3 py-1 text-[13px] text-ink-secondary transition-colors hover:border-gold-dim hover:text-gold"
      title="切换主题(玄穹 / 宣纸 / 跟随系统)"
    >
      {label}
    </button>
  );
}
