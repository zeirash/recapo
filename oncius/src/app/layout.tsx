"use client"

import { Inter } from 'next/font/google'
import { ThemeProvider } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import { QueryClient, QueryClientProvider } from 'react-query'
import { NextIntlClientProvider } from 'next-intl'
import '@/styles/globals.css'
import { useEffect, useState } from 'react'
import theme from '@/theme'

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
  const [locale, setLocale] = useState('en')
  const [messages, setMessages] = useState<Record<string, unknown> | null>(null)

  useEffect(() => {
    // Get locale from localStorage or browser preference
    const savedLocale = localStorage.getItem('locale')
    const browserLocale = navigator.language.startsWith('id') ? 'id' : 'en'
    const currentLocale = savedLocale || browserLocale
    setLocale(currentLocale)

    // Load messages
    import(`../../messages/${currentLocale}.json`)
      .then((module) => setMessages(module.default))
      .catch(() => {
        // Fallback to English
        import('../../messages/en.json').then((module) => setMessages(module.default))
      })
  }, [])

  if (!messages) {
    return (
      <html lang={locale}>
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
      <body className={inter.className}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <QueryClientProvider client={queryClient}>
            <ThemeProvider theme={theme}>
              <CssBaseline />
              {children}
            </ThemeProvider>
          </QueryClientProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
