"use client";

import { useEffect, useState } from "react";

export const THEMES = [
  "high-fantasy",
  "gothic-horror",
  "space-opera",
  "cyber-noir",
  "post-apocalyptic",
  "fairytale",
] as const;

export type Theme = (typeof THEMES)[number];

const THEME_KEY = "talerole-theme";
const THEME_EVENT = "talerole-theme";

export function isTheme(v: string): v is Theme {
  return (THEMES as readonly string[]).includes(v);
}

export function applyTheme(id: Theme) {
  document.documentElement.dataset.theme = id;
}

export function useTaleTheme() {
  const [theme, setThemeState] = useState<Theme>("high-fantasy");

  useEffect(() => {
    function read() {
      const saved = window.localStorage.getItem(THEME_KEY);
      const next = saved && isTheme(saved) ? saved : "high-fantasy";
      setThemeState(next);
      applyTheme(next);
    }
    read();
    window.addEventListener(THEME_EVENT, read);
    window.addEventListener("storage", read);
    return () => {
      window.removeEventListener(THEME_EVENT, read);
      window.removeEventListener("storage", read);
    };
  }, []);

  function setTheme(next: Theme) {
    setThemeState(next);
    applyTheme(next);
    window.localStorage.setItem(THEME_KEY, next);
    window.dispatchEvent(new Event(THEME_EVENT));
  }

  return { theme, setTheme };
}
