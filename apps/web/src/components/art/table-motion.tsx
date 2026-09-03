"use client";

import { useEffect, useRef, useState } from "react";
import { ProfilePortrait } from "@/components/art/profile-portrait";
import { portraitFromSeed } from "@/lib/portraits";

export function TypeProse({ text }: { text: string }) {
  const [shown, setShown] = useState(text);
  useEffect(() => {
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduce || text.length < 8) {
      setShown(text);
      return;
    }
    setShown("");
    let i = 0;
    const id = window.setInterval(() => {
      i += 1;
      setShown(text.slice(0, i));
      if (i >= text.length) window.clearInterval(id);
    }, 16);
    return () => window.clearInterval(id);
  }, [text]);
  return <p className="prose">{shown}</p>;
}

export function DiceDrop({ rolls }: { rolls: number[] }) {
  if (!rolls.length) return null;
  return (
    <div className="dice-drop" aria-hidden="true">
      {rolls.map((n, i) => (
        <span className="die" style={{ animationDelay: `${i * 90}ms` }} key={`${n}-${i}`}>
          {n}
        </span>
      ))}
    </div>
  );
}

export type DiceCastShow = {
  rolls: number[];
  total: number;
  system: string;
  success?: boolean | null;
};

export function DiceCast({ cast, onDone }: { cast: DiceCastShow | null; onDone: () => void }) {
  const done = useRef(onDone);
  done.current = onDone;
  useEffect(() => {
    if (!cast) return;
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const id = window.setTimeout(() => done.current(), reduce ? 900 : 2200);
    return () => window.clearTimeout(id);
  }, [cast]);

  if (!cast?.rolls.length) return null;
  const d20 = cast.system !== "2d6";
  const faces = d20 ? [cast.rolls[0] ?? cast.total] : cast.rolls;
  return (
    <div className="dice-cast" role="status" aria-live="polite">
      <div className="dice-cast-stage">
        {faces.map((n, i) => {
          const glow = d20 && n === 20 ? "crit" : d20 && n === 1 ? "fail" : "";
          return (
            <span className={`dice-poly ${glow}`} style={{ animationDelay: `${i * 120}ms` }} key={`${n}-${i}`}>
              <PolyDie value={n} d20={d20} />
            </span>
          );
        })}
      </div>
    </div>
  );
}

function PolyDie({ value, d20 }: { value: number; d20: boolean }) {
  if (!d20) {
    return (
      <svg viewBox="0 0 88 88" className="dice-poly-svg" aria-hidden="true">
        <rect x="8" y="8" width="72" height="72" rx="10" fill="#1a140c" stroke="#e2c07a" strokeWidth="3" />
        <rect x="16" y="16" width="56" height="56" rx="6" fill="#2a2116" stroke="#c9a227" strokeWidth="1.2" />
        <text x="44" y="56" textAnchor="middle" fill="#f3eadc" fontSize="32" fontFamily="Georgia, serif">
          {value}
        </text>
      </svg>
    );
  }
  return (
    <svg viewBox="0 0 120 120" className="dice-poly-svg" aria-hidden="true">
      <polygon points="60,6 112,38 94,102 26,102 8,38" fill="#24180e" stroke="#e2c07a" strokeWidth="3" />
      <polygon points="60,6 112,38 60,46" fill="#e2c07a" opacity="0.28" />
      <polygon points="112,38 94,102 60,46" fill="#000" opacity="0.22" />
      <polygon points="8,38 26,102 60,46" fill="#000" opacity="0.12" />
      <polygon points="60,18 98,42 84,90 36,90 22,42" fill="none" stroke="#f3e0a8" strokeWidth="1.2" opacity="0.7" />
      <text x="60" y="72" textAnchor="middle" fill="#f8f0dc" fontSize="36" fontFamily="Georgia, serif">
        {value}
      </text>
    </svg>
  );
}

export function PortraitMark({
  name,
  you,
  portraitId,
}: {
  name: string;
  you?: boolean;
  portraitId?: string;
}) {
  const id = you && portraitId ? portraitId : portraitFromSeed(name);
  return <ProfilePortrait id={id} className={`portrait-mark ${you ? "you" : ""}`} />;
}
