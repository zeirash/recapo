"use client"

import React, { ReactNode } from 'react'
import { Box } from '@mui/material'
import Header from './Header'

interface AuthLayoutProps {
  children: ReactNode
}

const AuthLayout = ({ children }: AuthLayoutProps) => {
  return (
    <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
      <Header />
      <Box component="main" sx={{ py: '24px' }}>
        {children}
      </Box>
    </Box>
  )
}

export default AuthLayout
