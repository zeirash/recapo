"use client"

import React, { ReactNode, useState, useEffect } from 'react'
import { Box, CircularProgress } from '@mui/material'
import { usePathname, useRouter } from 'next/navigation'
import { useQuery } from 'react-query'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import { getAuthToken, api } from '@/utils/api'
import SideMenu from './SideMenu'
import BottomNav from './BottomNav'
import type { Subscription } from '@/types'

interface LayoutProps {
  children: ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  const pathname = usePathname()
  const router = useRouter()
  const t = useTranslations()
  const [selectedMenu, setSelectedMenu] = useState('dashboard')
  const { user, isLoadingUser, userError, isSubscriptionRequired } = useAuth()
  const isSystem = user?.role === 'system'

  const { data: subRes } = useQuery('subscription', () => api.getSubscription(), { enabled: !!user && !isSystem })
  const subscription = subRes?.data as Subscription | undefined

  const warningDaysLeft = (() => {
    if (isSystem || !subscription) return null
    const endDate =
      subscription.status === 'trialing' && subscription.trial_ends_at
        ? new Date(subscription.trial_ends_at)
        : subscription.status === 'active'
        ? new Date(subscription.current_period_end)
        : null
    if (!endDate) return null
    const days = Math.ceil((endDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24))
    return days > 0 && days <= 3 ? days : null
  })()

  // Redirect to login if unauthenticated, or to billing if subscription is required
  useEffect(() => {
    if (!getAuthToken() || (!isLoadingUser && userError)) {
      router.replace('/login')
    } else if (isSubscriptionRequired && !pathname.startsWith('/subscription')) {
      router.replace('/subscription')
    }
  }, [isLoadingUser, userError, isSubscriptionRequired, pathname, router])

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

  if (!user && !isSubscriptionRequired) {
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
          {warningDaysLeft !== null && !pathname.startsWith('/subscription') && (
            <Box
              sx={{
                bgcolor: warningDaysLeft === 1 ? '#C62828' : '#E65100',
                color: 'white',
                px: '16px',
                py: '10px',
                fontSize: '13px',
                fontWeight: 500,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '12px',
              }}
            >
              {t('layout.expiringSoon', { days: warningDaysLeft })}
              <Link href="/subscription" style={{ color: 'white', fontWeight: 700, textDecoration: 'underline' }}>
                {t('layout.subscribeNow')}
              </Link>
            </Box>
          )}
          {children}
        </Box>
      </Box>

      {/* Bottom nav — mobile only */}
      <BottomNav selectedMenu={selectedMenu} onMenuSelect={setSelectedMenu} />
    </Box>
  )
}

export default Layout
