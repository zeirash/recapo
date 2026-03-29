"use client"

import { useState } from 'react'
import { Box, Button } from '@mui/material'
import { useRouter, usePathname } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { useThemeMode } from '@/providers/ThemeProvider'
import { LayoutDashboard, Store, CreditCard, User, Moon, Sun, ArrowLeftRight, LogOut, type LucideIcon } from 'lucide-react'
import { alpha, useTheme } from '@mui/material/styles'

const menuItems: { id: string; label: string; icon: LucideIcon; path: string }[] = [
  { id: 'system',   label: 'Overview', icon: LayoutDashboard, path: '/system' },
  { id: 'shops',    label: 'Shops',    icon: Store,           path: '/system/shops' },
  { id: 'payments', label: 'Payments', icon: CreditCard,      path: '/system/payments' },
]

export default function SystemBottomNav() {
  const router = useRouter()
  const pathname = usePathname()
  const theme = useTheme()
  const { user, logout } = useAuth()
  const { mode, toggleTheme } = useThemeMode()
  const [profileOpen, setProfileOpen] = useState(false)
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)

  const selectedMenu = menuItems.find(item => pathname === item.path)?.id ?? 'system'

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
          justifyContent: 'center',
        }}
      >
        {menuItems.map((item) => {
          const isActive = selectedMenu === item.id
          return (
            <Box
              key={item.id}
              onClick={() => router.push(item.path)}
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
                color: isActive ? 'warning.main' : 'grey.500',
                bgcolor: isActive ? alpha(theme.palette.warning.main, 0.07) : 'transparent',
                borderTop: '2px solid',
                borderColor: isActive ? 'warning.main' : 'transparent',
                transition: 'color 0.15s, background-color 0.15s',
                minWidth: '64px',
              }}
            >
              <item.icon size={20} />
              <Box sx={{ fontSize: '10px', fontWeight: isActive ? 600 : 400, whiteSpace: 'nowrap', lineHeight: 1 }}>
                {item.label}
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
          <Box sx={{ fontSize: '10px', whiteSpace: 'nowrap', lineHeight: 1 }}>Profile</Box>
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
            {/* Header */}
            <Box sx={{ px: '20px', py: '12px', borderBottom: '1px solid', borderColor: 'divider' }}>
              <Box sx={{ fontSize: '11px', color: 'warning.main', fontWeight: 600, mb: '2px' }}>SYSTEM MODE</Box>
              <Box sx={{ fontSize: '14px', fontWeight: 500, wordBreak: 'break-all' }}>{user?.email}</Box>
            </Box>
            {/* Actions */}
            {[
              {
                icon: <ArrowLeftRight size={18} />,
                label: 'Go to Merchant',
                onClick: () => { setProfileOpen(false); router.push('/dashboard') },
                color: 'text.primary',
              },
              {
                icon: mode === 'dark' ? <Sun size={18} /> : <Moon size={18} />,
                label: mode === 'dark' ? 'Light Mode' : 'Dark Mode',
                onClick: () => { toggleTheme(); setProfileOpen(false) },
                color: 'text.primary',
              },
              {
                icon: <LogOut size={18} />,
                label: 'Logout',
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
            <Box sx={{ fontSize: '18px', fontWeight: 600, mb: '8px' }}>Logout?</Box>
            <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '24px' }}>You will be signed out.</Box>
            <Box sx={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
              <Button variant="outlined" onClick={() => setShowLogoutDialog(false)}>Cancel</Button>
              <Button variant="contained" color="error" disableElevation onClick={async () => { setShowLogoutDialog(false); await logout() }}>Logout</Button>
            </Box>
          </Box>
        </Box>
      )}
    </>
  )
}
