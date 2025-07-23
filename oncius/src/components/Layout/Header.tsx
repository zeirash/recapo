"use client"

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Box, Flex, Text, Button, IconButton } from 'theme-ui'
import { useAuth } from '@/hooks/useAuth'

const Header = () => {
  const { user, logout, isAuthenticated } = useAuth()
  const router = useRouter()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  const navigationItems = [
    { href: '/dashboard', label: 'Dashboard' },
    { href: '/products', label: 'Products' },
    { href: '/orders', label: 'Orders' },
    { href: '/customers', label: 'Customers' },
  ]

  const handleLogout = () => {
    logout()
  }

  return (
    <Box as="header" sx={{ bg: 'background', borderBottom: '1px solid', borderColor: 'border', py: 2 }}>
      <Flex sx={{ maxWidth: 'container', mx: 'auto', px: 3, alignItems: 'center', justifyContent: 'space-between' }}>
        {/* Logo */}
        <Link href="/">
          <Text sx={{ fontSize: 3, fontWeight: 'bold', color: 'primary', textDecoration: 'none' }}>
            Recapo
          </Text>
        </Link>

        {/* Desktop Navigation */}
        <Flex sx={{ display: ['none', 'flex'], gap: 4, alignItems: 'center' }}>
          {isAuthenticated && (
            <>
              {navigationItems.map((item) => (
                <Link key={item.href} href={item.href}>
                  <Text sx={{ color: 'text', textDecoration: 'none', '&:hover': { color: 'primary' } }}>
                    {item.label}
                  </Text>
                </Link>
              ))}
            </>
          )}
        </Flex>

        {/* User Menu / Auth Buttons */}
        <Flex sx={{ alignItems: 'center', gap: 2 }}>
          {isAuthenticated ? (
            <Flex sx={{ alignItems: 'center', gap: 2 }}>
              <Box sx={{ display: ['none', 'block'] }}>
                <Text sx={{ color: 'text.secondary' }}>Welcome, {user?.name}</Text>
              </Box>
              <Button variant="secondary" onClick={() => router.push('/profile')}>
                Profile
              </Button>
              <Button variant="secondary" onClick={handleLogout}>
                Logout
              </Button>
            </Flex>
          ) : (
            <Flex sx={{ gap: 2 }}>
              <Link href="/login">
                <Button variant="secondary">
                  Login
                </Button>
              </Link>
              <Link href="/register">
                <Button variant="primary">
                  Register
                </Button>
              </Link>
            </Flex>
          )}

          {/* Mobile Menu Button */}
          <IconButton
            sx={{ display: ['block', 'none'] }}
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            aria-label="Toggle mobile menu"
          >
            <Box sx={{ width: 20, height: 2, bg: 'text', mb: 1 }} />
            <Box sx={{ width: 20, height: 2, bg: 'text', mb: 1 }} />
            <Box sx={{ width: 20, height: 2, bg: 'text' }} />
          </IconButton>
        </Flex>
      </Flex>

      {/* Mobile Navigation */}
      {isMobileMenuOpen && (
        <Box sx={{ display: ['block', 'none'], bg: 'background.secondary', borderTop: '1px solid', borderColor: 'border' }}>
          <Flex sx={{ flexDirection: 'column', py: 2 }}>
            {isAuthenticated && (
              <>
                {navigationItems.map((item) => (
                  <Link key={item.href} href={item.href}>
                    <Text
                      sx={{
                        display: 'block',
                        px: 3,
                        py: 2,
                        color: 'text',
                        textDecoration: 'none',
                        '&:hover': { bg: 'background', color: 'primary' },
                      }}
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      {item.label}
                    </Text>
                  </Link>
                ))}
                <Box sx={{ borderTop: '1px solid', borderColor: 'border', mt: 2, pt: 2 }}>
                  <Text
                    sx={{
                      display: 'block',
                      px: 3,
                      py: 2,
                      color: 'text',
                      cursor: 'pointer',
                      '&:hover': { bg: 'background', color: 'error' },
                    }}
                    onClick={() => {
                      handleLogout()
                      setIsMobileMenuOpen(false)
                    }}
                  >
                    Logout
                  </Text>
                </Box>
              </>
            )}
          </Flex>
        </Box>
      )}
    </Box>
  )
}

export default Header
