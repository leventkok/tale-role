"use client";

import { useLayoutEffect } from "react";
import { usePathname } from "next/navigation";
import { isDesktopShell, markDesktopShell } from "@/lib/desktop-shell";

/** Keeps the Electron shell flag on <html> after Next.js reconciliation (Win + Mac). */
export function DesktopShellMark() {
  const pathname = usePathname();

  useLayoutEffect(() => {
    if (!isDesktopShell()) {
      return;
    }
    markDesktopShell();
    const root = document.documentElement;
    const observer = new MutationObserver(() => {
      if (root.dataset.taleroleShell !== "desktop") {
        markDesktopShell();
      }
    });
    observer.observe(root, { attributes: true, attributeFilter: ["data-talerole-shell"] });
    return () => observer.disconnect();
  }, [pathname]);

  return null;
}
