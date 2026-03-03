"use client"

import React, { ReactNode, useState, useEffect } from 'react'
import { Box } from '@mui/material'
import { usePathname } from 'next/navigation'
import SideMenu from './SideMenu'
import Header from './Header'

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
    }

    const menuId = pathToMenuMap[pathname]
    if (menuId) {
      setSelectedMenu(menuId)
    }
  }, [pathname])

  // Show header for login and register pages
  const isAuthPage = pathname === '/login' || pathname === '/register' || pathname === '/forgot-password'

  if (isAuthPage) {
    return (
      <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
        <Header />
        <Box component="main" sx={{ py: '24px' }}>
          {children}
        </Box>
      </Box>
    )
  }

  // Show sidebar for other pages
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
