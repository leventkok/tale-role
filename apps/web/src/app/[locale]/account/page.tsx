import { getTranslations, setRequestLocale } from "next-intl/server";
import { getSessionToken } from "@/lib/session";
import { AccountPanel } from "@/components/account-panel";
import { Link } from "@/i18n/routing";

export default async function AccountPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("account");
  const nav = await getTranslations("nav");
  const signedIn = Boolean(await getSessionToken());
  return (
    <main className="page">
      <section className="hero">
        <h1>{t("title")}</h1>
        <p className="lede">{t("lead")}</p>
      </section>
      <div className="panel account-panel">
        {signedIn ? (
          <AccountPanel />
        ) : (
          <p>
            {t("needSignIn")} <Link href="/login">{nav("signIn")}</Link>
          </p>
        )}
      </div>
    </main>
  );
}
