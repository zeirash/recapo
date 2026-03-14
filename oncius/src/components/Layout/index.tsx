"use client"

import React, { ReactNode, useState, useEffect } from 'react'
import { Box } from '@mui/material'
import { usePathname } from 'next/navigation'
import SideMenu from './SideMenu'

interface LayoutProps {
  children: ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  const pathname = usePathname()
  const [selectedMenu, setSelectedMenu] = useState('dashboard')

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
