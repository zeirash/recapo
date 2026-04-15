"use client"

import { useState } from 'react'
import { Box, Button } from '@mui/material'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useTheme } from '@mui/material/styles'
import { alpha } from '@mui/material/styles'
import { LayoutDashboard, Package, ClipboardList, ShoppingCart, ShoppingBag, Users, CreditCard, User, Moon, Sun, MessageSquare, LogOut, Settings, UserCog, type LucideIcon } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { USER_ROLES } from '@/constants/roles'
import { useThemeMode } from '@/providers/ThemeProvider'
import FeedbackDialog from '@/components/ui/FeedbackDialog'

interface BottomNavProps {
  selectedMenu: string
  onMenuSelect: (menu: string) => void
}

const baseMenuItems: { id: string; labelKey: string; icon: LucideIcon; path: string; ownerOnly?: boolean }[] = [
  { id: 'dashboard',    labelKey: 'dashboard',   icon: LayoutDashboard, path: '/dashboard' },
  { id: 'products',     labelKey: 'products',    icon: Package,         path: '/products' },
  { id: 'orders',       labelKey: 'orders',      icon: ClipboardList,   path: '/orders' },
  { id: 'temp_orders',  labelKey: 'tempOrders',  icon: ShoppingCart,    path: '/temp-orders' },
  { id: 'purchase',     labelKey: 'purchase',    icon: ShoppingBag,     path: '/purchase' },
  { id: 'customers',    labelKey: 'customers',   icon: Users,           path: '/customers' },
  { id: 'subscription', labelKey: 'subscription',icon: CreditCard,      path: '/subscription' },
  { id: 'admin',        labelKey: 'admin',       icon: UserCog,        path: '/admin', ownerOnly: true },
]

