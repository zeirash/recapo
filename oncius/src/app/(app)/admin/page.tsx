"use client"

import { useState } from 'react'
import { Box, Typography, Button, Paper, OutlinedInput, Alert } from '@mui/material'
import { useTranslations } from 'next-intl'
import { UserCog, Send } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { api } from '@/utils/api'
import { USER_ROLES } from '@/constants/roles'

const AdminPage = () => {
  const t = useTranslations()
  const { user } = useAuth()

  const [email, setEmail] = useState('')
  const [emailError, setEmailError] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState('')

  const isOwner = user?.role === USER_ROLES.OWNER

  const validate = () => {
    if (!email) {
      setEmailError(t('validation.emailRequired'))
      return false
    }
    if (!/\S+@\S+\.\S+/.test(email)) {
      setEmailError(t('validation.emailInvalid'))
      return false
    }
    setEmailError('')
    return true
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    setError('')
    setSuccess('')
    setLoading(true)
    try {
      const res = await api.inviteAdmin(email)
      if (res.success) {
        setSuccess(t('invite.sentSuccess', { email }))
        setEmail('')
      } else {
        setError(res.message || t('invite.sendFailed'))
      }
    } catch (err: any) {
      setError(err.message || t('invite.sendFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box sx={{ maxWidth: 900, mx: 'auto', px: { xs: '16px', sm: '24px' }, py: '32px' }}>
      <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '8px' }}>
        {t('invite.inviteAdmin')}
      </Typography>
      <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '32px' }}>
        {t('invite.inviteDescription')}
      </Box>

      {!isOwner ? (
        <Paper sx={{ p: '24px', borderRadius: '12px', border: '1px solid', borderColor: 'grey.200' }}>
          <Alert severity="error">{t('invite.ownerOnly')}</Alert>
        </Paper>
      ) : (
        <Paper sx={{ p: '24px', borderRadius: '12px', border: '1px solid', borderColor: 'grey.200' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px', mb: '20px' }}>
            <UserCog size={18} />
            <Typography sx={{ fontWeight: 600, fontSize: '16px' }}>
              {t('invite.sendInvite')}
            </Typography>
          </Box>

          {success && <Alert severity="success" sx={{ mb: '16px' }}>{success}</Alert>}
          {error && <Alert severity="error" sx={{ mb: '16px' }}>{error}</Alert>}

          <Box component="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontWeight: 600, color: 'text.primary', fontSize: '14px' }}>
                {t('common.email')}
              </Box>
              <Box sx={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
                <OutlinedInput
                  size="small"
                  type="email"
                  value={email}
                  onChange={(e) => { setEmail(e.target.value); setEmailError('') }}
                  placeholder={t('auth.enterEmail')}
                  sx={{ flex: 1, minWidth: '200px' }}
                />
                <Button
                  type="submit"
                  variant="contained"
                  disableElevation
                  disabled={loading}
                  startIcon={<Send size={15} />}
                  sx={{ whiteSpace: 'nowrap' }}
                >
                  {loading ? t('invite.sending') : t('invite.sendInvite')}
                </Button>
              </Box>
              {emailError && <Box sx={{ color: 'error.main', fontSize: '12px', mt: '4px' }}>{emailError}</Box>}
            </Box>
          </Box>
        </Paper>
      )}
    </Box>
  )
}

export default AdminPage
