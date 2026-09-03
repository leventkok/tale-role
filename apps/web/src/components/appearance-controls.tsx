"use client";

import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import { THEMES, useTaleTheme } from "@/lib/theme";

export function AppearanceControls({ signOut }: { signOut?: boolean }) {
  const t = useTranslations();
  const locale = useLocale();
  const pathname = usePathname();
  const { theme, setTheme } = useTaleTheme();

  return (
    <div className="appearance-controls">
      <label className="pill">
        {t("ui.theme")}
        <select
          aria-label={t("ui.theme")}
          value={theme}
          onChange={(e) => setTheme(e.target.value as (typeof THEMES)[number])}
        >
          {THEMES.map((id) => (
            <option key={id} value={id}>
              {t(`themes.${id}`)}
            </option>
          ))}
        </select>
      </label>
      <p className="locale-row">
        <Link className="locale" href={pathname} locale="en" aria-current={locale === "en" ? "true" : undefined}>
          EN
        </Link>
        <Link className="locale" href={pathname} locale="tr" aria-current={locale === "tr" ? "true" : undefined}>
          TR
        </Link>
      </p>
      {signOut ? (
        <form action="/api/auth/logout" method="post">
          <button className="ghost" type="submit">
            {t("nav.signOut")}
          </button>
        </form>
      ) : null}
    </div>
  );
}
