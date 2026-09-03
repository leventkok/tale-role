import { getTranslations, setRequestLocale } from "next-intl/server";
import { UniverseWizard } from "@/components/universe-wizard";

export default async function UniverseNewPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("universe");
  return (
    <main className="page">
      <section className="hero">
        <h1>{t("title")}</h1>
        <p className="lede">{t("lead")}</p>
      </section>
      <div className="panel world-panel">
        <UniverseWizard />
      </div>
    </main>
  );
}
