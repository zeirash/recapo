"use client"

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Box, Button, IconButton } from '@mui/material'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import LanguageSwitcher from '@/components/LanguageSwitcher'
import { Menu } from 'lucide-react'

const Header = () => {
  const t = useTranslations('nav')
  const { user, logout, isAuthenticated } = useAuth()
  const router = useRouter()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  const navigationItems = [
    { href: '/dashboard', label: t('dashboard') },
    { href: '/products', label: t('products') },
    { href: '/orders', label: t('orders') },
    { href: '/purchase', label: t('purchase') },
    { href: '/customers', label: t('customers') },
  ]

  const handleLogout = () => {
    logout()
  }

  return (
    <Box component="header" sx={{ bgcolor: 'white', borderBottom: '1px solid', borderColor: '#e5e7eb', py: '8px' }}>
      <Box sx={{ display: 'flex', maxWidth: 'container', mx: 'auto', px: '16px', alignItems: 'center', justifyContent: 'space-between' }}>
        {/* Logo */}
        <Link href="/">
          <Box sx={{ fontSize: '18px', fontWeight: 'bold', color: '#3b82f6', textDecoration: 'none' }}>
            Recapo
          </Box>
        </Link>

        {/* Desktop Navigation */}
        <Box sx={{ display: { xs: 'none', sm: 'flex' }, gap: '24px', alignItems: 'center' }}>
          {isAuthenticated && (
            <>
              {navigationItems.map((item) => (
                <Link key={item.href} href={item.href}>
                  <Box sx={{ color: '#1f2937', textDecoration: 'none', '&:hover': { color: '#3b82f6' } }}>
                    {item.label}
                  </Box>
                </Link>
              ))}
            </>
          )}
        </Box>

        {/* User Menu / Auth Buttons */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <LanguageSwitcher />
          {isAuthenticated ? (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <Box sx={{ display: { xs: 'none', sm: 'block' } }}>
                <Box sx={{ color: '#6b7280' }}>{t('welcome', { name: user?.name ?? '' })}</Box>
              </Box>
              <Button variant="outlined" onClick={() => router.push('/profile')}>
                {t('profile')}
              </Button>
              <Button variant="outlined" onClick={handleLogout}>
                {t('logout')}
              </Button>
            </Box>
          ) : (
            <Box sx={{ display: 'flex', gap: '8px' }}>
              <Link href="/login">
                <Button variant="outlined">
                  {t('login')}
                </Button>
              </Link>
              <Link href="/register">
                <Button variant="contained" disableElevation>
                  {t('register')}
                </Button>
              </Link>
            </Box>
          )}

          {/* Mobile Menu Button */}
          <IconButton
            sx={{ display: { xs: 'block', sm: 'none' } }}
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            aria-label="Toggle mobile menu"
          >
            <Menu size={20} />
          </IconButton>
        </Box>
      </Box>

      {/* Mobile Navigation */}
      {isMobileMenuOpen && (
        <Box sx={{ display: { xs: 'block', sm: 'none' }, bgcolor: '#f9fafb', borderTop: '1px solid', borderColor: '#e5e7eb' }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', py: '8px' }}>
            {isAuthenticated && (
              <>
                {navigationItems.map((item) => (
                  <Link key={item.href} href={item.href}>
                    <Box
                      sx={{
                        display: 'block',
                        px: '16px',
                        py: '8px',
                        color: '#1f2937',
                        textDecoration: 'none',
                        '&:hover': { bgcolor: 'white', color: '#3b82f6' },
                      }}
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      {item.label}
                    </Box>
                  </Link>
                ))}
                <Box sx={{ borderTop: '1px solid', borderColor: '#e5e7eb', mt: '8px', pt: '8px' }}>
                  <Box
                    sx={{
                      display: 'block',
                      px: '16px',
                      py: '8px',
                      color: '#1f2937',
                      cursor: 'pointer',
                      '&:hover': { bgcolor: 'white', color: '#ef4444' },
                    }}
                    onClick={() => {
                      handleLogout()
                      setIsMobileMenuOpen(false)
                    }}
                  >
                    {t('logout')}
                  </Box>
                </Box>
              </>
            )}
          </Box>
        </Box>
      )}
    </Box>
  )
}

export default Header
