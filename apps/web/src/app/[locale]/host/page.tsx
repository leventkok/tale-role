import { getTranslations, setRequestLocale } from "next-intl/server";
import { HostForm } from "@/components/host-form";

export default async function HostPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main>
      <h1>{t("nav.host")}</h1>
      <HostForm />
    </main>
  );
}
