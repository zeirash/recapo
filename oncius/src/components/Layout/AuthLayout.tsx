"use client"

import React, { ReactNode, useEffect } from 'react'
import { Box } from '@mui/material'
import { ThemeProvider } from '@mui/material/styles'
import { useRouter } from 'next/navigation'
import { createAppTheme } from '@/theme'
import Header from './Header'

interface AuthLayoutProps {
  children: ReactNode
}

const lightTheme = createAppTheme('light')

const AuthLayout = ({ children }: AuthLayoutProps) => {
  const router = useRouter()

  useEffect(() => {
    if (localStorage.getItem('authToken')) {
      router.replace('/dashboard')
    }
  }, [router])

  return (
    <ThemeProvider theme={lightTheme}>
      <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
        <Header />
        <Box component="main" sx={{ py: '24px' }}>
          {children}
        </Box>
      </Box>
    </ThemeProvider>
  )
}

export default AuthLayout
