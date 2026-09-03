import type { ReactNode } from "react";

type MarkProps = { className?: string };

function frame(children: ReactNode, className?: string) {
  return (
    <svg
      className={className}
      viewBox="0 0 48 48"
      width="40"
      height="40"
      aria-hidden="true"
      focusable="false"
    >
      <rect
        x="1.5"
        y="1.5"
        width="45"
        height="45"
        rx="12"
        fill="color-mix(in srgb, var(--bg) 55%, var(--elev))"
        stroke="var(--line)"
      />
      {children}
    </svg>
  );
}

export function UniverseMark({ className }: MarkProps) {
  return frame(
    <>
      <path
        d="M14 32 V16.5 h14.5 a6 6 0 0 1 0 12 H20"
        fill="none"
        stroke="var(--accent)"
        strokeWidth="2.2"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
      <path
        d="M20 28.5 h10.5 a4.5 4.5 0 0 1 0 9 H14"
        fill="none"
        stroke="var(--ink)"
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
        opacity="0.7"
      />
    </>,
    className,
  );
}

export function HostMark({ className }: MarkProps) {
  return frame(
    <>
      <ellipse cx="24" cy="30" rx="12" ry="4.2" fill="none" stroke="var(--line)" strokeWidth="1.6" />
      <path
        d="M12 30 V18.5 A12 6 0 0 1 36 18.5 V30"
        fill="color-mix(in srgb, var(--elev) 80%, var(--accent))"
        stroke="var(--accent)"
        strokeWidth="1.8"
      />
      <ellipse cx="24" cy="18.5" rx="12" ry="5" fill="var(--elev)" stroke="var(--accent)" strokeWidth="1.8" />
    </>,
    className,
  );
}

export function PlayMark({ className }: MarkProps) {
  return frame(
    <>
      <polygon
        points="24,10 34,16 34,28 24,34 14,28 14,16"
        fill="color-mix(in srgb, var(--elev) 50%, var(--accent))"
        stroke="var(--accent)"
        strokeWidth="1.8"
        strokeLinejoin="round"
      />
      <polygon points="24,10 34,16 24,20 14,16" fill="var(--accent)" opacity="0.4" />
      <circle cx="24" cy="23" r="2.2" fill="var(--ink)" />
    </>,
    className,
  );
}
