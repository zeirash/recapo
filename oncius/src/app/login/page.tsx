"use client"

import { useState } from 'react'
import { useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { Box, Container, Typography, OutlinedInput, Button, Alert } from '@mui/material'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import AuthLayout from '@/components/Layout/AuthLayout'

const LoginPage = () => {
  const t = useTranslations()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<{ [key: string]: string }>({})

  const { login, loginLoading, loginError } = useAuth()
  const searchParams = useSearchParams()
  const message = searchParams.get('message')

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!email) {
      newErrors.email = t('validation.emailRequired')
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = t('validation.emailInvalid')
    }

    if (!password) {
      newErrors.password = t('validation.passwordRequired')
    } else if (password.length < 6) {
      newErrors.password = t('validation.passwordMinLength')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    login({ email, password })
  }

  return (
    <AuthLayout>
      <Container maxWidth="xs">
        <Box sx={{ bgcolor: 'white', p: '24px', borderRadius: '12px', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
          <Typography component="h1" sx={{ textAlign: 'center', mb: '24px' }}>
            {t('auth.login')}
          </Typography>

          {message && (
            <Alert severity="success" sx={{ mb: '16px' }}>
              {message}
            </Alert>
          )}

          {loginError && (
            <Alert severity="error" sx={{ mb: '16px' }}>
              {loginError instanceof Error ? loginError.message : t('auth.loginFailed')}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('common.email')}
              </Box>
              <OutlinedInput
                size="small"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={t('auth.enterEmail')}
                sx={{ width: '100%' }}
                className={errors.email ? 'error' : ''}
              />
              {errors.email && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {errors.email}
                </Box>
              )}
            </Box>

            <Box sx={{ mb: '24px' }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', mb: '4px' }}>
                <Box component="label" sx={{ fontWeight: 600 }}>
                  {t('common.password')}
                </Box>
                <Link href="/forgot-password">
                  <Box sx={{ color: 'primary.main', fontSize: '12px', cursor: 'pointer' }}>
                    {t('auth.forgotPassword')}
                  </Box>
                </Link>
              </Box>
              <OutlinedInput
                size="small"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t('auth.enterPassword')}
                sx={{ width: '100%' }}
                className={errors.password ? 'error' : ''}
              />
              {errors.password && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {errors.password}
                </Box>
              )}
            </Box>

            <Button
              type="submit"
              variant="contained"
              disableElevation
              sx={{ width: '100%', mb: '16px' }}
              disabled={loginLoading}
            >
              {loginLoading ? t('auth.loggingIn') : t('auth.login')}
            </Button>

            <Box sx={{ display: 'flex', justifyContent: 'center', gap: '4px' }}>
              <Box sx={{ color: 'grey.500' }}>{t('auth.noAccount')}</Box>
              <Link href="/register">
                <Box
                  sx={{ color: 'primary.main', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } }}
                >
                  {t('auth.registerHere')}
                </Box>
              </Link>
            </Box>
          </Box>
        </Box>
      </Container>
    </AuthLayout>
  )
}

export default LoginPage
