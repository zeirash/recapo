"use client"

import { useState } from 'react'
import Link from 'next/link'
import { Box, Container, Typography, OutlinedInput, Button, Alert } from '@mui/material'
import { useTranslations } from 'next-intl'
import Layout from '@/components/Layout'
import { useAuth } from '@/hooks/useAuth'

const ForgotPasswordPage = () => {
  const t = useTranslations()
  const [email, setEmail] = useState('')
  const [errors, setErrors] = useState<{ [key: string]: string }>({})

  const { forgotPassword, forgotPasswordLoading, forgotPasswordError } = useAuth()

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!email) {
      newErrors.email = t('validation.emailRequired')
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = t('validation.emailInvalid')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    forgotPassword(email)
  }

  return (
    <Layout>
      <Container maxWidth="xs">
        <Box sx={{ bgcolor: 'white', p: '24px', borderRadius: '12px', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
          <Typography component="h1" sx={{ textAlign: 'center', mb: '8px' }}>
            {t('auth.forgotPasswordTitle')}
          </Typography>

          <Box sx={{ textAlign: 'center', color: 'grey.500', mb: '24px', display: 'block' }}>
            {t('auth.forgotPasswordDescription')}
          </Box>

          {forgotPasswordError && (
            <Alert severity="error" sx={{ mb: '16px' }}>
              {forgotPasswordError instanceof Error ? forgotPasswordError.message : t('auth.forgotPasswordFailed')}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: '24px' }}>
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

            <Button
              type="submit"
              variant="contained"
              disableElevation
              sx={{ width: '100%', mb: '16px' }}
              disabled={forgotPasswordLoading}
            >
              {forgotPasswordLoading ? t('auth.sendingCode') : t('auth.forgotPasswordSubmit')}
            </Button>

            <Box sx={{ display: 'flex', justifyContent: 'center' }}>
              <Link href="/login">
                <Box sx={{ color: 'primary.main', cursor: 'pointer' }}>
                  {t('auth.backToLogin')}
                </Box>
              </Link>
            </Box>
          </Box>
        </Box>
      </Container>
    </Layout>
  )
}

export default ForgotPasswordPage
