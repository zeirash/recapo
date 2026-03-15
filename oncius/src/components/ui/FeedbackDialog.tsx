"use client"

import { useState } from 'react'
import { useMutation } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, OutlinedInput, Typography } from '@mui/material'
import { Bug, CheckCircle, Lightbulb } from 'lucide-react'
import { api } from '@/utils/api'

interface FeedbackDialogProps {
  open: boolean
  onClose: () => void
}

type FeedbackType = 'bug' | 'enhancement'

const FeedbackDialog = ({ open, onClose }: FeedbackDialogProps) => {
  const t = useTranslations('feedback')
  const tCommon = useTranslations('common')

  const [type, setType] = useState<FeedbackType>('bug')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')

  const resetForm = () => {
    setType('bug')
    setTitle('')
    setDescription('')
  }

  const mutation = useMutation(
    async () => {
      const res = await api.createFeedback({ type, title, description: description || undefined })
      if (!res.success) throw new Error(res.message || t('errorMessage'))
      return res
    }
  )

  const handleClose = () => {
    if (mutation.isLoading) return
    resetForm()
    mutation.reset()
    onClose()
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate()
  }

  const typeOptions: { value: FeedbackType; label: string; icon: typeof Bug }[] = [
    { value: 'bug', label: t('typeBug'), icon: Bug },
    { value: 'enhancement', label: t('typeEnhancement'), icon: Lightbulb },
  ]

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      {mutation.isSuccess ? (
        <>
          <DialogContent sx={{ pt: '32px !important', pb: 2, textAlign: 'center' }}>
            <Box sx={{ display: 'flex', justifyContent: 'center', mb: 2, color: 'success.main' }}>
              <CheckCircle size={48} />
            </Box>
            <Typography variant="h6" fontWeight={600} mb={1}>
              {t('successTitle')}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t('successMessage')}
            </Typography>
          </DialogContent>
          <DialogActions sx={{ justifyContent: 'center', pb: 3 }}>
            <Button variant="contained" disableElevation onClick={handleClose}>
              {tCommon('close')}
            </Button>
          </DialogActions>
        </>
      ) : (
        <form onSubmit={handleSubmit}>
          <DialogTitle>{t('title')}</DialogTitle>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '8px !important' }}>
            {/* Type toggle */}
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                {t('typeLabel')}
              </Typography>
              <Box sx={{ display: 'flex', gap: 1 }}>
                {typeOptions.map(({ value, label, icon: Icon }) => (
                  <Box
                    key={value}
                    onClick={() => setType(value)}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1,
                      px: 2,
                      py: 1,
                      borderRadius: '8px',
                      border: '1px solid',
                      borderColor: type === value ? 'primary.main' : 'grey.300',
                      bgcolor: type === value ? 'primary.50' : 'transparent',
                      color: type === value ? 'primary.main' : 'text.secondary',
                      cursor: 'pointer',
                      fontSize: '14px',
                      fontWeight: type === value ? 600 : 400,
                      transition: 'all 0.15s',
                      '&:hover': {
                        borderColor: 'primary.main',
                        bgcolor: 'primary.50',
                      },
                    }}
                  >
                    <Icon size={16} />
                    {label}
                  </Box>
                ))}
              </Box>
            </Box>

            {/* Title */}
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: 'block' }}>
                {t('titleLabel')}
              </Typography>
              <OutlinedInput
                size="small"
                fullWidth
                required
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={t('titlePlaceholder')}
                disabled={mutation.isLoading}
              />
            </Box>

            {/* Description */}
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: 'block' }}>
                {t('descriptionLabel')}
              </Typography>
              <OutlinedInput
                size="small"
                fullWidth
                multiline
                rows={4}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={t('descriptionPlaceholder')}
                disabled={mutation.isLoading}
              />
            </Box>

            {mutation.isError && (
              <Typography color="error" variant="body2">
                {mutation.error instanceof Error ? mutation.error.message : t('errorMessage')}
              </Typography>
            )}
          </DialogContent>
          <DialogActions>
            <Button variant="outlined" onClick={handleClose} disabled={mutation.isLoading}>
              {tCommon('cancel')}
            </Button>
            <Button variant="contained" disableElevation type="submit" disabled={mutation.isLoading}>
              {mutation.isLoading ? t('submitting') : t('submit')}
            </Button>
          </DialogActions>
        </form>
      )}
    </Dialog>
  )
}

export default FeedbackDialog
