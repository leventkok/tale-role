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

export function UniverseWizard() {
  const t = useTranslations("universe");
  const themes = useTranslations("themes");
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [nameEn, setNameEn] = useState("Ashwood");
  const [nameTr, setNameTr] = useState("");
  const [era, setEra] = useState("");
  const [tone, setTone] = useState("");
  const [taboos, setTaboos] = useState("");
  const [themeId, setThemeId] = useState<(typeof THEMES)[number]>("high-fantasy");
  const [dice, setDice] = useState("d20");
  const [rating, setRating] = useState("teen");
  const [npcName, setNpcName] = useState("The Warden");
  const [npcAlign, setNpcAlign] = useState("neutral");
  const [npcVoice, setNpcVoice] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function finish() {
    setBusy(true);
    setError(null);
    const npcs = npcName.trim()
      ? [{ nameEn: npcName.trim(), alignment: npcAlign, voice: npcVoice.trim() }]
      : [];
    const result = await gql<{ createUniverse: { id: string } }>(
      `mutation (
        $nameEn: String!, $nameTr: String, $themeId: String!, $diceSystem: String,
        $contentRating: String, $era: String, $tone: String, $taboos: String, $npcs: [NPCInput]
      ) {
        createUniverse(
          nameEn: $nameEn, nameTr: $nameTr, themeId: $themeId, diceSystem: $diceSystem,
          contentRating: $contentRating, era: $era, tone: $tone, taboos: $taboos, npcs: $npcs
        ) { id }
      }`,
      {
        nameEn,
        nameTr: nameTr || null,
        themeId,
        diceSystem: dice,
        contentRating: rating,
        era,
        tone,
        taboos,
        npcs,
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
    <div>
      <p className="muted">
        {t("step", { n: String(step + 1) })} · {t("stubHint")}
      </p>
      {step === 0 ? (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setStep(1);
          }}
        >
          <label>
            {t("nameEn")}
            <input value={nameEn} onChange={(e) => setNameEn(e.target.value)} required />
          </label>
          <label>
            {t("nameTr")}
            <input value={nameTr} onChange={(e) => setNameTr(e.target.value)} />
          </label>
          <label>
            {t("era")}
            <input value={era} onChange={(e) => setEra(e.target.value)} />
          </label>
          <label>
            {t("tone")}
            <input value={tone} onChange={(e) => setTone(e.target.value)} />
          </label>
          <label>
            {t("taboos")}
            <textarea value={taboos} onChange={(e) => setTaboos(e.target.value)} rows={3} />
          </label>
          <button type="submit">{t("next")}</button>
        </form>
      ) : null}
      {step === 1 ? (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setStep(2);
          }}
        >
          <label>
            {t("theme")}
            <select value={themeId} onChange={(e) => setThemeId(e.target.value as (typeof THEMES)[number])}>
              {THEMES.map((id) => (
                <option key={id} value={id}>
                  {themes(id)}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t("dice")}
            <select value={dice} onChange={(e) => setDice(e.target.value)}>
              <option value="d20">d20</option>
              <option value="2d6">2d6</option>
            </select>
          </label>
          <label>
            {t("rating")}
            <select value={rating} onChange={(e) => setRating(e.target.value)}>
              <option value="everyone">{t("ratingEveryone")}</option>
              <option value="teen">{t("ratingTeen")}</option>
              <option value="mature">{t("ratingMature")}</option>
            </select>
          </label>
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
          onSubmit={(e) => {
            e.preventDefault();
            void finish();
          }}
        >
          <p className="muted">{t("npcHint")}</p>
          <label>
            {t("npcName")}
            <input value={npcName} onChange={(e) => setNpcName(e.target.value)} />
          </label>
          <label>
            {t("alignment")}
            <select value={npcAlign} onChange={(e) => setNpcAlign(e.target.value)}>
              <option value="good">{t("alignGood")}</option>
              <option value="neutral">{t("alignNeutral")}</option>
              <option value="evil">{t("alignEvil")}</option>
            </select>
          </label>
          <label>
            {t("voice")}
            <input value={npcVoice} onChange={(e) => setNpcVoice(e.target.value)} />
          </label>
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
