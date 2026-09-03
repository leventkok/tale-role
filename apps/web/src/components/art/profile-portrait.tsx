import type { ReactNode } from "react";
import {
  defaultPortraitId,
  normalizePortraitId,
  type PortraitId,
} from "@/lib/portraits";

function Frame({ id, children }: { id: string; children: ReactNode }) {
  const clip = `${id}-clip`;
  const shine = `${id}-shine`;
  const rim = `${id}-rim`;
  return (
    <>
      <defs>
        <clipPath id={clip}>
          <rect x="10" y="10" width="140" height="140" rx="10" />
        </clipPath>
        <radialGradient id={shine} cx="38%" cy="28%" r="55%">
          <stop offset="0" stopColor="#fff6d6" stopOpacity="0.35" />
          <stop offset="1" stopColor="#000" stopOpacity="0" />
        </radialGradient>
        <linearGradient id={rim} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#f3e0a8" />
          <stop offset="0.45" stopColor="#c9a227" />
          <stop offset="1" stopColor="#6e4e14" />
        </linearGradient>
      </defs>
      <rect x="4" y="4" width="152" height="152" rx="14" fill="#120e0a" />
      <g clipPath={`url(#${clip})`}>{children}</g>
      <rect x="10" y="10" width="140" height="140" fill={`url(#${shine})`} />
      <rect x="8" y="8" width="144" height="144" rx="12" fill="none" stroke={`url(#${rim})`} strokeWidth="5" />
      <rect x="14" y="14" width="132" height="132" rx="8" fill="none" stroke="#7a5a1e" strokeWidth="1.2" opacity="0.85" />
      <path d="M18 28 L28 18" stroke="#f3e0a8" strokeWidth="2" />
      <path d="M142 28 L132 18" stroke="#f3e0a8" strokeWidth="2" />
      <path d="M18 132 L28 142" stroke="#c9a227" strokeWidth="2" />
      <path d="M142 132 L132 142" stroke="#c9a227" strokeWidth="2" />
      <circle cx="18" cy="18" r="3.2" fill="#e2c07a" />
      <circle cx="142" cy="18" r="3.2" fill="#e2c07a" />
      <circle cx="18" cy="142" r="3.2" fill="#c9a227" />
      <circle cx="142" cy="142" r="3.2" fill="#c9a227" />
    </>
  );
}

function Warden() {
  return (
    <Frame id="warden">
      <rect width="160" height="160" fill="#1a2418" />
      <ellipse cx="80" cy="48" rx="48" ry="28" fill="#c9a227" opacity="0.18" />
      <path d="M18 160 Q80 92 142 160" fill="#2a2214" />
      <path d="M28 118 L48 78 L72 92 L80 70 L88 92 L112 78 L132 118 L118 160 H42z" fill="#4a3a22" />
      <path d="M36 124 L52 90 L70 102 L80 82 L90 102 L108 90 L124 124" fill="#c9a227" opacity="0.55" />
      <ellipse cx="80" cy="72" rx="22" ry="26" fill="#c4a07a" />
      <path d="M58 62 Q80 38 102 62 L96 78 Q80 68 64 78z" fill="#3a2a14" />
      <path d="M62 54 L80 42 L98 54 L92 64 L80 58 L68 64z" fill="#c9a227" />
      <rect x="66" y="58" width="28" height="7" rx="2" fill="#1a140c" />
      <rect x="70" y="59.5" width="20" height="2.2" fill="#e2c07a" />
      <circle cx="72" cy="74" r="2" fill="#1a140c" />
      <circle cx="88" cy="74" r="2" fill="#1a140c" />
      <path d="M74 84 q6 5 12 0" fill="none" stroke="#6a4030" strokeWidth="1.4" />
      <rect x="72" y="108" width="16" height="22" rx="3" fill="#e2c07a" />
      <rect x="75" y="111" width="10" height="16" rx="2" fill="#fff3c0" opacity="0.7" />
      <path d="M80 108 v-8" stroke="#c9a227" strokeWidth="2" />
    </Frame>
  );
}

function Blade() {
  return (
    <Frame id="blade">
      <rect width="160" height="160" fill="#12161a" />
      <path d="M0 90 Q80 40 160 90 V160 H0z" fill="#1c2420" />
      <path d="M24 160 Q80 96 136 160" fill="#24302a" />
      <path d="M40 120 L56 78 L80 88 L104 78 L120 120 L110 160 H50z" fill="#2a322c" />
      <path d="M48 118 L60 86 L80 94 L100 86 L112 118" fill="#6ed3a8" opacity="0.22" />
      <ellipse cx="80" cy="74" rx="20" ry="24" fill="#7a5a44" />
      <path d="M52 70 Q80 28 108 70 L100 92 Q80 78 60 92z" fill="#1a1e18" />
      <path d="M58 78 Q80 62 102 78" fill="none" stroke="#6ed3a8" strokeWidth="1.2" opacity="0.7" />
      <circle cx="72" cy="76" r="2.1" fill="#9affee" />
      <circle cx="88" cy="76" r="2.1" fill="#1a140c" />
      <path d="M70 86 h8" stroke="#4a3024" strokeWidth="1.3" />
      <path d="M28 108 L52 92 L56 102 L34 118z" fill="#c9d6c8" />
      <path d="M132 108 L108 92 L104 102 L126 118z" fill="#c9d6c8" />
      <path d="M50 94 L30 70" stroke="#6ed3a8" strokeWidth="1.4" />
    </Frame>
  );
}

