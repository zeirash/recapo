'use client';

import { useLocale } from 'next-intl';
import { useChangeLocale } from '@/hooks/useLocale';

export default function LanguageSwitcher() {
  const locale = useLocale();
  const changeLocale = useChangeLocale();

  return (
    <select
      value={locale}
      onChange={(e) => changeLocale(e.target.value as 'en' | 'id')}
      style={{
        padding: '4px 8px',
        borderRadius: '4px',
        border: '1px solid #ccc',
        cursor: 'pointer',
        fontSize: '14px',
      }}
    >
      <option value="en">English</option>
      <option value="id">Indonesia</option>
    </select>
  );
}
