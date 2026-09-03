"use client";

import { useEffect, useState } from "react";
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
