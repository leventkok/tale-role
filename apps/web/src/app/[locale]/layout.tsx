import { Cinzel, Outfit, Source_Serif_4 } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations, setRequestLocale } from "next-intl/server";
import { notFound } from "next/navigation";
import { isLocale, locales } from "@tale-role/i18n";
import { getSessionToken, gqlUpstream } from "@/lib/session";
import { SiteHeader } from "@/components/site-header";
import { DesktopShellMark } from "@/components/desktop-shell-mark";
import { Link } from "@/i18n/routing";
import "../globals.css";

const desktopShellScript =
  '(function(){try{if(window.taleRoleDesktop||/\\bTaleRoleDesktop\\b/i.test(navigator.userAgent)){document.documentElement.setAttribute("data-talerole-shell","desktop");}}catch(e){}})();';

const display = Cinzel({ subsets: ["latin"], variable: "--font-cinzel" });
const body = Source_Serif_4({ subsets: ["latin"], variable: "--font-serif" });
const sans = Outfit({ subsets: ["latin"], variable: "--font-sans" });

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  if (!isLocale(locale)) {
    notFound();
  }
  setRequestLocale(locale);
  const messages = await getMessages();
  const t = await getTranslations();
  const token = await getSessionToken();
  let email: string | null = null;
  if (token) {
    const me = await gqlUpstream<{ me: { email: string } | null }>(token, "{ me { email } }");
    email = me.data?.me?.email ?? null;
  }

  return (
    <html lang={locale} suppressHydrationWarning data-theme="high-fantasy" className={`${display.variable} ${body.variable} ${sans.variable}`}>
      <head>
        <script dangerouslySetInnerHTML={{ __html: desktopShellScript }} />
      </head>
      <body className="shell" style={{ fontFamily: "var(--font-serif), Georgia, serif" }}>
        <DesktopShellMark />
        <NextIntlClientProvider messages={messages}>
          <SiteHeader email={email} />
          {children}
          <footer className="site-foot">
            <Link href="/privacy">{t("nav.privacy")}</Link>
          </footer>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
