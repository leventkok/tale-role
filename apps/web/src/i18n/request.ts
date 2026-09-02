import { getRequestConfig } from "next-intl/server";
import { defaultLocale, isLocale, messagesFor } from "@tale-role/i18n";
import { routing } from "./routing";

export default getRequestConfig(async ({ requestLocale }) => {
  const requested = await requestLocale;
  const locale = isLocale(requested) ? requested : routing.defaultLocale;
  return {
    locale: locale ?? defaultLocale,
    messages: messagesFor(locale ?? defaultLocale),
  };
});
