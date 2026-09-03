import { getTranslations, setRequestLocale } from "next-intl/server";
import { getSessionToken, gqlUpstream } from "@/lib/session";
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
    const me = await gqlUpstream<{ me: { email: string } | null }>(token, "{ me { email } }");
    email = me.data?.me?.email ?? null;
  }

  return (
    <main className="page">
      <section className="hero">
        <p className="pill">{t("app.tagline")}</p>
        <h1>{t("app.name")}</h1>
        <p className="lede">{email ? t("home.signedInAs", { email }) : t("home.cta")}</p>
      </section>
      {email ? (
        <div className="grid-2">
          <article className="card">
            <h2>{t("nav.universe")}</h2>
            <p className="muted">{t("home.universeLead")}</p>
            <Link className="btn" href="/universe/new">
              {t("nav.universe")}
            </Link>
          </article>
          <article className="card">
            <h2>{t("nav.host")}</h2>
            <p className="muted">{t("home.hostLead")}</p>
            <Link className="btn" href="/host">
              {t("nav.host")}
            </Link>
          </article>
          <article className="card">
            <h2>{t("nav.play")}</h2>
            <p className="muted">{t("home.playLead")}</p>
            <Link className="btn" href="/play">
              {t("nav.play")}
            </Link>
          </article>
        </div>
      ) : (
        <div className="btn-row">
          <Link className="btn" href="/login">
            {t("nav.signIn")}
          </Link>
          <Link className="btn ghost" href="/register">
            {t("nav.register")}
          </Link>
        </div>
      )}
    </main>
  );
}
