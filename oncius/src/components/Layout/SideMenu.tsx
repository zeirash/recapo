"use client"

import { useState } from 'react'
import { Box, Flex, Text, IconButton } from 'theme-ui'
import { useRouter, usePathname } from 'next/navigation'
import { useAuth } from '@/hooks/useAuth'

interface SideMenuProps {
  selectedMenu: string
  onMenuSelect: (menu: string) => void
}

const SideMenu = ({ selectedMenu, onMenuSelect }: SideMenuProps) => {
  const { user } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: '🏠', path: '/dashboard' },
    { id: 'products', label: 'Products', icon: '🛍️', path: '/products' },
    { id: 'orders', label: 'Orders', icon: '📦', path: '/orders' },
    { id: 'customers', label: 'Customers', icon: '👥', path: '/customers' },
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
        {/* Profile Account */}
        <Flex sx={{ alignItems: 'center', justifyContent: 'center' }}>
          <Box
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
            }}
          >
            {user?.name?.charAt(0) || 'U'}
          </Box>
        </Flex>
      </Box>
    </Box>
  )
}

export default SideMenu
