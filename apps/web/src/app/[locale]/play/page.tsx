import { getTranslations, setRequestLocale } from "next-intl/server";
import { LobbyBoard } from "@/components/lobby-board";

export default async function PlayPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main className="page">
      <section className="hero hero-plain">
        <div className="hero-copy">
          <h1>{t("lobby.title")}</h1>
          <p className="lede">{t("lobby.lead")}</p>
        </div>
      </section>
      <LobbyBoard />
    </main>
  );
}
