'use client'

import Link from 'next/link'
import { Box, Flex, Text, Button } from 'theme-ui'
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
      as="header"
      sx={{
        position: 'sticky',
        top: 0,
        zIndex: 10,
        bg: 'rgba(255,255,255,0.9)',
        backdropFilter: 'saturate(180%) blur(10px)',
        borderBottom: '1px solid',
        borderColor: 'border',
      }}
    >
      <Flex
        sx={{
          maxWidth: 1200,
          mx: 'auto',
          px: [3, 4],
          py: 3,
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Link href="/" style={{ textDecoration: 'none' }}>
          <Text as="span" sx={{ fontSize: [3, 4], fontWeight: 700, color: 'primary' }}>
            Recapo
          </Text>
        </Link>
        <Flex sx={{ gap: 3, alignItems: 'center' }}>
          <Flex sx={{ gap: 2, mr: 2 }}>
            <Box
              as="button"
              onClick={() => changeLocale('en')}
              sx={{
                bg: 'transparent',
                border: 'none',
                cursor: 'pointer',
                fontSize: 1,
                fontWeight: locale === 'en' ? 600 : 400,
                color: locale === 'en' ? 'primary' : 'text.secondary',
                '&:hover': { color: 'primary' },
              }}
            >
              EN
            </Box>
            <Text sx={{ color: 'border' }}>|</Text>
            <Box
              as="button"
              onClick={() => changeLocale('id')}
              sx={{
                bg: 'transparent',
                border: 'none',
                cursor: 'pointer',
                fontSize: 1,
                fontWeight: locale === 'id' ? 600 : 400,
                color: locale === 'id' ? 'primary' : 'text.secondary',
                '&:hover': { color: 'primary' },
              }}
            >
              ID
            </Box>
          </Flex>
          {isAuthenticated ? (
            <Link href="/dashboard">
              <Button variant="primary" sx={{ py: 2 }}>
                {t('landing.goToDashboard')}
              </Button>
            </Link>
          ) : (
            <>
              <Link href="/login">
                <Button variant="secondary" sx={{ py: 2 }}>
                  {t('nav.login')}
                </Button>
              </Link>
              <Link href="/register">
                <Button variant="primary" sx={{ py: 2 }}>
                  {t('landing.getStarted')}
                </Button>
              </Link>
            </>
          )}
        </Flex>
      </Flex>
    </Box>
  )
}
