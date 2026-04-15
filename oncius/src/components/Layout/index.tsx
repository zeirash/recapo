"use client"

import React, { ReactNode, useState, useEffect } from 'react'
import { Box, CircularProgress } from '@mui/material'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { getAuthToken } from '@/utils/api'
import SideMenu from './SideMenu'
import BottomNav from './BottomNav'

interface LayoutProps {
  children: ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  const pathname = usePathname()
  const router = useRouter()
  const [selectedMenu, setSelectedMenu] = useState('dashboard')
  const { user, isLoadingUser, userError, isSubscriptionRequired } = useAuth()
  const isSubscriptionError = isSubscriptionRequired || (userError as any)?.status === 402

  // Redirect to login if unauthenticated (402 = billing issue, not auth — handled in useAuth)
  useEffect(() => {
    if (!getAuthToken() || (!isLoadingUser && userError && !isSubscriptionError)) {
      router.replace('/login')
    }
  }, [isLoadingUser, userError, isSubscriptionError, router])

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

  if (!user && !isSubscriptionError) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
      <Box sx={{ display: 'flex', height: '100vh' }}>
        {/* Sidebar — desktop only */}
        <Box sx={{ display: { xs: 'none', sm: 'flex' } }}>
          <SideMenu selectedMenu={selectedMenu} onMenuSelect={setSelectedMenu} />
        </Box>

        {/* Main content — extra bottom padding on mobile for bottom nav */}
        <Box
          component="main"
          sx={{
            flex: 1,
            bgcolor: 'background.default',
            overflowY: 'auto',
            pb: { xs: '64px', sm: 0 },
          }}
        >
          {children}
        </Box>
      </Box>

      {/* Bottom nav — mobile only */}
      <BottomNav selectedMenu={selectedMenu} onMenuSelect={setSelectedMenu} />
    </Box>
  )
}

export default Layout
