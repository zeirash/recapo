"use client"

import { Inter } from 'next/font/google'
import { QueryClient, QueryClientProvider } from 'react-query'
import { NextIntlClientProvider } from 'next-intl'
import '@/styles/globals.css'
import { useEffect, useState } from 'react'
import { AppThemeProvider } from '@/providers/ThemeProvider'

const inter = Inter({ subsets: ['latin'] })

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const [locale, setLocale] = useState('id')
  const [messages, setMessages] = useState<Record<string, unknown> | null>(null)

  useEffect(() => {
    // Get locale from localStorage or browser preference
    const savedLocale = localStorage.getItem('locale')
    const browserLocale = navigator.language.startsWith('en') ? 'en' : 'id'
    const currentLocale = savedLocale || browserLocale
    setLocale(currentLocale)

    // Load messages
    import(`../../messages/${currentLocale}.json`)
      .then((module) => setMessages(module.default))
      .catch(() => {
        // Fallback to Indonesian
        import('../../messages/id.json').then((module) => setMessages(module.default))
      })
  }, [])

  if (!messages) {
    return (
      <html lang={locale}>
        <head>
          <meta name="viewport" content="width=device-width, initial-scale=1" />
        </head>
        <body className={inter.className}>
          <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
            Loading...
          </div>
        </body>
      </html>
    )
  }

  return (
    <html lang={locale}>
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
      </head>
      <body className={inter.className}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <QueryClientProvider client={queryClient}>
            <AppThemeProvider>
              {children}
            </AppThemeProvider>
          </QueryClientProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
