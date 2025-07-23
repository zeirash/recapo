"use client"

import React, { ReactNode, useState } from 'react'
import { Box, Flex } from 'theme-ui'
import { usePathname } from 'next/navigation'
import Header from './Header'
import SideMenu from './SideMenu'

interface LayoutProps {
  children: ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  const pathname = usePathname()
  const [selectedMenu, setSelectedMenu] = useState('dashboard')

  // Show header for login and register pages
  const isAuthPage = pathname === '/login' || pathname === '/register'

  if (isAuthPage) {
    return (
      <Box sx={{ minHeight: '100vh', bg: 'background.secondary' }}>
        <Header />
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
