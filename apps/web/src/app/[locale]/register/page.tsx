import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthForm } from "@/components/auth-form";
import { Link } from "@/i18n/routing";

export default async function RegisterPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main>
      <h1>{t("nav.register")}</h1>
      <AuthForm mode="register" />
      <p>
        <Link href="/login">{t("nav.signIn")}</Link>
      </p>
    </main>
  );
}
