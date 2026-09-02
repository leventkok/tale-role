import { getTranslations, setRequestLocale } from "next-intl/server";
import { HostForm } from "@/components/host-form";

export default async function HostPage({
  params,
  searchParams,
}: {
  params: Promise<{ locale: string }>;
  searchParams: Promise<{ universe?: string }>;
}) {
  const { locale } = await params;
  const { universe } = await searchParams;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main className="page">
      <section className="hero">
        <h1>{t("nav.host")}</h1>
        <p className="lede">{t("home.hostLead")}</p>
      </section>
      <div className="panel narrow">
        <HostForm universeId={universe} />
      </div>
    </main>
  );
}
