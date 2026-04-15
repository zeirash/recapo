"use client"

import { useState, useEffect } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Box, Container, Typography, OutlinedInput, Button, Alert, CircularProgress } from '@mui/material'
import { useTranslations } from 'next-intl'
import AuthLayout from '@/components/Layout/AuthLayout'
import { api, setAuthToken, setRefreshToken } from '@/utils/api'

const AcceptInvitePage = () => {
  const t = useTranslations()
  const router = useRouter()
  const searchParams = useSearchParams()
  const token = searchParams.get('token') ?? ''

  const [inviteData, setInviteData] = useState<{ email: string; shop_name: string } | null>(null)
  const [validateError, setValidateError] = useState('')
  const [validating, setValidating] = useState(true)

  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [errors, setErrors] = useState<{ [key: string]: string }>({})
  const [submitError, setSubmitError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!token) {
      setValidateError(t('invite.invalidLink'))
      setValidating(false)
      return
    }
    api.validateInvite(token)
      .then((res) => {
        if (res.success && res.data) {
          setInviteData(res.data)
        } else {
          setValidateError(t('invite.invalidLink'))
        }
      })
      .catch(() => setValidateError(t('invite.invalidLink')))
      .finally(() => setValidating(false))
  }, [token]) // eslint-disable-line react-hooks/exhaustive-deps

  const validate = () => {
    const newErrors: { [key: string]: string } = {}
    if (!name.trim()) newErrors.name = t('validation.nameRequired')
    if (!password) {
      newErrors.password = t('validation.passwordRequired')
    } else if (password.length < 8 || !/[a-zA-Z]/.test(password) || !/[0-9]/.test(password)) {
      newErrors.password = t('validation.passwordMinLength')
    }
    if (!confirmPassword) {
      newErrors.confirmPassword = t('validation.confirmPasswordRequired')
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = t('validation.passwordsMismatch')
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    setSubmitError('')
    setLoading(true)
    try {
      const res = await api.acceptInvite(token, name, password)
      if (res.success && res.data?.access_token) {
        setAuthToken(res.data.access_token)
        if (res.data.refresh_token) setRefreshToken(res.data.refresh_token)
        router.push('/dashboard')
      } else {
        setSubmitError(res.message || t('invite.acceptFailed'))
      }
    } catch (err: any) {
      setSubmitError(err.message || t('invite.acceptFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthLayout>
      <Container maxWidth="xs">
        <Box sx={{ bgcolor: 'white', p: '24px', borderRadius: '12px', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
          {validating ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: '24px' }}>
              <CircularProgress size={32} />
            </Box>
          ) : validateError ? (
            <>
              <Typography component="h1" sx={{ textAlign: 'center', mb: '16px', color: 'text.primary' }}>
                {t('invite.title')}
              </Typography>
              <Alert severity="error">{validateError}</Alert>
            </>
          ) : (
            <>
              <Typography component="h1" sx={{ textAlign: 'center', mb: '4px', color: 'text.primary' }}>
                {t('invite.title')}
              </Typography>
              <Box sx={{ textAlign: 'center', mb: '24px', color: 'text.secondary', fontSize: '14px' }}>
                {t('invite.subtitle', { shopName: inviteData?.shop_name ?? '' })}
              </Box>

              <Box sx={{ mb: '16px', p: '12px', bgcolor: 'grey.50', borderRadius: '8px', fontSize: '14px', color: 'text.secondary' }}>
                {t('common.email')}: <Box component="span" sx={{ fontWeight: 600, color: 'text.primary' }}>{inviteData?.email}</Box>
              </Box>

              {submitError && (
                <Alert severity="error" sx={{ mb: '16px' }}>{submitError}</Alert>
              )}

              <Box component="form" onSubmit={handleSubmit}>
                <Box sx={{ mb: '16px' }}>
                  <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600, color: 'text.primary' }}>
                    {t('auth.fullName')}
                  </Box>
                  <OutlinedInput
                    size="small"
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder={t('auth.enterFullName')}
                    sx={{ width: '100%' }}
                  />
                  {errors.name && <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>{errors.name}</Box>}
                </Box>

                <Box sx={{ mb: '16px' }}>
                  <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600, color: 'text.primary' }}>
                    {t('common.password')}
                  </Box>
                  <OutlinedInput
                    size="small"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder={t('auth.enterPassword')}
                    sx={{ width: '100%' }}
                  />
                  {errors.password && <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>{errors.password}</Box>}
                </Box>

                <Box sx={{ mb: '24px' }}>
                  <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600, color: 'text.primary' }}>
                    {t('auth.confirmPassword')}
                  </Box>
                  <OutlinedInput
                    size="small"
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder={t('auth.confirmYourPassword')}
                    sx={{ width: '100%' }}
                  />
                  {errors.confirmPassword && <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>{errors.confirmPassword}</Box>}
                </Box>

                <Button
                  type="submit"
                  variant="contained"
                  disableElevation
                  fullWidth
                  disabled={loading}
                >
                  {loading ? t('invite.accepting') : t('invite.accept')}
                </Button>
              </Box>
            </>
          )}
        </Box>
      </Container>
    </AuthLayout>
  )
}

export default AcceptInvitePage
