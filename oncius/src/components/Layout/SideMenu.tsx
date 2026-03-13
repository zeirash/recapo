"use client"

import { useState, Fragment, useEffect } from 'react'
import { Box, Button, Tooltip } from '@mui/material'
import { useRouter } from 'next/navigation'
import { useLocale, useTranslations } from 'next-intl'
import { useQuery } from 'react-query'
import { useAuth } from '@/hooks/useAuth'
import { useChangeLocale } from '@/hooks/useLocale'
import { LayoutDashboard, ShoppingBag, Package, ClipboardList, ShoppingCart, Users, CreditCard, MessageSquare, LogOut, User, type LucideIcon } from 'lucide-react'
import RecapoLogo from '@/components/ui/RecapoLogo'
import { api } from '@/utils/api'
import type { Subscription } from '@/types'

interface SideMenuProps {
  selectedMenu: string
  onMenuSelect: (menu: string) => void
}

const SideMenu = ({ selectedMenu, onMenuSelect }: SideMenuProps) => {
  const t = useTranslations('nav')
  const tCommon = useTranslations('common')
  const { user, logout } = useAuth()
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)
  const [showDropdown, setShowDropdown] = useState(false)
  const router = useRouter()
  const locale = useLocale()
  const changeLocale = useChangeLocale()

  const { data: subRes } = useQuery('subscription', () => api.getSubscription(), {
    staleTime: 5 * 60 * 1000,
  })
  const subscription: Subscription | null = subRes?.data ?? null
  const trialExpired = subscription?.status === 'trialing' && !!subscription.trial_ends_at && new Date(subscription.trial_ends_at) < new Date()
  const periodExpired = subscription?.status === 'active' && new Date(subscription.current_period_end) < new Date()
  const isLocked = !!subscription && (['expired', 'past_due', 'cancelled'].includes(subscription.status) || trialExpired || periodExpired)
  const menuItems: { id: string; label: string; icon: LucideIcon; path: string }[] = [
    { id: 'dashboard', label: t('dashboard'), icon: LayoutDashboard, path: '/dashboard' },
    { id: 'products', label: t('products'), icon: Package, path: '/products' },
    { id: 'orders', label: t('orders'), icon: ClipboardList, path: '/orders' },
    { id: 'temp_orders', label: t('tempOrders'), icon: ShoppingCart, path: '/temp-orders' },
    { id: 'purchase', label: t('purchase'), icon: ShoppingBag, path: '/purchase' },
    { id: 'customers', label: t('customers'), icon: Users, path: '/customers' },
    { id: 'subscription', label: t('subscription'), icon: CreditCard, path: '/subscription' },
  ]

  useEffect(() => {
    menuItems.forEach(item => router.prefetch(item.path))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleMenuClick = (item: typeof menuItems[0]) => {
    onMenuSelect(item.id)
    router.push(item.path)
  }

  return (
    <Box
      sx={{
        width: '96px',
        bgcolor: 'white',
        borderRight: '1px solid',
        borderColor: 'grey.200',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'stretch',
      }}
    >
      {/* Top Section */}
      <Box sx={{ p: '16px', borderBottom: '1px solid', borderColor: 'grey.200' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }} onClick={() => router.push('/dashboard')}>
          {/* Logo/Icon */}
          <RecapoLogo />
        </Box>
      </Box>

      {/* Menu Items */}
      <Box sx={{ flex: 1, overflowY: 'auto', p: '4px' }}>
        {menuItems.map((item) => {
          const locked = isLocked && item.id !== 'subscription'
          const itemBox = (
            <Box
              sx={{
                py: '8px',
                px: '4px',
                mb: '4px',
                borderRadius: '8px',
                cursor: locked ? 'not-allowed' : 'pointer',
                textAlign: 'center',
                bgcolor: selectedMenu === item.id ? '#eff6ff' : 'transparent',
                '&:hover': {
                  bgcolor: locked ? 'transparent' : (selectedMenu === item.id ? '#eff6ff' : 'grey.100'),
                },
              }}
              onClick={locked ? undefined : () => handleMenuClick(item)}
            >
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: '4px', opacity: locked ? 0.4 : 1 }}>
                <item.icon size={20} />
                <Box sx={{ fontSize: '12px', lineHeight: 1, mt: '4px' }}>{item.label}</Box>
              </Box>
            </Box>
          )
          return locked ? (
            <Tooltip key={item.id} title="Subscribe to access" placement="right" arrow>
              {itemBox}
            </Tooltip>
          ) : (
            <Fragment key={item.id}>{itemBox}</Fragment>
          )
        })}
      </Box>

      {/* Bottom Section */}
      <Box sx={{ p: '16px', borderTop: '1px solid', borderColor: 'grey.200' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', mb: '16px' }}>
          <Box
            component="button"
            onClick={() => changeLocale('en')}
            sx={{
              bgcolor: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: '12px',
              fontWeight: locale === 'en' ? 600 : 400,
              color: locale === 'en' ? 'primary.main' : 'grey.500',
              textDecoration: 'none',
              p: 0,
              '&:hover': { color: 'primary.main' },
            }}
          >
            EN
          </Box>
          <Box sx={{ color: 'grey.200', fontSize: '12px' }}>|</Box>
          <Box
            component="button"
            onClick={() => changeLocale('id')}
            sx={{
              bgcolor: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: '12px',
              fontWeight: locale === 'id' ? 600 : 400,
              color: locale === 'id' ? 'primary.main' : 'grey.500',
              textDecoration: 'none',
              p: 0,
              '&:hover': { color: 'primary.main' },
            }}
          >
            ID
          </Box>
        </Box>
        {/* Profile Account */}
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Box
            onClick={() => setShowDropdown(true)}
            sx={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              background: 'linear-gradient(135deg,rgb(92, 151, 245) 0%,rgb(26, 94, 239) 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'white',
              cursor: 'pointer',
              '&:hover': { opacity: 0.9 },
            }}
          >
            <User size={16} />
          </Box>
        </Box>
      </Box>

      {/* Dropdown Menu */}
      {showDropdown && (
        <>
          <Box
            sx={{ position: 'fixed', inset: 0, zIndex: 999 }}
            onClick={() => setShowDropdown(false)}
          />
          <Box
            sx={{
              position: 'fixed',
              left: '70px',
              bottom: '16px',
              zIndex: 1000,
              bgcolor: 'white',
              borderRadius: '8px',
              boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05)',
              minWidth: '180px',
              overflow: 'hidden',
            }}
          >
            <Box
              onClick={() => setShowDropdown(false)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                py: '10px',
                px: '16px',
                cursor: 'pointer',
                '&:hover': { bgcolor: 'grey.50' },
              }}
            >
              <MessageSquare size={16} />
              <Box sx={{ fontSize: '14px' }}>{t('feedback')}</Box>
            </Box>
            <Box
              onClick={() => {
                setShowDropdown(false)
                setShowLogoutDialog(true)
              }}
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                py: '10px',
                px: '16px',
                cursor: 'pointer',
                color: 'error.main',
                '&:hover': { bgcolor: 'grey.50' },
              }}
            >
              <LogOut size={16} />
              <Box sx={{ fontSize: '14px' }}>{t('logout')}</Box>
            </Box>
          </Box>
        </>
      )}

      {/* Logout Dialog */}
      {showLogoutDialog && (
        <Box
          sx={{
            position: 'fixed',
            inset: 0,
            bgcolor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
          }}
          onClick={() => setShowLogoutDialog(false)}
        >
          <Box
            sx={{
              bgcolor: 'white',
              borderRadius: '12px',
              p: '24px',
              maxWidth: 360,
              boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <Box sx={{ fontSize: '18px', fontWeight: 600, mb: '8px', display: 'block' }}>
              {t('logoutConfirmTitle')}
            </Box>
            <Box sx={{ fontSize: '14px', color: 'grey.500', mb: '24px' }}>
              {t('logoutConfirmMessage')}
            </Box>
            <Box sx={{ display: 'flex', gap: '16px', justifyContent: 'flex-end' }}>
              <Button variant="outlined" onClick={() => setShowLogoutDialog(false)}>
                {tCommon('cancel')}
              </Button>
              <Button
                variant="contained"
                disableElevation
                onClick={() => {
                  setShowLogoutDialog(false)
                  logout()
                }}
              >
                {t('logout')}
              </Button>
            </Box>
          </Box>
        </Box>
      )}
    </Box>
  )
}

export default SideMenu
