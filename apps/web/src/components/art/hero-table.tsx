import { useId } from "react";

function Chair({ x, y, scale = 1 }: { x: number; y: number; scale?: number }) {
  return (
    <g transform={`translate(${x} ${y}) scale(${scale})`}>
      <rect
        x="-24"
        y="-62"
        width="48"
        height="54"
        rx="7"
        fill="color-mix(in srgb, var(--elev) 88%, var(--bg))"
        stroke="var(--line)"
        strokeWidth="2"
      />
      <path
        d="M-14 -54 h28 v8 h-28z"
        fill="color-mix(in srgb, var(--accent) 22%, var(--elev))"
      />
      <rect
        x="-28"
        y="-12"
        width="56"
        height="16"
        rx="4"
        fill="color-mix(in srgb, var(--elev) 70%, var(--bg))"
        stroke="var(--line)"
        strokeWidth="2"
      />
      <line x1="-20" y1="4" x2="-24" y2="42" stroke="var(--line)" strokeWidth="3.2" strokeLinecap="round" />
      <line x1="20" y1="4" x2="24" y2="42" stroke="var(--line)" strokeWidth="3.2" strokeLinecap="round" />
    </g>
  );
}

export function HeroTable() {
  const uid = useId().replace(/:/g, "");
  const haze = `${uid}-haze`;
  const wood = `${uid}-wood`;
  const flame = `${uid}-flame`;

  return (
    <svg viewBox="0 0 720 520" role="img" aria-hidden="true" focusable="false" className="hero-table">
      <defs>
        <radialGradient id={haze} cx="50%" cy="28%" r="62%">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.28" />
          <stop offset="55%" stopColor="var(--accent)" stopOpacity="0.06" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
        </radialGradient>
        <linearGradient id={wood} x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="var(--elev)" />
          <stop offset="100%" stopColor="color-mix(in srgb, var(--bg) 55%, var(--elev))" />
        </linearGradient>
        <linearGradient id={flame} x1="50%" y1="100%" x2="50%" y2="0%">
          <stop offset="0%" stopColor="var(--accent)" />
          <stop offset="100%" stopColor="var(--ink)" />
        </linearGradient>
      </defs>

      <ellipse cx="360" cy="430" rx="250" ry="28" fill="var(--bg)" opacity="0.55" />
      <circle cx="360" cy="150" r="210" fill={`url(#${haze})`} />

      <Chair x={168} y={248} scale={0.92} />
      <Chair x={360} y={228} />
      <Chair x={552} y={248} scale={0.92} />

      <path
        d="M96 338 A264 74 0 0 0 624 338 L610 372 A250 70 0 0 1 110 372 Z"
        fill={`url(#${wood})`}
        stroke="var(--line)"
        strokeWidth="2"
      />
      <line x1="150" y1="368" x2="136" y2="438" stroke="var(--line)" strokeWidth="5" strokeLinecap="round" />
      <line x1="570" y1="368" x2="584" y2="438" stroke="var(--line)" strokeWidth="5" strokeLinecap="round" />
      <line x1="250" y1="372" x2="244" y2="448" stroke="var(--line)" strokeWidth="4" strokeLinecap="round" />
      <line x1="470" y1="372" x2="476" y2="448" stroke="var(--line)" strokeWidth="4" strokeLinecap="round" />

      <ellipse
        cx="360"
        cy="332"
        rx="264"
        ry="74"
        fill="var(--elev)"
        stroke="var(--line)"
        strokeWidth="2.4"
      />
      <ellipse
        cx="360"
        cy="332"
        rx="228"
        ry="54"
        fill="none"
        stroke="color-mix(in srgb, var(--accent) 32%, var(--line))"
        strokeWidth="1.4"
      />

      <g transform="translate(360 18)">
        <line x1="0" y1="0" x2="0" y2="52" stroke="var(--accent)" strokeWidth="2" />
        <path d="M-22 68 L0 50 L22 68 Z" fill="var(--accent)" />
        <path
          d="M-20 68 L-24 132 L24 132 L20 68 Z"
          fill="color-mix(in srgb, var(--elev) 75%, var(--accent))"
          stroke="var(--accent)"
          strokeWidth="2"
          strokeLinejoin="round"
        />
        <path d="M-12 80 H12 L16 118 H-16 Z" fill="color-mix(in srgb, var(--bg) 50%, transparent)" />
        <ellipse className="lantern-glow" cx="0" cy="104" rx="12" ry="16" fill="var(--accent)" />
        <path d="M0 122 c-6-10 -5-20 0-30 5 8 7 18 0 30z" fill={`url(#${flame})`} />
        <rect x="-26" y="132" width="52" height="9" rx="2" fill="var(--accent)" />
      </g>

      <g transform="translate(214 300)">
        <path
          d="M-38 18 L-32 -8 Q0 -22 38 -6 L32 22 Q0 10 -38 18 Z"
          fill="color-mix(in srgb, var(--ink) 14%, var(--elev))"
          stroke="var(--line)"
          strokeWidth="1.6"
        />
        <path d="M-8 -10 Q0 -2 6 14" fill="none" stroke="var(--muted)" strokeWidth="1.2" />
        <path d="M-22 2 Q-10 8 4 6" fill="none" stroke="var(--muted)" strokeWidth="1.1" opacity="0.7" />
      </g>

      <g transform="translate(292 286)">
        <rect x="-7" y="8" width="14" height="22" rx="2" fill="color-mix(in srgb, var(--ink) 12%, var(--elev))" stroke="var(--line)" />
        <path d="M0 10 C-8 -8 -3 -22 0 -30 C4 -18 9 -6 0 10Z" fill={`url(#${flame})`} />
        <ellipse className="lantern-glow" cx="0" cy="-6" rx="16" ry="20" fill="var(--accent)" opacity="0.25" />
      </g>

      <g transform="translate(388 292)">
        <polygon
          points="0,-30 26,-15 26,15 0,30 -26,15 -26,-15"
          fill="color-mix(in srgb, var(--elev) 45%, var(--accent))"
          stroke="var(--accent)"
          strokeWidth="2"
          strokeLinejoin="round"
        />
        <polygon points="0,-30 26,-15 0,4 -26,-15" fill="var(--accent)" opacity="0.38" />
        <path d="M0 4 L26 -15 L26 15 Z" fill="var(--accent)" opacity="0.12" />
        <path d="M0 4 L-26 -15 L-26 15 Z" fill="var(--bg)" opacity="0.2" />
      </g>

      <g transform="translate(478 308)">
        <path
          d="M-40 6 C-20 -16 12 -14 42 4 L38 22 C10 8 -22 6 -40 6 Z"
          fill="color-mix(in srgb, var(--accent) 18%, var(--elev))"
          stroke="var(--line)"
          strokeWidth="1.5"
        />
        <path d="M-18 4 L8 -6" stroke="var(--muted)" strokeWidth="1" />
        <path d="M-12 12 L16 2" stroke="var(--muted)" strokeWidth="1" opacity="0.7" />
      </g>

      <g transform="translate(248 338)">
        <ellipse cx="0" cy="0" rx="16" ry="7" fill="var(--elev)" stroke="var(--line)" />
        <path d="M-10 -2 C-8 -18 8 -18 10 -2" fill="none" stroke="var(--accent)" strokeWidth="2" />
      </g>

      <g transform="translate(430 340)">
        <ellipse cx="0" cy="4" rx="14" ry="6" fill="color-mix(in srgb, var(--elev) 80%, var(--bg))" stroke="var(--line)" />
        <path d="M-8 2 L-6 -10 H6 L8 2" fill="var(--elev)" stroke="var(--line)" />
      </g>
    </svg>
  );
}
