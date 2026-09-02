import { defineRouting } from "next-intl/routing";
import { createNavigation } from "next-intl/navigation";
import { defaultLocale, locales } from "@tale-role/i18n";

export const routing = defineRouting({
  locales: [...locales],
  defaultLocale,
});

export const { Link, redirect, usePathname, useRouter } = createNavigation(routing);
