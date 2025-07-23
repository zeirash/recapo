"use client"

import { Inter } from 'next/font/google'
import { ThemeProvider } from 'theme-ui'
import { QueryClient, QueryClientProvider } from 'react-query'
import { ReactQueryDevtools } from 'react-query/devtools'
import theme from '@/styles/theme'
import '@/styles/globals.css'

const inter = Inter({ subsets: ['latin'] })

// Note: metadata export won't work in client components
// You'll need to handle SEO differently (e.g., using next/head in individual pages)

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
  return (
    <html lang="en">
      <body className={inter.className}>
        <QueryClientProvider client={queryClient}>
          <ThemeProvider theme={theme}>
            {children}
            <ReactQueryDevtools initialIsOpen={false} />
          </ThemeProvider>
        </QueryClientProvider>
      </body>
    </html>
  )
}
