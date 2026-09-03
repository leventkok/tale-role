"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { desktopDownloadUrl, type DesktopPlatform } from "@/lib/desktop-download";

function guessPlatform(): DesktopPlatform {
  if (typeof navigator === "undefined") {
    return "win";
  }
  const ua = navigator.userAgent.toLowerCase();
  if (/mac|iphone|ipad|ipod/.test(ua) && !/windows/.test(ua)) {
    return "mac";
  }
  return "win";
}

export function DesktopDownload() {
  const t = useTranslations("home");
  const [platform, setPlatform] = useState<DesktopPlatform>("win");

  useEffect(() => {
    setPlatform(guessPlatform());
  }, []);

  const href = desktopDownloadUrl(platform);

  return (
    <div className="desktop-download">
      <p className="muted">{t("downloadLead")}</p>
      <div className="platform-pick" role="group" aria-label={t("downloadPlatform")}>
        <button
          type="button"
          className={platform === "win" ? "platform on" : "platform"}
          aria-pressed={platform === "win"}
          onClick={() => setPlatform("win")}
        >
          {t("downloadWindows")}
        </button>
        <button
          type="button"
          className={platform === "mac" ? "platform on" : "platform"}
          aria-pressed={platform === "mac"}
          onClick={() => setPlatform("mac")}
        >
          {t("downloadMac")}
        </button>
      </div>
      <a className="btn" href={href} download rel="noopener">
        {platform === "win" ? t("downloadCtaWin") : t("downloadCtaMac")}
      </a>
    </div>
  );
}
