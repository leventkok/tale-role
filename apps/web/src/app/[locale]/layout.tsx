import { Cinzel, Outfit, Source_Serif_4 } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { notFound } from "next/navigation";
import { isLocale, locales } from "@tale-role/i18n";
import { apiBase, getSessionToken } from "@/lib/session";
import { SiteHeader } from "@/components/site-header";
import "../globals.css";

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
  const token = await getSessionToken();
  let email: string | null = null;
  if (token) {
    const me = await fetch(`${apiBase()}/api/v1/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (me.ok) {
      email = ((await me.json()) as { email?: string }).email ?? null;
    }
  }

  return (
    <html lang={locale} data-theme="high-fantasy" className={`${display.variable} ${body.variable} ${sans.variable}`}>
      <body className="shell" style={{ fontFamily: "var(--font-serif), Georgia, serif" }}>
        <NextIntlClientProvider messages={messages}>
          <SiteHeader email={email} />
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
