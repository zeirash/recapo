"use client"

import React, { ReactNode, useState, useEffect } from 'react'
import { Box, CircularProgress } from '@mui/material'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { getAuthToken } from '@/utils/api'
import SideMenu from './SideMenu'

interface LayoutProps {
  children: ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  const pathname = usePathname()
  const router = useRouter()
  const [selectedMenu, setSelectedMenu] = useState('dashboard')
  const { user, isLoadingUser, userError } = useAuth()

  // Redirect to login if unauthenticated
  useEffect(() => {
    if (!getAuthToken() || (!isLoadingUser && userError)) {
      router.replace('/login')
    }
  }, [isLoadingUser, userError, router])

  // Sync selectedMenu with current pathname
  useEffect(() => {
    const pathToMenuMap: Record<string, string> = {
      '/dashboard': 'dashboard',
      '/products': 'products',
      '/orders': 'orders',
      '/temp-orders': 'temp_orders',
      '/purchase': 'purchase',
      '/customers': 'customers',
      '/subscription': 'subscription',
    }

    const menuId = pathToMenuMap[pathname]
    if (menuId) {
      setSelectedMenu(menuId)
    }
  }, [pathname])

  if (!user) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: '#f9fafb' }}>
      <Box sx={{ display: 'flex', height: '100vh' }}>
        <SideMenu
          selectedMenu={selectedMenu}
          onMenuSelect={setSelectedMenu}
        />
        <Box component="main" sx={{ flex: 1, bgcolor: 'grey.50', overflowY: 'auto' }}>
          {children}
        </Box>
      </Box>
    </Box>
  )
}

export default Layout
