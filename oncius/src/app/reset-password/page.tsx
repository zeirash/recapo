"use client"

import { useState } from 'react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Box, Container, Typography, OutlinedInput, Button, Alert } from '@mui/material'
import { KeyRound } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Layout from '@/components/Layout'
import { useAuth } from '@/hooks/useAuth'

const ResetPasswordPage = () => {
  const t = useTranslations()
  const searchParams = useSearchParams()
  const email = searchParams.get('email') || ''

  const [otp, setOtp] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [formErrors, setFormErrors] = useState<{ [key: string]: string }>({})

  const { resetPassword, resetPasswordLoading, resetPasswordError } = useAuth()

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!otp) {
      newErrors.otp = t('validation.otpRequired')
    } else if (otp.length !== 6) {
      newErrors.otp = t('validation.otpInvalid')
    }

    if (!password) {
      newErrors.password = t('validation.passwordRequired')
    } else if (password.length < 6) {
      newErrors.password = t('validation.passwordMinLength')
    }

    if (!confirmPassword) {
      newErrors.confirmPassword = t('validation.confirmPasswordRequired')
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = t('validation.passwordsMismatch')
    }

    setFormErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    resetPassword({ email, otp, password })
  }

  return (
    <Layout>
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
              <KeyRound size={32} />
            </Box>
          </Box>

          <Typography component="h1" sx={{ textAlign: 'center', mb: '8px' }}>
            {t('auth.resetPasswordTitle')}
          </Typography>

          <Box sx={{ textAlign: 'center', color: 'grey.500', mb: '4px', display: 'block' }}>
            {t('auth.resetPasswordDescription')}
          </Box>

          {email && (
            <Box sx={{ textAlign: 'center', fontWeight: 600, mb: '24px' }}>
              {email}
            </Box>
          )}

          {resetPasswordError && (
            <Alert severity="error" sx={{ mb: '16px' }}>
              {resetPasswordError instanceof Error ? resetPasswordError.message : t('auth.resetPasswordFailed')}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: '16px' }}>
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
                  if (formErrors.otp) setFormErrors(prev => ({ ...prev, otp: '' }))
                }}
                placeholder={t('auth.otpPlaceholder')}
                inputProps={{ maxLength: 6, inputMode: 'numeric' }}
                sx={{ width: '100%', letterSpacing: '0.3em', fontSize: '20px' }}
              />
              {formErrors.otp && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {formErrors.otp}
                </Box>
              )}
            </Box>

            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('auth.newPassword')}
              </Box>
              <OutlinedInput
                size="small"
                type="password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value)
                  if (formErrors.password) setFormErrors(prev => ({ ...prev, password: '' }))
                }}
                placeholder={t('auth.enterPassword')}
                sx={{ width: '100%' }}
              />
              {formErrors.password && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {formErrors.password}
                </Box>
              )}
            </Box>

            <Box sx={{ mb: '24px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('auth.confirmPassword')}
              </Box>
              <OutlinedInput
                size="small"
                type="password"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value)
                  if (formErrors.confirmPassword) setFormErrors(prev => ({ ...prev, confirmPassword: '' }))
                }}
                placeholder={t('auth.confirmYourPassword')}
                sx={{ width: '100%' }}
              />
              {formErrors.confirmPassword && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {formErrors.confirmPassword}
                </Box>
              )}
            </Box>

            <Button
              type="submit"
              variant="contained"
              disableElevation
              sx={{ width: '100%', mb: '16px' }}
              disabled={resetPasswordLoading}
            >
              {resetPasswordLoading ? t('auth.verifying') : t('auth.resetPasswordSubmit')}
            </Button>
          </Box>

          <Box sx={{ display: 'flex', justifyContent: 'center', mt: '8px' }}>
            <Link href="/login">
              <Box sx={{ color: 'grey.500', fontSize: '14px', cursor: 'pointer', '&:hover': { color: 'primary.main' } }}>
                {t('auth.backToLogin')}
              </Box>
            </Link>
          </Box>
        </Box>
      </Container>
    </Layout>
  )
}

export default ResetPasswordPage
