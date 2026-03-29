"use client"

import React, { useEffect } from 'react'
import { Box, CircularProgress } from '@mui/material'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { getAuthToken } from '@/utils/api'
import SystemSideMenu from '@/components/SystemLayout/SystemSideMenu'
import SystemBottomNav from '@/components/SystemLayout/SystemBottomNav'

export default function SystemLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const { user, isLoadingUser, userError } = useAuth()

  useEffect(() => {
    if (!getAuthToken() || (!isLoadingUser && userError)) {
      router.replace('/login')
      return
    }
    if (!isLoadingUser && user && user.role !== 'system') {
      router.replace('/dashboard')
    }
  }, [isLoadingUser, userError, user, router])

  if (!user) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
      <Box sx={{ display: 'flex', height: '100vh' }}>
        <Box sx={{ display: { xs: 'none', sm: 'flex' } }}>
          <SystemSideMenu />
        </Box>
        <Box component="main" sx={{ flex: 1, bgcolor: 'background.default', overflowY: 'auto', pb: { xs: '64px', sm: 0 } }}>
          {children}
        </Box>
      </Box>
      <SystemBottomNav />
    </Box>
  )
}
