import { getTranslations, setRequestLocale } from "next-intl/server";
import { apiBase, getSessionToken } from "@/lib/session";
import { Link } from "@/i18n/routing";
import { locales } from "@tale-role/i18n";

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations();
  const token = await getSessionToken();
  let email: string | null = null;
  if (token) {
    const me = await fetch(`${apiBase()}/api/v1/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (me.ok) {
      const body = (await me.json()) as { email?: string };
      email = body.email ?? null;
    }
  }

  return (
    <main>
      <h1>{t("app.name")}</h1>
      <p>{t("app.tagline")}</p>
      {email ? (
        <>
          <p>{t("home.signedInAs", { email })}</p>
          <p>{t("home.cta")}</p>
          <form action="/api/auth/logout" method="post">
            <button type="submit">{t("nav.signOut")}</button>
          </form>
        </>
      ) : (
        <p>
          <Link href="/login">{t("nav.signIn")}</Link>
          {" · "}
          <Link href="/register">{t("nav.register")}</Link>
        </p>
      )}
    </main>
  );
}
