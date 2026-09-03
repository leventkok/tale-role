import { getTranslations, setRequestLocale } from "next-intl/server";
import { JoinForm } from "@/components/join-form";

export default async function JoinPage({
  params,
}: {
  params: Promise<{ locale: string; roomId: string }>;
}) {
  const { locale, roomId } = await params;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main className="page">
      <section className="hero">
        <h1>{t("nav.play")}</h1>
        <p className="lede">{t("home.playLead")}</p>
      </section>
      <div className="panel narrow">
        <JoinForm initialRoomId={roomId} />
      </div>
    </main>
  );
}
