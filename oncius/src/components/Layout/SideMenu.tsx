"use client"

import { useState } from 'react'
import { Box, Flex, Text, Button } from 'theme-ui'
import { useRouter, usePathname } from 'next/navigation'
import { useLocale, useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import { useChangeLocale } from '@/hooks/useLocale'

interface SideMenuProps {
  selectedMenu: string
  onMenuSelect: (menu: string) => void
}

const SideMenu = ({ selectedMenu, onMenuSelect }: SideMenuProps) => {
  const t = useTranslations('nav')
  const { user, logout } = useAuth()
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)
  const router = useRouter()
  const pathname = usePathname()
  const locale = useLocale()
  const changeLocale = useChangeLocale()

  const menuItems = [
    { id: 'dashboard', label: t('dashboard'), icon: '🏠', path: '/dashboard' },
    { id: 'products', label: t('products'), icon: '🛍️', path: '/products' },
    { id: 'orders', label: t('orders'), icon: '📦', path: '/orders' },
    { id: 'temp_orders', label: t('tempOrders'), icon: '📋', path: '/temp_orders' },
    { id: 'purchase', label: t('purchase'), icon: '🛒', path: '/purchase' },
    { id: 'customers', label: t('customers'), icon: '👥', path: '/customers' },
  ]

  const handleMenuClick = (item: typeof menuItems[0]) => {
    onMenuSelect(item.id)
    router.push(item.path)
  }

  return (
    <Box
      sx={{
        width: '96px',
        bg: 'background',
        borderRight: '1px solid',
        borderColor: 'border',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'stretch',
      }}
    >
      {/* Top Section */}
      <Box sx={{ p: 3, borderBottom: '1px solid', borderColor: 'border' }}>
        <Flex sx={{ alignItems: 'center', justifyContent: 'center' }}>
          {/* Logo/Icon */}
          <Box
            sx={{
              width: '40px',
              height: '40px',
              borderRadius: '50%',
              bg: 'primary',
            }}
          />
        </Flex>
      </Box>

      {/* Menu Items */}
      <Box sx={{ flex: 1, overflowY: 'auto', p: 1 }}>
        {menuItems.map((item) => (
          <Box
            key={item.id}
            sx={{
              py: 2,
              px: 1,
              mb: 1,
              borderRadius: 'medium',
              cursor: 'pointer',
              textAlign: 'center',
              bg: selectedMenu === item.id ? 'primary.light' : 'transparent',
              '&:hover': {
                bg: selectedMenu === item.id ? 'primary.light' : 'background.light',
              },
            }}
            onClick={() => handleMenuClick(item)}
          >
            <Flex sx={{ flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
              <Box sx={{ fontSize: 3, lineHeight: 1 }}>{item.icon}</Box>
              <Text sx={{ fontSize: 0, lineHeight: 1, mt: 1 }}>{item.label}</Text>
            </Flex>
          </Box>
        ))}
      </Box>

      {/* Bottom Section */}
      <Box sx={{ p: 3, borderTop: '1px solid', borderColor: 'border' }}>
        <Flex sx={{ alignItems: 'center', justifyContent: 'center', gap: 2, mb: 3 }}>
          <Box
            as="button"
            onClick={() => changeLocale('en')}
            sx={{
              bg: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: 0,
              fontWeight: locale === 'en' ? 600 : 400,
              color: locale === 'en' ? 'primary' : 'text.secondary',
              textDecoration: 'none',
              p: 0,
              '&:hover': { color: 'primary' },
            }}
          >
            EN
          </Box>
          <Text sx={{ color: 'border', fontSize: 0 }}>|</Text>
          <Box
            as="button"
            onClick={() => changeLocale('id')}
            sx={{
              bg: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: 0,
              fontWeight: locale === 'id' ? 600 : 400,
              color: locale === 'id' ? 'primary' : 'text.secondary',
              textDecoration: 'none',
              p: 0,
              '&:hover': { color: 'primary' },
            }}
          >
            ID
          </Box>
        </Flex>
        {/* Profile Account */}
        <Flex sx={{ alignItems: 'center', justifyContent: 'center' }}>
          <Box
            onClick={() => setShowLogoutDialog(true)}
            sx={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              background: 'linear-gradient(135deg, #4CAF50, #81C784)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 1,
              color: 'white',
              fontWeight: 'bold',
              cursor: 'pointer',
              '&:hover': { opacity: 0.9 },
            }}
          >
            {user?.name?.charAt(0) || 'U'}
          </Box>
        </Flex>
      </Box>

      {/* Logout Dialog */}
      {showLogoutDialog && (
        <Box
          sx={{
            position: 'fixed',
            inset: 0,
            bg: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
          }}
          onClick={() => setShowLogoutDialog(false)}
        >
          <Box
            sx={{
              bg: 'white',
              borderRadius: 'large',
              p: 4,
              maxWidth: 360,
              boxShadow: 'large',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <Text sx={{ fontSize: 3, fontWeight: 600, mb: 2, display: 'block' }}>
              Log out?
            </Text>
            <Text sx={{ fontSize: 1, color: 'text.secondary', mb: 4 }}>
              Are you sure you want to log out?
            </Text>
            <Flex sx={{ gap: 3, justifyContent: 'flex-end' }}>
              <Button variant="secondary" onClick={() => setShowLogoutDialog(false)}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={() => {
                  setShowLogoutDialog(false)
                  logout()
                }}
              >
                Log out
              </Button>
            </Flex>
          </Box>
        </Box>
      )}
    </Box>
  )
}

export default SideMenu
