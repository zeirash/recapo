'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { Box, Button } from '@mui/material'
import { useTranslations, useLocale } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import { useChangeLocale } from '@/hooks/useLocale'
import RecapoLogoText from '@/components/ui/RecapoLogoText'
import { Menu, X } from 'lucide-react'

export default function Header({ actions }: { actions?: React.ReactNode }) {
  const t = useTranslations()
  const locale = useLocale()
  const changeLocale = useChangeLocale()
  const { isAuthenticated } = useAuth()
  const [menuOpen, setMenuOpen] = useState(false)

  const localeToggle = (
    <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
      <Box
        component="button"
        onClick={() => changeLocale('en')}
        sx={{
          bgcolor: 'transparent',
          border: 'none',
          cursor: 'pointer',
          fontSize: '14px',
          fontWeight: locale === 'en' ? 600 : 400,
          color: locale === 'en' ? 'primary.main' : 'grey.500',
          '&:hover': { color: 'primary.main' },
        }}
      >
        EN
      </Box>
      <Box sx={{ color: 'grey.200' }}>|</Box>
      <Box
        component="button"
        onClick={() => changeLocale('id')}
        sx={{
          bgcolor: 'transparent',
          border: 'none',
          cursor: 'pointer',
          fontSize: '14px',
          fontWeight: locale === 'id' ? 600 : 400,
          color: locale === 'id' ? 'primary.main' : 'grey.500',
          '&:hover': { color: 'primary.main' },
        }}
      >
        ID
      </Box>
    </Box>
  )

  const menuLinkSx = {
    display: 'block',
    py: '12px',
    px: '16px',
    fontSize: '15px',
    fontWeight: 500,
    color: 'grey.800',
    textDecoration: 'none',
    textAlign: 'center',
    borderBottom: '1px solid',
    borderColor: 'grey.100',
    '&:hover': { color: 'primary.main' },
  }

  const ctaButtons = actions ?? (isAuthenticated ? (
    <Link href="/dashboard" onClick={() => setMenuOpen(false)}>
      <Box sx={menuLinkSx}>{t('landing.goToDashboard')}</Box>
    </Link>
  ) : (
    <>
      <Link href="/login" onClick={() => setMenuOpen(false)}>
        <Box sx={menuLinkSx}>{t('nav.login')}</Box>
      </Link>
      <Link href="/register" onClick={() => setMenuOpen(false)}>
        <Box sx={{ ...menuLinkSx, borderBottom: 'none' }}>{t('landing.getStarted')}</Box>
      </Link>
    </>
  ))

  return (
    <Box
      component="header"
      sx={{
        position: 'sticky',
        top: 0,
        zIndex: 10,
        bgcolor: 'rgba(255,255,255,0.9)',
        backdropFilter: 'saturate(180%) blur(10px)',
        borderBottom: '1px solid',
        borderColor: 'grey.200',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          maxWidth: 1200,
          mx: 'auto',
          px: { xs: '8px', sm: '24px' },
          py: '16px',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Link href="/" style={{ textDecoration: 'none', flexShrink: 0 }}>
          <RecapoLogoText width={170} height={50} />
        </Link>

        {/* Desktop nav */}
        <Box sx={{ display: { xs: 'none', sm: 'flex' }, gap: '16px', alignItems: 'center' }}>
          <Box sx={{ mr: '8px' }}>{localeToggle}</Box>
          {actions ?? (isAuthenticated ? (
            <Link href="/dashboard">
              <Button variant="contained" disableElevation sx={{ py: '8px' }}>
                {t('landing.goToDashboard')}
              </Button>
            </Link>
          ) : (
            <>
              <Link href="/login">
                <Button variant="outlined" sx={{ py: '8px' }}>
                  {t('nav.login')}
                </Button>
              </Link>
              <Link href="/register">
                <Button variant="contained" disableElevation sx={{ py: '8px' }}>
                  {t('landing.getStarted')}
                </Button>
              </Link>
            </>
          ))}
        </Box>

        {/* Mobile: locale + hamburger */}
        <Box sx={{ display: { xs: 'flex', sm: 'none' }, alignItems: 'center', gap: '12px' }}>
          {localeToggle}
          <Box
            component="button"
            onClick={() => setMenuOpen((v) => !v)}
            aria-label="Toggle menu"
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              bgcolor: 'transparent',
              border: '1px solid',
              borderColor: 'grey.300',
              borderRadius: '8px',
              p: '6px',
              cursor: 'pointer',
              color: 'grey.700',
            }}
          >
            {menuOpen ? <X size={20} /> : <Menu size={20} />}
          </Box>
        </Box>
      </Box>

      {/* Mobile dropdown */}
      {menuOpen && (
        <Box
          sx={{
            display: { xs: 'block', sm: 'none' },
            borderTop: '1px solid',
            borderColor: 'grey.200',
            bgcolor: 'white',
          }}
        >
          {ctaButtons}
        </Box>
      )}
    </Box>
  )
}
