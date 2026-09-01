import en from "../messages/en.json";
import tr from "../messages/tr.json";
import { type Locale, defaultLocale } from "./locales";

export type MessageCatalog = typeof en;

const catalogs: Record<Locale, MessageCatalog> = { en, tr };

export function messagesFor(locale: Locale): MessageCatalog {
  return catalogs[locale] ?? catalogs[defaultLocale];
}

export { catalogs };
