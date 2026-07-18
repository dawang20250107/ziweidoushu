"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";

type Theme = "dark" | "light";

const ThemeContext = createContext<{
  theme: Theme;
  setTheme: (t: Theme) => void;
}>({ theme: "dark", setTheme: () => {} });

const STORAGE_KEY = "ziwei-theme";

/** 主题控制(v2):玄穹为产品默认(无 data-theme 即深空),宣纸经 data-theme="light" 显式启用。 */
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>("dark");

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === "light") {
      setThemeState("light");
      document.documentElement.setAttribute("data-theme", "light");
    }
  }, []);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    localStorage.setItem(STORAGE_KEY, t);
    if (t === "light") {
      document.documentElement.setAttribute("data-theme", "light");
    } else {
      document.documentElement.removeAttribute("data-theme");
    }
  }, []);

  return <ThemeContext.Provider value={{ theme, setTheme }}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  return useContext(ThemeContext);
}

/** 无闪烁初始化脚本:SSR HTML 里内联执行,先于首帧应用存储的主题。 */
export const themeInitScript = `(function(){try{if(localStorage.getItem("${STORAGE_KEY}")==="light"){document.documentElement.setAttribute("data-theme","light")}}catch(e){}})()`;

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const dark = theme === "dark";
  return (
    <button
      type="button"
      onClick={() => setTheme(dark ? "light" : "dark")}
      className="rounded-[2px] border border-line-strong px-3 py-1 text-[13px] text-ink-secondary transition-colors hover:border-gold-dim hover:text-gold"
      title={dark ? "切换到宣纸(浅色阅读)" : "切换到玄穹(深空)"}
    >
      {dark ? "玄穹" : "宣纸"}
    </button>
  );
}