function Sage() {
  return (
    <Frame id="sage">
      <rect width="160" height="160" fill="#14102a" />
      <ellipse cx="80" cy="44" rx="50" ry="30" fill="#7ad0ff" opacity="0.2" />
      <circle cx="118" cy="52" r="10" fill="#7ad0ff" opacity="0.55" />
      <circle cx="118" cy="52" r="5" fill="#e8f4ff" />
      <path d="M22 160 Q80 88 138 160" fill="#2a2048" />
      <path d="M34 128 L58 78 L80 90 L102 78 L126 128 L112 160 H48z" fill="#3a2a60" />
      <path d="M46 122 L62 88 L80 98 L98 88 L114 122" fill="#7ad0ff" opacity="0.28" />
      <ellipse cx="80" cy="76" rx="19" ry="23" fill="#d4b08a" />
      <path d="M54 68 Q80 22 106 68 L98 86 Q80 74 62 86z" fill="#2a1860" />
      <path d="M60 48 Q80 36 100 48" fill="none" stroke="#7ad0ff" strokeWidth="1.3" />
      <circle cx="72" cy="78" r="1.8" fill="#1a1030" />
      <circle cx="88" cy="78" r="1.8" fill="#1a1030" />
      <path d="M74 88 q6 4 12 0" fill="none" stroke="#7a5040" strokeWidth="1.2" />
      <rect x="112" y="58" width="4" height="62" rx="1" fill="#c9a227" />
      <path d="M44 140 h72" stroke="#7ad0ff" strokeWidth="1" opacity="0.4" />
    </Frame>
  );
}

function Ranger() {
  return (
    <Frame id="ranger">
      <rect width="160" height="160" fill="#142016" />
      <path d="M0 70 Q80 20 160 78 V160 H0z" fill="#1c2a1c" />
      <path d="M20 160 Q80 100 140 160" fill="#243424" />
      <path d="M32 126 L54 80 L80 92 L106 80 L128 126 L114 160 H46z" fill="#3a4a30" />
      <ellipse cx="80" cy="74" rx="20" ry="24" fill="#c4a07a" />
      <path d="M54 66 Q80 30 106 66 L98 88 Q80 76 62 88z" fill="#2a3a20" />
      <path d="M70 40 L78 28 L86 42" fill="#8faf78" />
      <path d="M108 70 Q130 88 118 128" fill="none" stroke="#c9a227" strokeWidth="3" />
      <path d="M112 78 Q124 92 116 118" fill="none" stroke="#efe6d4" strokeWidth="1.2" />
      <circle cx="72" cy="76" r="2" fill="#1a140c" />
      <circle cx="88" cy="76" r="2" fill="#1a140c" />
      <path d="M74 86 q6 4 12 0" fill="none" stroke="#6a4030" strokeWidth="1.3" />
      <path d="M40 112 L62 98" stroke="#8faf78" strokeWidth="2" />
    </Frame>
  );
}

function Oath() {
  return (
    <Frame id="oath">
      <rect width="160" height="160" fill="#1a1610" />
      <circle cx="80" cy="48" r="34" fill="#e2c07a" opacity="0.22" />
      <path d="M80 18 L86 40 L80 36 L74 40z" fill="#f3e0a8" />
      <path d="M24 160 Q80 96 136 160" fill="#2a2418" />
      <path d="M30 120 L50 78 L72 90 L80 68 L88 90 L110 78 L130 120 L116 160 H44z" fill="#efe6d4" opacity="0.2" />
      <path d="M38 122 L54 86 L70 98 L80 76 L90 98 L106 86 L122 122" fill="#e2c07a" opacity="0.65" />
      <ellipse cx="80" cy="74" rx="21" ry="25" fill="#e6c8a4" />
      <path d="M62 52 L80 44 L98 52 L92 62 L80 58 L68 62z" fill="#f3e0a8" />
      <circle cx="80" cy="48" r="3.4" fill="#fff6d6" />
      <circle cx="72" cy="76" r="2" fill="#1a140c" />
      <circle cx="88" cy="76" r="2" fill="#1a140c" />
      <path d="M74 86 q6 5 12 0" fill="none" stroke="#8a5a40" strokeWidth="1.3" />
      <path d="M48 108 L80 128 L112 108" fill="none" stroke="#e2c07a" strokeWidth="2.2" />
    </Frame>
  );
}

const bodies: Record<PortraitId, () => ReactNode> = {
  warden: Warden,
  blade: Blade,
  sage: Sage,
  ranger: Ranger,
  oath: Oath,
};

export function ProfilePortrait({
  id,
  className,
}: {
  id?: string | null;
  className?: string;
}) {
  const key = normalizePortraitId(id ?? defaultPortraitId);
  const Body = bodies[key];
  return (
    <svg className={className} viewBox="0 0 160 160" aria-hidden="true" focusable="false">
      <Body />
    </svg>
  );
}
