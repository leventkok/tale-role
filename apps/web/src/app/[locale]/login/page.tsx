import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthForm } from "@/components/auth-form";
import { AuthLantern } from "@/components/art/auth-lantern";
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
    <main className="auth-wrap">
      <div className="auth-art">
        <AuthLantern />
      </div>
      <div className="panel auth-card">
        <h1>{t("nav.signIn")}</h1>
        <AuthForm mode="login" />
        <p className="muted">
          {t("auth.needAccount")} <Link href="/register">{t("nav.register")}</Link>
        </p>
      </div>
    </main>
  );
}
