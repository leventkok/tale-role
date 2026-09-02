import { getTranslations, setRequestLocale } from "next-intl/server";
import { JoinForm } from "@/components/join-form";

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
      <section className="hero">
        <h1>{t("nav.play")}</h1>
        <p className="lede">{t("home.playLead")}</p>
      </section>
      <div className="panel narrow">
        <JoinForm />
      </div>
    </main>
  );
}
