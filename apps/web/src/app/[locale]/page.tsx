import { getTranslations, setRequestLocale } from "next-intl/server";
import { getSessionToken, gqlUpstream } from "@/lib/session";
import { Link } from "@/i18n/routing";
import { locales } from "@tale-role/i18n";
import { HeroTable } from "@/components/art/hero-table";
import { HostMark, PlayMark, UniverseMark } from "@/components/art/beat-marks";
import { DesktopDownload } from "@/components/desktop-download";

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
        <div className="hero-copy">
          <p className="pill">{t("app.tagline")}</p>
          <h1>{t("app.name")}</h1>
          <p className="lede">{email ? t("home.signedInAs", { email }) : t("home.cta")}</p>
          <p className="muted">{t("home.pitch")}</p>
          {email ? null : (
            <div className="btn-row">
              <Link className="btn" href="/login">
                {t("nav.signIn")}
              </Link>
              <Link className="btn ghost" href="/register">
                {t("nav.register")}
              </Link>
            </div>
          )}
          <DesktopDownload />
        </div>
        <div className="hero-art">
          <HeroTable />
        </div>
      </section>
      <div className="hero-beats">
        <article className="card beat">
          <UniverseMark />
          <h2>{t("nav.universe")}</h2>
          <p className="muted">{t("home.universeLead")}</p>
          {email ? (
            <Link className="btn" href="/universe/new">
              {t("nav.universe")}
            </Link>
          ) : null}
        </article>
        <article className="card beat">
          <HostMark />
          <h2>{t("nav.host")}</h2>
          <p className="muted">{t("home.hostLead")}</p>
          {email ? (
            <Link className="btn" href="/host">
              {t("nav.host")}
            </Link>
          ) : null}
        </article>
        <article className="card beat">
          <PlayMark />
          <h2>{t("nav.play")}</h2>
          <p className="muted">{t("home.playLead")}</p>
          {email ? (
            <Link className="btn" href="/play">
              {t("nav.play")}
            </Link>
          ) : null}
        </article>
      </div>
    </main>
  );
}
