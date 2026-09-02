import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthForm } from "@/components/auth-form";

export default async function VerifyPage({
  params,
  searchParams,
}: {
  params: Promise<{ locale: string }>;
  searchParams: Promise<{ email?: string }>;
}) {
  const { locale } = await params;
  const { email } = await searchParams;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main>
      <h1>{t("auth.verify")}</h1>
      <AuthForm mode="verify" email={email} />
    </main>
  );
}
