"use client"

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Box, Button, IconButton, Typography } from '@mui/material'
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
    <Box component="header" sx={{ position: 'sticky', top: 0, zIndex: 10, bgcolor: 'rgba(255,255,255,0.9)', backdropFilter: 'saturate(180%) blur(10px)', borderBottom: '1px solid', borderColor: 'divider' }}>
      <Box sx={{ display: 'flex', mx: 'auto', px: { xs: 3, sm: 4 }, py: 2, alignItems: 'center', justifyContent: 'space-between' }}>
        {/* Logo */}
        <Link href="/" style={{ textDecoration: 'none' }}>
          <Typography sx={{ fontSize: { xs: '1.25rem', sm: '1.5rem' }, fontWeight: 700, color: 'primary.main' }}>
            Recapo
          </Typography>
        </Link>

        {/* Desktop Navigation */}
        <Box sx={{ display: { xs: 'none', sm: 'flex' }, gap: '24px', alignItems: 'center' }}>
          {isAuthenticated && (
            <>
              {navigationItems.map((item) => (
                <Link key={item.href} href={item.href}>
                  <Typography sx={{ color: 'grey.800', textDecoration: 'none', '&:hover': { color: 'primary.main' } }}>
                    {item.label}
                  </Typography>
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
                <Typography sx={{ color: 'grey.500' }}>{t('welcome', { name: user?.name ?? '' })}</Typography>
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
        <Box sx={{ display: { xs: 'block', sm: 'none' }, bgcolor: 'grey.50', borderTop: '1px solid', borderColor: 'grey.200' }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', py: '8px' }}>
            {isAuthenticated && (
              <>
                {navigationItems.map((item) => (
                  <Link key={item.href} href={item.href}>
                    <Typography
                      sx={{
                        display: 'block',
                        px: '16px',
                        py: '8px',
                        color: 'grey.800',
                        textDecoration: 'none',
                        '&:hover': { bgcolor: 'white', color: 'primary.main' },
                      }}
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      {item.label}
                    </Typography>
                  </Link>
                ))}
                <Box sx={{ borderTop: '1px solid', borderColor: 'grey.200', mt: '8px', pt: '8px' }}>
                  <Typography
                    sx={{
                      display: 'block',
                      px: '16px',
                      py: '8px',
                      color: 'grey.800',
                      cursor: 'pointer',
                      '&:hover': { bgcolor: 'white', color: 'error.main' },
                    }}
                    onClick={() => {
                      handleLogout()
                      setIsMobileMenuOpen(false)
                    }}
                  >
                    {t('logout')}
                  </Typography>
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
