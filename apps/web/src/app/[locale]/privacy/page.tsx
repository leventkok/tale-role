import { getTranslations, setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/routing";

export default async function PrivacyPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("privacy");
  const nav = await getTranslations("nav");
  return (
    <main className="page">
      <section className="hero">
        <h1>{t("title")}</h1>
        <p className="lede">{t("lead")}</p>
      </section>
      <article className="panel narrow">
        <h2>{t("store")}</h2>
        <p>{t("storeBody")}</p>
        <h2>{t("llm")}</h2>
        <p>{t("llmBody")}</p>
        <h2>{t("rights")}</h2>
        <p>{t("rightsBody")}</p>
        <p>
          <Link href="/account">{nav("account")}</Link>
        </p>
        <h2>{t("not")}</h2>
        <p className="muted">{t("notBody")}</p>
      </article>
    </main>
  );
}
