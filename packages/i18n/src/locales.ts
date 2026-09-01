export const locales = ["en", "tr"] as const;

export type Locale = (typeof locales)[number];

/** Product default. User preference and then browser language override this. */
export const defaultLocale: Locale = "en";

export function isLocale(value: string | undefined | null): value is Locale {
  return value === "en" || value === "tr";
}

/**
 * Resolve UI locale.
 * Order: explicit user preference → browser/OS tag → default `en`.
 */
export function resolveLocale(
  preferred?: string | null,
  browser?: string | null,
): Locale {
  if (isLocale(preferred)) {
    return preferred;
  }
  const fromBrowser = normalizeTag(browser);
  if (fromBrowser) {
    return fromBrowser;
  }
  return defaultLocale;
}

function normalizeTag(tag?: string | null): Locale | undefined {
  if (!tag) {
    return undefined;
  }
  const primary = tag.trim().toLowerCase().split(/[-_]/)[0];
  return isLocale(primary) ? primary : undefined;
}
