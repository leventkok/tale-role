import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthForm } from "@/components/auth-form";
import { Link } from "@/i18n/routing";

export default async function LoginPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations();
  return (
    <main>
      <h1>{t("nav.signIn")}</h1>
      <AuthForm mode="login" />
      <p>
        <Link href="/register">{t("nav.register")}</Link>
      </p>
    </main>
  );
}
