"use client";

import { useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";

const THEMES = [
  "high-fantasy",
  "gothic-horror",
  "space-opera",
  "cyber-noir",
  "post-apocalyptic",
  "fairytale",
] as const;

type Theme = (typeof THEMES)[number];

function isTheme(v: string): v is Theme {
  return (THEMES as readonly string[]).includes(v);
}

export function SiteHeader({ email }: { email: string | null }) {
  const t = useTranslations();
  const locale = useLocale();
  const pathname = usePathname();
  const [theme, setTheme] = useState<Theme>("high-fantasy");

  useEffect(() => {
    const saved = window.localStorage.getItem("talerole-theme");
    if (saved && isTheme(saved)) {
      setTheme(saved);
      document.documentElement.dataset.theme = saved;
    } else {
      document.documentElement.dataset.theme = "high-fantasy";
    }
  }, []);

  function onTheme(next: Theme) {
    setTheme(next);
    document.documentElement.dataset.theme = next;
    window.localStorage.setItem("talerole-theme", next);
  }

  return (
    <header className="topbar">
      <Link href="/" className="brand">
        <strong>{t("app.name")}</strong>
        <span>{t("app.tagline")}</span>
      </Link>
      <nav>
        {email ? (
          <>
            <Link href="/host">{t("nav.host")}</Link>
            <Link href="/universe/new">{t("nav.universe")}</Link>
            <Link href="/play">{t("nav.play")}</Link>
          </>
        ) : (
          <>
            <Link href="/login">{t("nav.signIn")}</Link>
            <Link href="/register">{t("nav.register")}</Link>
          </>
        )}
      </nav>
      <div className="topbar-tools">
        <label className="pill">
          {t("ui.theme")}
          <select
            aria-label={t("ui.theme")}
            value={theme}
            onChange={(e) => onTheme(e.target.value as Theme)}
          >
            {THEMES.map((id) => (
              <option key={id} value={id}>
                {t(`themes.${id}`)}
              </option>
            ))}
          </select>
        </label>
        <Link className="locale" href={pathname} locale="en" aria-current={locale === "en" ? "true" : undefined}>
          EN
        </Link>
        <Link className="locale" href={pathname} locale="tr" aria-current={locale === "tr" ? "true" : undefined}>
          TR
        </Link>
        {email ? (
          <form action="/api/auth/logout" method="post">
            <button className="ghost" type="submit">
              {t("nav.signOut")}
            </button>
          </form>
        ) : null}
      </div>
    </header>
  );
}
