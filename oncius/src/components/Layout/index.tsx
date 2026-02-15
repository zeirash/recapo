"use client"

import React, { ReactNode, useState, useEffect } from 'react'
import { Box, Flex } from 'theme-ui'
import { usePathname } from 'next/navigation'
import Header from './Header'
import SideMenu from './SideMenu'
import LandingHeader from '@/components/LandingHeader'

interface LayoutProps {
  children: ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  const pathname = usePathname()
  const [selectedMenu, setSelectedMenu] = useState('dashboard')
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)

  // Sync selectedMenu with current pathname
  useEffect(() => {
    const pathToMenuMap: Record<string, string> = {
      '/dashboard': 'dashboard',
      '/products': 'products',
      '/orders': 'orders',
      '/temp_orders': 'temp_orders',
      '/customers': 'customers',
    }

    const menuId = pathToMenuMap[pathname]
    if (menuId) {
      setSelectedMenu(menuId)
    }
  }, [pathname])

  // Show header for login and register pages
  const isAuthPage = pathname === '/login' || pathname === '/register'

  if (isAuthPage) {
    return (
      <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
        <LandingHeader />
        <Box as="main" sx={{ py: 4 }}>
          {children}
        </Box>
      </Box>
    )
  }

  // Show sidebar for other pages
  return (
    <Box sx={{ minHeight: '100vh', bg: 'background.secondary' }}>
      <Flex sx={{ height: '100vh' }}>
        <SideMenu
          selectedMenu={selectedMenu}
          onMenuSelect={setSelectedMenu}
        />
        <Box as="main" sx={{ flex: 1, bg: 'white', overflowY: 'auto' }}>
          {children}
        </Box>
      </Flex>
    </Box>
  )
}

export default Layout
