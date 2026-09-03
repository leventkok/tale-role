"use client";

import { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { gql, gqlData } from "@/lib/gql";

const THEMES = [
  "high-fantasy",
  "gothic-horror",
  "space-opera",
  "cyber-noir",
  "post-apocalyptic",
  "fairytale",
] as const;

type NpcDraft = { name: string; alignment: string; detail: string };

const emptyNpc = (): NpcDraft => ({ name: "", alignment: "neutral", detail: "" });

export function UniverseWizard() {
  const t = useTranslations("universe");
  const themes = useTranslations("themes");
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [name, setName] = useState("");
  const [era, setEra] = useState("");
  const [tone, setTone] = useState("");
  const [themeId, setThemeId] = useState<(typeof THEMES)[number]>("high-fantasy");
  const [story, setStory] = useState("");
  const [opening, setOpening] = useState("");
  const [npcs, setNpcs] = useState<NpcDraft[]>([emptyNpc()]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  function patchNpc(i: number, patch: Partial<NpcDraft>) {
    setNpcs((rows) => rows.map((row, idx) => (idx === i ? { ...row, ...patch } : row)));
  }

  async function finish() {
    setBusy(true);
    setError(null);
    const packed = npcs
      .filter((n) => n.name.trim())
      .map((n) => ({ nameEn: n.name.trim(), alignment: n.alignment, voice: n.detail.trim() }));
    const result = await gql<{ createUniverse: { id: string } }>(
      `mutation (
        $nameEn: String!, $themeId: String!, $era: String, $tone: String,
        $description: String, $opening: String, $npcs: [NPCInput]
      ) {
        createUniverse(
          nameEn: $nameEn, themeId: $themeId, era: $era, tone: $tone,
          description: $description, opening: $opening, npcs: $npcs
        ) { id }
      }`,
      {
        nameEn: name.trim(),
        themeId,
        era,
        tone,
        description: story,
        opening,
        npcs: packed,
      },
    );
    const created = gqlData(result)?.createUniverse;
    setBusy(false);
    if (!created?.id) {
      setError(result.errors?.[0]?.message ?? "error");
      return;
    }
    router.push(`/host?universe=${created.id}`);
  }

  return (
    <div className="world-wizard">
      <ol className="world-steps" aria-label={t("title")}>
        <li className={step === 0 ? "on" : undefined}>{t("stepWorld")}</li>
        <li className={step === 1 ? "on" : undefined}>{t("stepPeople")}</li>
        <li className={step === 2 ? "on" : undefined}>{t("stepReady")}</li>
      </ol>

      {step === 0 ? (
        <form
          className="world-sheet"
          onSubmit={(e) => {
            e.preventDefault();
            setStep(1);
          }}
        >
          <label>
            {t("name")}
            <span className="hint">{t("nameHint")}</span>
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <div className="grid-2">
            <label>
              {t("era")}
              <span className="hint">{t("eraHint")}</span>
              <input value={era} onChange={(e) => setEra(e.target.value)} required />
            </label>
            <label>
              {t("tone")}
              <span className="hint">{t("toneHint")}</span>
              <input value={tone} onChange={(e) => setTone(e.target.value)} required />
            </label>
          </div>
          <label>
            {t("look")}
            <span className="hint">{t("lookHint")}</span>
            <select value={themeId} onChange={(e) => setThemeId(e.target.value as (typeof THEMES)[number])}>
              {THEMES.map((id) => (
                <option key={id} value={id}>
                  {themes(id)}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t("story")}
            <span className="hint">{t("storyHint")}</span>
            <textarea value={story} onChange={(e) => setStory(e.target.value)} rows={6} required />
          </label>
          <label>
            {t("opening")}
            <span className="hint">{t("openingHint")}</span>
            <textarea value={opening} onChange={(e) => setOpening(e.target.value)} rows={3} />
          </label>
          <button type="submit">{t("next")}</button>
        </form>
      ) : null}

      {step === 1 ? (
        <form
          className="world-sheet"
          onSubmit={(e) => {
            e.preventDefault();
            setStep(2);
          }}
        >
          <p className="lede">{t("npcLead")}</p>
          {npcs.map((npc, i) => (
            <fieldset className="npc-card" key={i}>
              <legend>
                {t("npcName")} {i + 1}
              </legend>
              <label>
                {t("npcName")}
                <input value={npc.name} onChange={(e) => patchNpc(i, { name: e.target.value })} />
              </label>
              <label>
                {t("alignment")}
                <select value={npc.alignment} onChange={(e) => patchNpc(i, { alignment: e.target.value })}>
                  <option value="good">{t("alignGood")}</option>
                  <option value="neutral">{t("alignNeutral")}</option>
                  <option value="evil">{t("alignEvil")}</option>
                </select>
              </label>
              <label>
                {t("npcDetail")}
                <span className="hint">{t("npcDetailHint")}</span>
                <textarea value={npc.detail} onChange={(e) => patchNpc(i, { detail: e.target.value })} rows={4} />
              </label>
              {npcs.length > 1 ? (
                <button className="ghost" type="button" onClick={() => setNpcs((rows) => rows.filter((_, idx) => idx !== i))}>
                  {t("removeNpc")}
                </button>
              ) : null}
            </fieldset>
          ))}
          <button className="ghost" type="button" onClick={() => setNpcs((rows) => [...rows, emptyNpc()])}>
            {t("addNpc")}
          </button>
          <div className="btn-row">
            <button className="ghost" type="button" onClick={() => setStep(0)}>
              {t("back")}
            </button>
            <button type="submit">{t("next")}</button>
          </div>
        </form>
      ) : null}

      {step === 2 ? (
        <form
          className="world-sheet"
          onSubmit={(e) => {
            e.preventDefault();
            void finish();
          }}
        >
          <p className="lede">{t("review")}</p>
          <article className="card world-review">
            <p className="pill">{themes(themeId)}</p>
            <h2>{name}</h2>
            <p className="muted">
              {era} · {tone}
            </p>
            <p>{story}</p>
            {opening ? <p className="muted">{opening}</p> : null}
            <ul className="npc-review">
              {npcs
                .filter((n) => n.name.trim())
                .map((n) => (
                  <li key={n.name}>
                    <strong>{n.name}</strong>
                    {" · "}
                    {n.alignment === "good" ? t("alignGood") : n.alignment === "evil" ? t("alignEvil") : t("alignNeutral")}
                    {n.detail ? <span className="muted"> — {n.detail}</span> : null}
                  </li>
                ))}
            </ul>
          </article>
          {error ? (
            <p className="alert" role="alert">
              {error}
            </p>
          ) : null}
          <div className="btn-row">
            <button className="ghost" type="button" onClick={() => setStep(1)}>
              {t("back")}
            </button>
            <button type="submit" disabled={busy}>
              {t("compile")}
            </button>
          </div>
        </form>
      ) : null}
    </div>
  );
}
