"use client"

import { useState } from 'react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Box, Container, Typography, OutlinedInput, Button, Alert } from '@mui/material'
import { Mail } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import AuthLayout from '@/components/Layout/AuthLayout'

const ConfirmEmailPage = () => {
  const t = useTranslations()
  const searchParams = useSearchParams()
  const email = searchParams.get('email') || ''
  const [otp, setOtp] = useState('')
  const [otpError, setOtpError] = useState('')

  const { register, registerLoading, registerError } = useAuth()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!otp) {
      setOtpError(t('validation.otpRequired'))
      return
    }
    if (otp.length !== 6) {
      setOtpError(t('validation.otpInvalid'))
      return
    }

    setOtpError('')
    register({ otp })
  }

  return (
    <AuthLayout>
      <Container maxWidth="xs">
        <Box sx={{ bgcolor: 'white', p: '24px', borderRadius: '12px', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
          <Box sx={{ display: 'flex', justifyContent: 'center', mb: '16px' }}>
            <Box
              sx={{
                bgcolor: 'primary.light',
                borderRadius: '50%',
                width: 64,
                height: 64,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'primary.main',
              }}
            >
              <Mail size={32} />
            </Box>
          </Box>

          <Typography component="h1" sx={{ textAlign: 'center', mb: '8px' }}>
            {t('auth.confirmEmailTitle')}
          </Typography>

          <Box sx={{ textAlign: 'center', color: 'grey.500', mb: '4px', display: 'block' }}>
            {t('auth.confirmEmailDescription')}
          </Box>

          {email && (
            <Box sx={{ textAlign: 'center', fontWeight: 600, mb: '24px' }}>
              {email}
            </Box>
          )}

          {registerError && (
            <Alert severity="error" sx={{ mb: '16px' }}>
              {registerError instanceof Error ? registerError.message : t('auth.registrationFailed')}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: '24px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('auth.otpLabel')}
              </Box>
              <OutlinedInput
                size="small"
                type="text"
                value={otp}
                onChange={(e) => {
                  const val = e.target.value.replace(/\D/g, '').slice(0, 6)
                  setOtp(val)
                  if (otpError) setOtpError('')
                }}
                placeholder={t('auth.otpPlaceholder')}
                inputProps={{ maxLength: 6, inputMode: 'numeric' }}
                sx={{ width: '100%', letterSpacing: '0.3em', fontSize: '20px' }}
                className={otpError ? 'error' : ''}
              />
              {otpError && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {otpError}
                </Box>
              )}
            </Box>

            <Button
              type="submit"
              variant="contained"
              disableElevation
              sx={{ width: '100%', mb: '16px' }}
              disabled={registerLoading}
            >
              {registerLoading ? t('auth.verifying') : t('auth.verifyAndCreate')}
            </Button>
          </Box>

          <Box sx={{ display: 'flex', justifyContent: 'center', gap: '4px' }}>
            <Box sx={{ color: 'grey.500', fontSize: '14px' }}>{t('auth.confirmEmailNoEmail')}</Box>
            <Link href="/register" style={{ textDecoration: 'none' }}>
              <Box sx={{ color: 'primary.main', fontSize: '14px', '&:hover': { textDecoration: 'underline' } }}>
                {t('auth.confirmEmailResend')}
              </Box>
            </Link>
          </Box>

          <Box sx={{ display: 'flex', justifyContent: 'center', mt: '16px' }}>
            <Link href="/login">
              <Box sx={{ color: 'grey.500', fontSize: '14px', cursor: 'pointer', '&:hover': { color: 'primary.main' } }}>
                {t('auth.backToLogin')}
              </Box>
            </Link>
          </Box>
        </Box>
      </Container>
    </AuthLayout>
  )
}

export default ConfirmEmailPage
