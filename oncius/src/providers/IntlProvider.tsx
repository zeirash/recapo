'use client';

import { NextIntlClientProvider } from 'next-intl';
import { ReactNode } from 'react';

type Props = {
  locale: string;
  messages: Record<string, unknown>;
  children: ReactNode;
};

export default function IntlProvider({ locale, messages, children }: Props) {
  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      {children}
    </NextIntlClientProvider>
  );
}
