export function BrandMark({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 32 32"
      width="28"
      height="28"
      aria-hidden="true"
      focusable="false"
    >
      <line x1="16" y1="2" x2="16" y2="7" stroke="var(--accent)" strokeWidth="1.6" />
      <path d="M16 2 l3.2 4.2 h-6.4 z" fill="var(--accent)" />
      <path d="M11.2 8.4 l4.8-3.4 4.8 3.4 v2.6 H11.2z" fill="var(--accent)" />
      <path
        d="M10.2 11 h11.6 l2.2 12.4 H8z"
        fill="color-mix(in srgb, var(--elev) 70%, var(--accent))"
        stroke="var(--accent)"
        strokeWidth="1.3"
        strokeLinejoin="round"
      />
      <path d="M12.4 13.2 h7.2 l1.3 8.2 H11.1z" fill="color-mix(in srgb, var(--bg) 45%, transparent)" />
      <ellipse className="lantern-glow" cx="16" cy="18" rx="3.2" ry="4.2" fill="var(--accent)" />
      <path d="M16 22.2c-1.6-2.2-1.2-4.4 0-6.4 1.4 1.6 1.8 3.8 0 6.4z" fill="var(--ink)" />
      <rect x="9.2" y="23.2" width="13.6" height="2.1" rx="0.6" fill="var(--accent)" />
    </svg>
  );
}