export default function BottomNav({ selectedMenu, onMenuSelect }: BottomNavProps) {
  const t = useTranslations('nav')
  const router = useRouter()
  const theme = useTheme()
  const { user, logout } = useAuth()
  const menuItems = baseMenuItems.filter(item => !item.ownerOnly || user?.role === USER_ROLES.OWNER)
  const { mode, toggleTheme } = useThemeMode()
  const [profileOpen, setProfileOpen] = useState(false)
  const [showFeedbackDialog, setShowFeedbackDialog] = useState(false)
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)
  const tCommon = useTranslations('common')

  return (
    <>
      <Box
        sx={{
          display: { xs: 'flex', sm: 'none' },
          position: 'fixed',
          bottom: 0,
          left: 0,
          right: 0,
          zIndex: 100,
          bgcolor: 'background.paper',
          borderTop: '1px solid',
          borderColor: 'divider',
          overflowX: 'auto',
          scrollbarWidth: 'none',
          '&::-webkit-scrollbar': { display: 'none' },
        }}
      >
        {menuItems.map((item) => {
          const isActive = selectedMenu === item.id
          return (
            <Box
              key={item.id}
              onClick={() => { onMenuSelect(item.id); router.push(item.path) }}
              sx={{
                flex: '0 0 auto',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '3px',
                px: '14px',
                py: '10px',
                cursor: 'pointer',
                color: isActive ? 'primary.main' : 'grey.500',
                bgcolor: isActive ? alpha(theme.palette.primary.main, 0.07) : 'transparent',
                borderTop: '2px solid',
                borderColor: isActive ? 'primary.main' : 'transparent',
                transition: 'color 0.15s, background-color 0.15s',
                minWidth: '64px',
              }}
            >
              <item.icon size={20} />
              <Box sx={{ fontSize: '10px', fontWeight: isActive ? 600 : 400, whiteSpace: 'nowrap', lineHeight: 1 }}>
                {t(item.labelKey)}
              </Box>
            </Box>
          )
        })}

        {/* Profile tab */}
        <Box
          onClick={() => setProfileOpen(true)}
          sx={{
            flex: '0 0 auto',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '3px',
            px: '14px',
            py: '10px',
            cursor: 'pointer',
            color: 'grey.500',
            borderTop: '2px solid transparent',
            minWidth: '64px',
          }}
        >
          <User size={20} />
          <Box sx={{ fontSize: '10px', whiteSpace: 'nowrap', lineHeight: 1 }}>{t('profile')}</Box>
        </Box>
      </Box>

      {/* Profile bottom sheet */}
      {profileOpen && (
        <>
          <Box
            sx={{ position: 'fixed', inset: 0, zIndex: 200, bgcolor: 'rgba(0,0,0,0.4)' }}
            onClick={() => setProfileOpen(false)}
          />
          <Box
            sx={{
              position: 'fixed',
              bottom: 0,
              left: 0,
              right: 0,
              zIndex: 201,
              bgcolor: 'background.paper',
              borderRadius: '16px 16px 0 0',
              pb: 'env(safe-area-inset-bottom)',
              display: { xs: 'block', sm: 'none' },
            }}
          >
            {/* Handle */}
            <Box sx={{ display: 'flex', justifyContent: 'center', pt: '12px', pb: '8px' }}>
              <Box sx={{ width: '36px', height: '4px', borderRadius: '2px', bgcolor: 'grey.300' }} />
            </Box>
            {/* Email */}
            <Box sx={{ px: '20px', py: '12px', borderBottom: '1px solid', borderColor: 'divider' }}>
              <Box sx={{ fontSize: '11px', color: 'text.secondary', mb: '2px' }}>{t('signedInAs')}</Box>
              <Box sx={{ fontSize: '14px', fontWeight: 500, wordBreak: 'break-all' }}>{user?.email}</Box>
            </Box>
            {/* Actions */}
            {[
              ...(user?.role === 'system' ? [{
                icon: <Settings size={18} />,
                label: 'System Dashboard',
                onClick: () => { setProfileOpen(false); router.push('/system') },
                color: 'text.primary',
              }] : []),
              {
                icon: <MessageSquare size={18} />,
                label: t('feedback'),
                onClick: () => { setProfileOpen(false); setShowFeedbackDialog(true) },
                color: 'text.primary',
              },
              {
                icon: mode === 'dark' ? <Sun size={18} /> : <Moon size={18} />,
                label: mode === 'dark' ? t('lightMode') : t('darkMode'),
                onClick: () => { toggleTheme(); setProfileOpen(false) },
                color: 'text.primary',
              },
              {
                icon: <LogOut size={18} />,
                label: t('logout'),
                onClick: () => { setProfileOpen(false); setShowLogoutDialog(true) },
                color: 'error.main',
              },
            ].map((action) => (
              <Box
                key={action.label}
                onClick={action.onClick}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '14px',
                  px: '20px',
                  py: '16px',
                  cursor: 'pointer',
                  color: action.color,
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  '&:hover': { bgcolor: 'action.hover' },
                }}
              >
                {action.icon}
                <Box sx={{ fontSize: '15px' }}>{action.label}</Box>
              </Box>
            ))}
            <Box sx={{ height: '8px' }} />
          </Box>
        </>
      )}

      <FeedbackDialog open={showFeedbackDialog} onClose={() => setShowFeedbackDialog(false)} />

      {/* Logout confirmation */}
      {showLogoutDialog && (
        <Box
          sx={{ position: 'fixed', inset: 0, bgcolor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 300, p: '16px' }}
          onClick={() => setShowLogoutDialog(false)}
        >
          <Box
            sx={{ bgcolor: 'background.paper', borderRadius: '12px', p: '24px', width: '100%', maxWidth: 360 }}
            onClick={(e) => e.stopPropagation()}
          >
            <Box sx={{ fontSize: '18px', fontWeight: 600, mb: '8px' }}>{t('logoutConfirmTitle')}</Box>
            <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '24px' }}>{t('logoutConfirmMessage')}</Box>
            <Box sx={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
              <Button variant="outlined" onClick={() => setShowLogoutDialog(false)}>{tCommon('cancel')}</Button>
              <Button variant="contained" disableElevation onClick={async () => { setShowLogoutDialog(false); await logout() }}>{t('logout')}</Button>
            </Box>
          </Box>
        </Box>
      )}
    </>
  )
}
