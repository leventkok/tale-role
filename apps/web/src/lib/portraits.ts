export const portraitIds = ["warden", "blade", "sage", "ranger", "oath"] as const;

export type PortraitId = (typeof portraitIds)[number];

export const defaultPortraitId: PortraitId = "warden";

export const PORTRAIT_KEY = "talerole-portrait";
export const PORTRAIT_EVENT = "talerole-portrait";

export function isPortraitId(v: string): v is PortraitId {
  return (portraitIds as readonly string[]).includes(v);
}

export function normalizePortraitId(v: string | null | undefined): PortraitId {
  return v && isPortraitId(v) ? v : defaultPortraitId;
}

export function portraitFromSeed(seed: string): PortraitId {
  let h = 0;
  for (const ch of seed) {
    h = (h * 33 + ch.charCodeAt(0)) >>> 0;
  }
  return portraitIds[h % portraitIds.length];
}

export function readStoredPortrait(): PortraitId {
  if (typeof window === "undefined") {
    return defaultPortraitId;
  }
  return normalizePortraitId(window.localStorage.getItem(PORTRAIT_KEY));
}

export function writeStoredPortrait(id: PortraitId) {
  window.localStorage.setItem(PORTRAIT_KEY, id);
  window.dispatchEvent(new Event(PORTRAIT_EVENT));
}
