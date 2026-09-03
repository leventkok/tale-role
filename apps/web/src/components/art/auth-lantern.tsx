import { useId } from "react";

export function AuthLantern() {
  const uid = useId().replace(/:/g, "");
  const glow = `${uid}-glow`;
  const flame = `${uid}-flame`;

  return (
    <svg viewBox="0 0 280 420" role="img" aria-hidden="true" focusable="false" className="auth-lantern">
      <defs>
        <radialGradient id={glow} cx="50%" cy="42%" r="55%">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.55" />
          <stop offset="70%" stopColor="var(--accent)" stopOpacity="0.08" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
        </radialGradient>
        <linearGradient id={flame} x1="50%" y1="100%" x2="50%" y2="0%">
          <stop offset="0%" stopColor="var(--accent)" />
          <stop offset="100%" stopColor="var(--ink)" />
        </linearGradient>
      </defs>
      <ellipse cx="140" cy="388" rx="78" ry="14" fill={`url(#${glow})`} />
      <circle cx="140" cy="168" r="118" fill={`url(#${glow})`} />
      <line x1="140" y1="8" x2="140" y2="58" stroke="var(--accent)" strokeWidth="2.2" />
      <path d="M140 8 l6 10 h-12 z" fill="var(--accent)" />
      <path
        d="M118 72 l22-16 22 16 v18 H118z"
        fill="var(--accent)"
        stroke="color-mix(in srgb, var(--accent-ink) 35%, var(--accent))"
        strokeWidth="1.2"
        strokeLinejoin="round"
      />
      <path
        d="M112 90 h56 l10 118 H102z"
        fill="color-mix(in srgb, var(--elev) 82%, var(--accent))"
        stroke="var(--accent)"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      <path d="M122 104 h36 l6 90 H116z" fill="color-mix(in srgb, var(--bg) 55%, transparent)" />
      <ellipse className="lantern-glow" cx="140" cy="152" rx="16" ry="22" fill="var(--accent)" opacity="0.55" />
      <path d="M140 176 c-8-14 -6-28 0-40 7 10 9 24 0 40z" fill={`url(#${flame})`} />
      <rect x="98" y="206" width="84" height="12" rx="3" fill="var(--accent)" />
      <path d="M110 218 h60 l-8 22 H118z" fill="color-mix(in srgb, var(--elev) 70%, var(--accent))" stroke="var(--line)" />
      <line x1="140" y1="240" x2="140" y2="268" stroke="var(--line)" strokeWidth="2" />
      <circle cx="140" cy="274" r="5" fill="var(--accent)" />
    </svg>
  );
}
