"use client";

import { useEffect, useState, type CSSProperties } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { BrandMark } from "@/components/art/brand-mark";
import { ProfilePortrait } from "@/components/art/profile-portrait";
import { AppearanceControls } from "@/components/appearance-controls";
import { gql, gqlData } from "@/lib/gql";
import {
  PORTRAIT_EVENT,
  defaultPortraitId,
  normalizePortraitId,
  readStoredPortrait,
  type PortraitId,
} from "@/lib/portraits";

function displayName(email: string) {
  return email.split("@")[0] || email;
}

export function SiteHeader({
  email,
  lanternXp = 0,
  lanternLevel = 1,
}: {
  email: string | null;
  lanternXp?: number;
  lanternLevel?: number;
}) {
  const t = useTranslations();
  const [xp, setXp] = useState(lanternXp);
  const [level, setLevel] = useState(lanternLevel < 1 ? 1 : lanternLevel);
  const [portrait, setPortrait] = useState<PortraitId>(defaultPortraitId);

  useEffect(() => {
    setXp(lanternXp);
    setLevel(lanternLevel < 1 ? 1 : lanternLevel);
  }, [lanternXp, lanternLevel]);

  useEffect(() => {
    function syncPortrait() {
      setPortrait(readStoredPortrait());
    }
    syncPortrait();
    window.addEventListener(PORTRAIT_EVENT, syncPortrait);
    window.addEventListener("storage", syncPortrait);
    return () => {
      window.removeEventListener(PORTRAIT_EVENT, syncPortrait);
      window.removeEventListener("storage", syncPortrait);
    };
  }, []);

  useEffect(() => {
    if (!email) {
      return;
    }
    void gql<{ me: { lanternXp?: number; lanternLevel?: number; portraitId?: string } | null }>(
      `{ me { lanternXp lanternLevel portraitId } }`,
    ).then((result) => {
      const me = gqlData(result)?.me;
      if (!me) {
        return;
      }
      setXp(me.lanternXp ?? 0);
      setLevel(me.lanternLevel && me.lanternLevel > 0 ? me.lanternLevel : 1);
      if (me.portraitId) {
        setPortrait(normalizePortraitId(me.portraitId));
      }
    });
  }, [email]);

  const need = 100 * level;
  const pct = Math.max(4, Math.min(100, Math.round((xp / need) * 100)));
  const name = email ? displayName(email) : "";

  return (
    <header className="topbar">
      <Link href="/" className="brand">
        <BrandMark />
        <strong>{t("app.name")}</strong>
        <span>{t("app.tagline")}</span>
      </Link>
      {email ? null : (
        <nav>
          <Link href="/login">{t("nav.signIn")}</Link>
          <Link href="/register">{t("nav.register")}</Link>
        </nav>
      )}
      <div className="topbar-tools">
        {email ? (
          <Link href="/account" className="nav-profile" aria-label={t("nav.openAccount", { name })}>
            <ProfilePortrait id={portrait} className="nav-portrait" />
            <span className="nav-profile-copy">
              <strong>{name}</strong>
              <span className="muted">{t("table.level", { n: level })}</span>
              <span className="nav-profile-xp">
                <span className="hp-bar xp">
                  <span style={{ ["--hp"]: `${pct}%` } as CSSProperties} />
                </span>
              </span>
            </span>
          </Link>
        ) : (
          <AppearanceControls />
        )}
      </div>
    </header>
  );
}
