"use client"

import { useState } from 'react'
import { Box } from '@mui/material'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'
import { useThemeMode } from '@/providers/ThemeProvider'
import { LayoutDashboard, Store, CreditCard, LogOut, User, Moon, Sun, ArrowLeftRight } from 'lucide-react'
import { alpha, useTheme } from '@mui/material/styles'
import RecapoLogo from '@/components/ui/RecapoLogo'
import { usePathname } from 'next/navigation'

const menuItems = [
  { id: 'system', label: 'Overview', icon: LayoutDashboard, path: '/system' },
  { id: 'shops', label: 'Shops', icon: Store, path: '/system/shops' },
  { id: 'payments', label: 'Payments', icon: CreditCard, path: '/system/payments' },
]

export default function SystemSideMenu() {
  const { user, logout } = useAuth()
  const router = useRouter()
  const pathname = usePathname()
  const theme = useTheme()
  const { mode, toggleTheme } = useThemeMode()
  const [showDropdown, setShowDropdown] = useState(false)
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)

  const selectedMenu = menuItems.find(item => pathname === item.path)?.id ?? 'system'

  return (
    <Box
      sx={{
        width: '96px',
        bgcolor: 'background.paper',
        borderRight: '1px solid',
        borderColor: 'grey.200',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'stretch',
      }}
    >
      {/* Top */}
      <Box sx={{ p: '16px', borderBottom: '1px solid', borderColor: 'grey.200' }}>
        <Box
          sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
          onClick={() => router.push('/system')}
        >
          <RecapoLogo />
        </Box>
      </Box>

      {/* Menu Items */}
      <Box sx={{ flex: 1, overflowY: 'auto', p: '4px' }}>
        {menuItems.map((item) => (
          <Box
            key={item.id}
            onClick={() => router.push(item.path)}
            sx={{
              py: '8px',
              px: '4px',
              mb: '4px',
              borderRadius: '8px',
              cursor: 'pointer',
              textAlign: 'center',
              bgcolor: selectedMenu === item.id ? alpha(theme.palette.primary.main, 0.1) : 'transparent',
              '&:hover': { bgcolor: alpha(theme.palette.primary.main, selectedMenu === item.id ? 0.1 : 0.05) },
            }}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px' }}>
              <item.icon size={20} />
              <Box sx={{ fontSize: '12px', lineHeight: 1, mt: '4px' }}>{item.label}</Box>
            </Box>
          </Box>
        ))}
      </Box>

      {/* Bottom - Profile */}
      <Box sx={{ p: '16px', borderTop: '1px solid', borderColor: 'grey.200' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Box
            onClick={() => setShowDropdown(true)}
            sx={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
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

      {/* Dropdown */}
      {showDropdown && (
        <>
          <Box sx={{ position: 'fixed', inset: 0, zIndex: 999 }} onClick={() => setShowDropdown(false)} />
          <Box
            sx={{
              position: 'fixed',
              left: '70px',
              bottom: '16px',
              zIndex: 1000,
              bgcolor: 'background.paper',
              borderRadius: '8px',
              boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05)',
              minWidth: '180px',
              overflow: 'hidden',
            }}
          >
            <Box sx={{ px: '16px', py: '10px', borderBottom: '1px solid', borderColor: 'grey.100' }}>
              <Box sx={{ fontSize: '11px', color: 'warning.main', fontWeight: 600, mb: '2px' }}>SYSTEM MODE</Box>
              <Box sx={{ fontSize: '13px', fontWeight: 500, color: 'text.primary', wordBreak: 'break-all' }}>
                {user?.email}
              </Box>
            </Box>
            <Box
              onClick={() => { setShowDropdown(false); router.push('/dashboard') }}
              sx={{
                display: 'flex', alignItems: 'center', gap: '10px',
                py: '10px', px: '16px', cursor: 'pointer',
                '&:hover': { bgcolor: 'action.hover' },
              }}
            >
              <ArrowLeftRight size={16} />
              <Box sx={{ fontSize: '14px' }}>Go to Merchant</Box>
            </Box>
            <Box
              onClick={() => { toggleTheme(); setShowDropdown(false) }}
              sx={{
                display: 'flex', alignItems: 'center', gap: '10px',
                py: '10px', px: '16px', cursor: 'pointer',
                '&:hover': { bgcolor: 'action.hover' },
              }}
            >
              {mode === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
              <Box sx={{ fontSize: '14px' }}>{mode === 'dark' ? 'Light Mode' : 'Dark Mode'}</Box>
            </Box>
            <Box
              onClick={() => { setShowDropdown(false); setShowLogoutDialog(true) }}
              sx={{
                display: 'flex', alignItems: 'center', gap: '10px',
                py: '10px', px: '16px', cursor: 'pointer',
                color: 'error.main',
                '&:hover': { bgcolor: 'action.hover' },
              }}
            >
              <LogOut size={16} />
              <Box sx={{ fontSize: '14px' }}>Logout</Box>
            </Box>
          </Box>
        </>
      )}

      {/* Logout Dialog */}
      {showLogoutDialog && (
        <Box
          sx={{
            position: 'fixed', inset: 0, bgcolor: 'rgba(0,0,0,0.5)',
            display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000,
          }}
          onClick={() => setShowLogoutDialog(false)}
        >
          <Box
            sx={{ bgcolor: 'background.paper', borderRadius: '12px', p: '24px', maxWidth: 360, boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)' }}
            onClick={(e) => e.stopPropagation()}
          >
            <Box sx={{ fontSize: '18px', fontWeight: 600, mb: '8px' }}>Logout?</Box>
            <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '24px' }}>You will be signed out.</Box>
            <Box sx={{ display: 'flex', gap: '16px', justifyContent: 'flex-end' }}>
              <Box
                component="button"
                onClick={() => setShowLogoutDialog(false)}
                sx={{ px: '16px', py: '8px', borderRadius: '6px', border: '1px solid', borderColor: 'grey.300', cursor: 'pointer', bgcolor: 'transparent', fontSize: '14px' }}
              >
                Cancel
              </Box>
              <Box
                component="button"
                onClick={async () => { setShowLogoutDialog(false); await logout() }}
                sx={{ px: '16px', py: '8px', borderRadius: '6px', border: 'none', cursor: 'pointer', bgcolor: 'error.main', color: 'white', fontSize: '14px' }}
              >
                Logout
              </Box>
            </Box>
          </Box>
        </Box>
      )}
    </Box>
  )
}
