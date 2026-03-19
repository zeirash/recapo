"use client"

import { useRef, useState } from 'react'
import { useMutation } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, OutlinedInput, Typography } from '@mui/material'
import { Bug, CheckCircle, ImagePlus, Lightbulb, X } from 'lucide-react'
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
  const [image, setImage] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const resetForm = () => {
    setType('bug')
    setTitle('')
    setDescription('')
    setImage(null)
    setImagePreview(null)
  }

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] ?? null
    setImage(file)
    if (file) {
      setImagePreview(URL.createObjectURL(file))
    } else {
      setImagePreview(null)
    }
    // reset so same file can be re-selected after removal
    e.target.value = ''
  }

  const removeImage = () => {
    setImage(null)
    if (imagePreview) URL.revokeObjectURL(imagePreview)
    setImagePreview(null)
  }

  const mutation = useMutation(
    async () => {
      const res = await api.createFeedback({
        type,
        title,
        description: description || undefined,
        image: image ?? undefined,
      })
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

            {/* Image upload */}
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: 'block' }}>
                {t('imageLabel')}
              </Typography>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                style={{ display: 'none' }}
                onChange={handleImageChange}
              />
              {imagePreview ? (
                <Box sx={{ position: 'relative', display: 'inline-block' }}>
                  <Box
                    component="img"
                    src={imagePreview}
                    alt="preview"
                    sx={{ maxWidth: '100%', maxHeight: 160, borderRadius: '8px', border: '1px solid', borderColor: 'grey.200', display: 'block' }}
                  />
                  <Box
                    onClick={removeImage}
                    sx={{
                      position: 'absolute',
                      top: 4,
                      right: 4,
                      bgcolor: 'rgba(0,0,0,0.55)',
                      borderRadius: '50%',
                      width: 22,
                      height: 22,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer',
                      color: 'white',
                      '&:hover': { bgcolor: 'rgba(0,0,0,0.75)' },
                    }}
                  >
                    <X size={13} />
                  </Box>
                </Box>
              ) : (
                <Box
                  onClick={() => !mutation.isLoading && fileInputRef.current?.click()}
                  sx={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 1,
                    px: 2,
                    py: 1,
                    borderRadius: '8px',
                    border: '1px dashed',
                    borderColor: 'grey.300',
                    color: 'text.secondary',
                    cursor: mutation.isLoading ? 'not-allowed' : 'pointer',
                    fontSize: '13px',
                    '&:hover': mutation.isLoading ? {} : { borderColor: 'primary.main', color: 'primary.main' },
                  }}
                >
                  <ImagePlus size={16} />
                  {t('imageChoose')}
                </Box>
              )}
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
            <Button variant="contained" disableElevation type="submit" disabled={mutation.isLoading} startIcon={mutation.isLoading ? <CircularProgress size={16} color="inherit" /> : null}>
              {t('submit')}
            </Button>
          </DialogActions>
        </form>
      )}
    </Dialog>
  )
}

export default FeedbackDialog
