"use client"

import { useState } from 'react'
import Link from 'next/link'
import { Box, Container, Typography, OutlinedInput, Button, Alert } from '@mui/material'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import AuthLayout from '@/components/Layout/AuthLayout'

const RegisterPage = () => {
  const t = useTranslations()
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [errors, setErrors] = useState<{ [key: string]: string }>({})

  const { sendOtp, sendOtpLoading, sendOtpError } = useAuth()

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!formData.name.trim()) {
      newErrors.name = t('validation.nameRequired')
    }

    if (!formData.email) {
      newErrors.email = t('validation.emailRequired')
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = t('validation.emailInvalid')
    }

    if (!formData.password) {
      newErrors.password = t('validation.passwordRequired')
    } else if (formData.password.length < 8 || !/[a-zA-Z]/.test(formData.password) || !/[0-9]/.test(formData.password)) {
      newErrors.password = t('validation.passwordMinLength')
    }

    if (!formData.confirmPassword) {
      newErrors.confirmPassword = t('validation.confirmPasswordRequired')
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = t('validation.passwordsMismatch')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))

    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    sendOtp({
      name: formData.name,
      email: formData.email,
      password: formData.password,
    })
  }

  return (
    <AuthLayout>
      <Container maxWidth="xs">
        <Box sx={{ bgcolor: 'white', p: '24px', borderRadius: '12px', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
          <Typography component="h1" sx={{ textAlign: 'center', mb: '24px' }}>
            {t('auth.register')}
          </Typography>

          {sendOtpError && (
            <Alert severity="error" sx={{ mb: '16px' }}>
              {sendOtpError instanceof Error ? sendOtpError.message : t('auth.registrationFailed')}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('auth.fullName')}
              </Box>
              <OutlinedInput
                size="small"
                name="name"
                type="text"
                value={formData.name}
                onChange={handleChange}
                placeholder={t('auth.enterFullName')}
                sx={{ width: '100%' }}
                className={errors.name ? 'error' : ''}
              />
              {errors.name && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {errors.name}
                </Box>
              )}
            </Box>

            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('common.email')}
              </Box>
              <OutlinedInput
                size="small"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleChange}
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

            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('common.password')}
              </Box>
              <OutlinedInput
                size="small"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleChange}
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

            <Box sx={{ mb: '24px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600 }}>
                {t('auth.confirmPassword')}
              </Box>
              <OutlinedInput
                size="small"
                name="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={handleChange}
                placeholder={t('auth.confirmYourPassword')}
                sx={{ width: '100%' }}
                className={errors.confirmPassword ? 'error' : ''}
              />
              {errors.confirmPassword && (
                <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>
                  {errors.confirmPassword}
                </Box>
              )}
            </Box>

            <Button
              type="submit"
              variant="contained"
              disableElevation
              fullWidth
              disabled={sendOtpLoading}
              sx={{ mb: '16px' }}
            >
              {sendOtpLoading ? t('auth.sendingCode') : t('auth.register')}
            </Button>

            <Box sx={{ display: 'flex', justifyContent: 'center', gap: '4px' }}>
              <Box sx={{ color: 'grey.500' }}>{t('auth.hasAccount')}</Box>
              <Link href="/login">
                <Box
                  sx={{ color: 'primary.main', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } }}
                >
                  {t('auth.loginHere')}
                </Box>
              </Link>
            </Box>
          </Box>
        </Box>
      </Container>
    </AuthLayout>
  )
}

export default RegisterPage
