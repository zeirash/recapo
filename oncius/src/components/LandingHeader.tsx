'use client'

import Link from 'next/link'
import { Box, Button } from '@mui/material'
import { useTranslations, useLocale } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import { useChangeLocale } from '@/hooks/useLocale'

export default function LandingHeader() {
  const t = useTranslations()
  const locale = useLocale()
  const changeLocale = useChangeLocale()
  const { isAuthenticated } = useAuth()

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
        borderColor: '#e5e7eb',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          maxWidth: 1200,
          mx: 'auto',
          px: { xs: '16px', sm: '24px' },
          py: '16px',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Link href="/" style={{ textDecoration: 'none' }}>
          <Box component="span" sx={{ fontSize: { xs: '18px', sm: '20px' }, fontWeight: 700, color: '#3b82f6' }}>
            Recapo
          </Box>
        </Link>
        <Box sx={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
          <Box sx={{ display: 'flex', gap: '8px', mr: '8px' }}>
            <Box
              component="button"
              onClick={() => changeLocale('en')}
              sx={{
                bgcolor: 'transparent',
                border: 'none',
                cursor: 'pointer',
                fontSize: '14px',
                fontWeight: locale === 'en' ? 600 : 400,
                color: locale === 'en' ? '#3b82f6' : '#6b7280',
                '&:hover': { color: '#3b82f6' },
              }}
            >
              EN
            </Box>
            <Box sx={{ color: '#e5e7eb' }}>|</Box>
            <Box
              component="button"
              onClick={() => changeLocale('id')}
              sx={{
                bgcolor: 'transparent',
                border: 'none',
                cursor: 'pointer',
                fontSize: '14px',
                fontWeight: locale === 'id' ? 600 : 400,
                color: locale === 'id' ? '#3b82f6' : '#6b7280',
                '&:hover': { color: '#3b82f6' },
              }}
            >
              ID
            </Box>
          </Box>
          {isAuthenticated ? (
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
          )}
        </Box>
      </Box>
    </Box>
  )
}
