import { getRequestConfig } from 'next-intl/server';

export const locales = ['en', 'id'] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = 'en';

export default getRequestConfig(async ({ locale }) => {
  const resolvedLocale = locale ?? defaultLocale;
  return {
    locale: resolvedLocale,
    messages: (await import(`../messages/${resolvedLocale}.json`)).default
  };
});
