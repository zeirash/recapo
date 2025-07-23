"use client"

import { useState } from 'react'
import { Box, Flex, Text } from 'theme-ui'
import { useAuth } from '@/hooks/useAuth'

interface SideMenuProps {
  selectedMenu: string
  onMenuSelect: (menu: string) => void
}

const SideMenu = ({ selectedMenu, onMenuSelect }: SideMenuProps) => {
  const { user } = useAuth()

  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: '🏠' },
    { id: 'products', label: 'Products', icon: '🛍️' },
    { id: 'orders', label: 'Orders', icon: '📦' },
    { id: 'customers', label: 'Customers', icon: '👥' },
  ]

  return (
    <Box
      sx={{
        width: '280px',
        bg: 'background',
        borderRight: '1px solid',
        borderColor: 'border',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Top Section */}
      <Box sx={{ p: 3, borderBottom: '1px solid', borderColor: 'border' }}>
        {/* Logo/Icon */}
        <Box
          sx={{
            width: '40px',
            height: '40px',
            borderRadius: '50%',
            bg: 'primary',
            mb: 3,
          }}
        />
      </Box>

      {/* Menu Items */}
      <Box sx={{ flex: 1, overflowY: 'auto', p: 2 }}>
        <Text sx={{ fontSize: 0, color: 'text.secondary', mb: 2, px: 1 }}>
          MENU
        </Text>

        {menuItems.map((item) => (
          <Box
            key={item.id}
            sx={{
              p: 2,
              mb: 1,
              borderRadius: 'medium',
              cursor: 'pointer',
              bg: selectedMenu === item.id ? 'primary.light' : 'transparent',
              '&:hover': {
                bg: selectedMenu === item.id ? 'primary.light' : 'background.light',
              },
            }}
            onClick={() => onMenuSelect(item.id)}
          >
            <Flex sx={{ alignItems: 'center', gap: 2 }}>
              <Box sx={{ fontSize: 2 }}>{item.icon}</Box>
              <Text sx={{ fontWeight: 'medium', fontSize: 1 }}>
                {item.label}
              </Text>
            </Flex>
          </Box>
        ))}
      </Box>

      {/* Bottom Section */}
      <Box sx={{ p: 3, borderTop: '1px solid', borderColor: 'border' }}>
        {/* Profile Account */}
        <Flex sx={{ alignItems: 'center', gap: 2 }}>
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
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Text sx={{ fontWeight: 'medium', fontSize: 1 }}>
              {user?.name || 'User'}
            </Text>
          </Box>
        </Flex>
      </Box>
    </Box>
  )
}

export default SideMenu
