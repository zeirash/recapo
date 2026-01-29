'use client';

import { useCallback } from 'react';

export function useChangeLocale() {
  const changeLocale = useCallback((newLocale: 'en' | 'id') => {
    localStorage.setItem('locale', newLocale);
    // Reload to apply new locale
    window.location.reload();
  }, []);

  return changeLocale;
}
