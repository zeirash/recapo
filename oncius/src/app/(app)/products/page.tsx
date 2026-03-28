"use client"

import { useEffect, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, CircularProgress, Container, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, IconButton, MenuItem, OutlinedInput, Paper, Select, Tooltip, Typography } from '@mui/material'
import SearchInput from '@/components/ui/SearchInput'
import AddButton from '@/components/ui/AddButton'
import { api, resolveImageURL } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { Share2, Check, Package, Pencil, Trash2, ImageIcon } from 'lucide-react'

type Product = {
  id: number
  name: string
  description: string
  price: number
  original_price?: number | null
  image_url?: string | null
  created_at?: string
  updated_at?: string | null
}

type FormState = {
  name: string
  description: string
  price: number
  originalPrice: number | ''
}

const emptyForm: FormState = { name: '', description: '', price: 0, originalPrice: '' }

export default function ProductsPage() {
  const t = useTranslations('common')
  const tp = useTranslations('products')
  const tErrors = useTranslations('errors')
  const queryClient = useQueryClient()
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [sortValue, setSortValue] = useState<string>('updated_at,desc')
  const [shareCopied, setShareCopied] = useState(false)
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [imagePreviewURL, setImagePreviewURL] = useState<string | null>(null)
  const [imageRemoved, setImageRemoved] = useState(false)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [productToDelete, setProductToDelete] = useState<Product | null>(null)
  const justUploadedImageURL = useRef<string | null>(null)
  const pendingDeleteImageURL = useRef<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Debounce search: only trigger API after user stops typing for 300ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  // Revoke object URLs when imagePreviewURL changes to avoid memory leaks
  useEffect(() => {
    return () => { if (imagePreviewURL?.startsWith('blob:')) URL.revokeObjectURL(imagePreviewURL) }
  }, [imagePreviewURL])

  const sortParam = sortValue || undefined

  const { data: productsRes, isLoading, isError, error } = useQuery(
    ['products', debouncedSearch, sortParam],
    async () => {
      const res = await api.getProducts(debouncedSearch || undefined, sortParam)
      if (!res.success) throw new Error(res.message || tp('fetchFailed'))
      return res.data as Product[]
    },
    { keepPreviousData: true }
  )

  const createMutation = useMutation(
    async (payload: FormState & { imageURL?: string }) => {
      const data: Parameters<typeof api.createProduct>[0] = {
        name: payload.name,
        description: payload.description || undefined,
        price: payload.price,
      }
      if (payload.originalPrice !== '') data.original_price = payload.originalPrice
      if (payload.imageURL) data.image_url = payload.imageURL
      const res = await api.createProduct(data)
      if (!res.success) throw new Error(res.message || tp('createFailed'))
      return res
    },
    {
      onSuccess: () => {
        justUploadedImageURL.current = null
        queryClient.invalidateQueries(['products'])
        closeForm()
      },
      onError: () => {
        if (justUploadedImageURL.current) {
          api.deleteProductImage(justUploadedImageURL.current)
          justUploadedImageURL.current = null
        }
      },
    }
  )

  const updateMutation = useMutation(
    async ({ id, payload, imageURL }: { id: number; payload: Partial<FormState>; imageURL?: string }) => {
      const data: Parameters<typeof api.updateProduct>[1] = {}
      if (payload.name !== undefined) data.name = payload.name
      if (payload.description !== undefined) data.description = payload.description
      if (payload.price !== undefined) data.price = payload.price
      if (payload.originalPrice !== undefined && payload.originalPrice !== '') data.original_price = payload.originalPrice
      if (imageURL !== undefined) data.image_url = imageURL
      const res = await api.updateProduct(id, data)
      if (!res.success) throw new Error(res.message || tp('updateFailed'))
      return res
    },
    {
      onSuccess: () => {
        justUploadedImageURL.current = null
        if (pendingDeleteImageURL.current) {
          api.deleteProductImage(pendingDeleteImageURL.current)
          pendingDeleteImageURL.current = null
        }
        queryClient.invalidateQueries(['products'])
        closeForm()
      },
    }
  )

  const deleteMutation = useMutation(
    async (id: number) => {
      const res = await api.deleteProduct(id)
      if (!res.success) throw new Error(res.message || tp('deleteFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['products'])
      },
    }
  )

  function openCreateForm() {
    setEditingProduct(null)
    setForm(emptyForm)
    setIsFormOpen(true)
  }

  function openEditForm(product: Product) {
    setEditingProduct(product)
    setForm({
      name: product.name,
      description: product.description,
      price: product.price,
      originalPrice: product.original_price ?? '',
    })
    setImagePreviewURL(resolveImageURL(product.image_url) ?? null)
    setIsFormOpen(true)
  }

  function closeForm() {
    if (justUploadedImageURL.current) {
      api.deleteProductImage(justUploadedImageURL.current)
      justUploadedImageURL.current = null
    }
    pendingDeleteImageURL.current = null
    setImageFile(null)
    setImagePreviewURL(null)
    setImageRemoved(false)
    setUploadError(null)
    setIsFormOpen(false)
    setForm(emptyForm)
    setEditingProduct(null)
  }

  function handleRemoveImage() {
    if (fileInputRef.current) fileInputRef.current.value = ''
    if (imageFile) {
      // User picked a new file but hasn't submitted — undo the selection
      if (imagePreviewURL?.startsWith('blob:')) URL.revokeObjectURL(imagePreviewURL)
      setImageFile(null)
      setImagePreviewURL(editingProduct?.image_url ?? null)
    } else if (imagePreviewURL) {
      // Remove the existing product image; schedule deletion after successful save
      setImagePreviewURL(null)
      setImageRemoved(true)
    }
  }

  async function submitForm(e: React.FormEvent) {
    e.preventDefault()
    setUploadError(null)
    setIsSubmitting(true)

    try {
      let finalImageURL: string | undefined = editingProduct?.image_url ?? undefined

      if (imageRemoved && !imageFile) {
        finalImageURL = ''
        pendingDeleteImageURL.current = editingProduct?.image_url ?? null
      } else if (imageFile) {
        const uploadRes = await api.uploadProductImage(imageFile)
        if (!uploadRes.success || !uploadRes.data?.image_url) {
          setUploadError(tp('imageUploadFailed'))
          return
        }
        finalImageURL = uploadRes.data.image_url
        justUploadedImageURL.current = finalImageURL
      }

      if (editingProduct) {
        updateMutation.mutate({ id: editingProduct.id, payload: form, imageURL: finalImageURL })
      } else {
        createMutation.mutate({ ...form, imageURL: finalImageURL })
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleShare() {
    try {
      const res = await api.getShopShareToken()
      if (!res.success || !res.data?.share_token) throw new Error(tp('fetchFailed'))
      const url = `${typeof window !== 'undefined' ? window.location.origin : ''}/share/${res.data.share_token}`
      try {
        await navigator.clipboard.writeText(url)
      } catch {
        const textarea = document.createElement('textarea')
        textarea.value = url
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
      }
      setShareCopied(true)
      setTimeout(() => setShareCopied(false), 2000)
    } catch {
      // Silent fail - user can try again
    }
  }

  const products = productsRes || []

  return (
    <Container disableGutters maxWidth={false} sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        {/* Top bar */}
        <Box sx={{ p: '24px', flexShrink: 0 }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: '8px', maxWidth: 960, mx: 'auto' }}>
            <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <SearchInput
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                placeholder={tp('searchPlaceholder')}
              />
              <AddButton onClick={openCreateForm} title={tp('addProduct')} />
              <Button
                variant="outlined"
                onClick={handleShare}
                title={shareCopied ? tp('linkCopied') : tp('shareButton')}
                sx={{
                  minWidth: 36,
                  width: 36,
                  height: 36,
                  minHeight: 36,
                  p: 0,
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  borderRadius: '8px',
                  fontSize: '14px',
                }}
              >
                {shareCopied ? <Check size={16} /> : <Share2 size={16} />}
              </Button>
            </Box>
            <Select
              size="small"
              displayEmpty
              value={sortValue}
              onChange={(e) => setSortValue(e.target.value)}
              sx={{ height: 36, fontSize: '14px', borderRadius: '8px', width: 'fit-content', minWidth: 160 }}
              MenuProps={{ anchorOrigin: { vertical: 'bottom', horizontal: 'left' }, transformOrigin: { vertical: 'top', horizontal: 'left' } }}
            >
              <MenuItem value="updated_at,desc">{tp('sortUpdatedAtDesc')}</MenuItem>
              <MenuItem value="updated_at,asc">{tp('sortUpdatedAtAsc')}</MenuItem>
              <MenuItem value="name,asc">{tp('sortNameAsc')}</MenuItem>
              <MenuItem value="name,desc">{tp('sortNameDesc')}</MenuItem>
              <MenuItem value="price,asc">{tp('sortPriceAsc')}</MenuItem>
              <MenuItem value="price,desc">{tp('sortPriceDesc')}</MenuItem>
            </Select>
          </Box>
        </Box>

        {/* Scrollable body */}
        <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', p: '24px' }}>
          {isLoading && <PageLoadingSkeleton />}
          {isError && (
            <Box sx={{ color: 'error.main' }}>{(error as Error)?.message || tErrors('loadingError', { resource: tp('title') })}</Box>
          )}

          {/* Empty state */}
          {!isLoading && !isError && products.length === 0 && (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: '8px', color: 'text.secondary', minHeight: 320 }}>
              <Package size={48} opacity={0.4} />
              <Typography>{tp('noProducts')}</Typography>
            </Box>
          )}

          {/* Product list */}
          {!isLoading && !isError && products.length > 0 && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: '8px', maxWidth: 960, mx: 'auto' }}>
              {products.map((p) => (
                <Paper
                  key={p.id}
                  elevation={0}
                  sx={{
                    border: '1px solid',
                    borderColor: 'grey.200',
                    borderRadius: '10px',
                    bgcolor: 'background.paper',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '16px',
                    p: '12px 16px',
                    '&:hover': { borderColor: 'grey.300', bgcolor: 'action.hover' },
                  }}
                >
                  {/* Thumbnail */}
                  <Box sx={{
                    width: 48,
                    height: 48,
                    borderRadius: '8px',
                    background: 'linear-gradient(135deg,rgb(92, 151, 245) 0%,rgb(26, 94, 239) 100%)',
                    color: 'white',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '18px',
                    flexShrink: 0,
                    overflow: 'hidden',
                  }}>
                    {p.image_url
                      ? <img src={resolveImageURL(p.image_url)} alt={p.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                      : <ImageIcon size={22} color="rgba(255,255,255,0.6)" strokeWidth={1.5} />
                    }
                  </Box>

                  {/* Info */}
                  <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Typography sx={{ fontWeight: 600, fontSize: '14px', lineHeight: 1.3, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {p.name}
                    </Typography>
                    <Box sx={{ display: 'flex', alignItems: 'baseline', gap: '6px', mt: '2px' }}>
                      <Box sx={{ fontSize: '13px', fontWeight: 600, color: 'primary.main' }}>
                        Rp. {p.price.toLocaleString()}
                      </Box>
                      {p.original_price != null && p.original_price !== p.price && (
                        <Box sx={{ fontSize: '12px', color: 'grey.400' }}>
                          Rp. {p.original_price.toLocaleString()}
                        </Box>
                      )}
                    </Box>
                    {p.description && (
                      <Typography sx={{ fontSize: '12px', color: 'text.secondary', mt: '2px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {p.description}
                      </Typography>
                    )}
                  </Box>

                  {/* Actions */}
                  <Box sx={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                    <Tooltip title={t('edit')}>
                      <IconButton size="small" onClick={() => openEditForm(p)}>
                        <Pencil size={16} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('delete')}>
                      <IconButton
                        size="small"
                        onClick={() => setProductToDelete(p)}
                        sx={{ color: 'error.main', '&:hover': { bgcolor: 'error.light' } }}
                      >
                        <Trash2 size={16} />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </Paper>
              ))}
            </Box>
          )}
        </Box>

        <Dialog open={!!productToDelete} onClose={() => setProductToDelete(null)}>
          <DialogTitle>{tp('deleteConfirm')}</DialogTitle>
          <DialogContent>
            <DialogContentText sx={{ color: 'warning.dark', fontSize: '14px' }}>
              {tp('deleteConfirmWarning')}
            </DialogContentText>
          </DialogContent>
          <DialogActions sx={{ px: '24px', pb: '16px', gap: '8px' }}>
            <Button variant="outlined" onClick={() => setProductToDelete(null)}>
              {t('cancel')}
            </Button>
            <Button
              variant="contained"
              color="error"
              disableElevation
              disabled={deleteMutation.isLoading}
              onClick={() => {
                if (productToDelete) deleteMutation.mutate(productToDelete.id)
                setProductToDelete(null)
              }}
            >
              {t('delete')}
            </Button>
          </DialogActions>
        </Dialog>

        <Dialog open={isFormOpen} onClose={closeForm} fullWidth maxWidth="sm">
          <Box component="form" onSubmit={submitForm}>
            <DialogTitle sx={{ pb: '8px' }}>{editingProduct ? tp('editProduct') : tp('newProduct')}</DialogTitle>
            <DialogContent sx={{ pt: '8px', pb: 0 }}>
              {/* Hidden native file input */}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                style={{ display: 'none' }}
                onChange={(e) => {
                  const file = e.target.files?.[0]
                  if (!file) return
                  setImageFile(file)
                  setImagePreviewURL(URL.createObjectURL(file))
                }}
              />
              {/* Image preview + choose button */}
              <Box sx={{ mb: '16px' }}>
                <Box component="label" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{tp('image')}</Box>
                {imagePreviewURL && (
                  <Box sx={{ mb: '8px', width: 120, height: 120, borderRadius: '8px', overflow: 'hidden', border: '1px solid', borderColor: 'grey.200' }}>
                    <img src={imagePreviewURL} alt="preview" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  </Box>
                )}
                <Box sx={{ gap: '8px', alignItems: 'center', display: 'flex' }}>
                  <Button type="button" variant="outlined" onClick={() => fileInputRef.current?.click()}>
                    {tp('chooseImage')}
                  </Button>
                  {(imagePreviewURL || imageFile) && (
                    <Button
                      type="button"
                      variant="outlined"
                      onClick={handleRemoveImage}
                      sx={{ color: 'error.main', borderColor: 'error.main', '&:hover': { bgcolor: 'error.light' } }}
                    >
                      {tp('removeImage')}
                    </Button>
                  )}
                </Box>
                {uploadError && <Box sx={{ color: 'error.main', fontSize: '14px', mt: '4px', display: 'block' }}>{uploadError}</Box>}
              </Box>
              <Box sx={{ mb: '16px' }}>
                <Box component="label" htmlFor="name" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('name')}</Box>
                <OutlinedInput size="small" id="name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required fullWidth />
              </Box>
              <Box sx={{ mb: '16px' }}>
                <Box component="label" htmlFor="description" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('description')}</Box>
                <OutlinedInput size="small" multiline rows={3} id="description" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} fullWidth />
              </Box>
              <Box sx={{ mb: '16px' }}>
                <Box component="label" htmlFor="price" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('price')}</Box>
                <OutlinedInput size="small" id="price" type="number" inputProps={{ step: '1', min: 0 }} value={form.price ?? ''} onChange={(e) => setForm({ ...form, price: Number(e.target.value) || 0 })} required fullWidth />
              </Box>
              <Box sx={{ mb: '8px' }}>
                <Box component="label" htmlFor="originalPrice" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{tp('originalPrice')}</Box>
                <OutlinedInput
                  size="small"
                  id="originalPrice"
                  type="number"
                  inputProps={{ step: '1', min: 0 }}
                  placeholder={tp('originalPricePlaceholder')}
                  value={form.originalPrice === '' ? '' : form.originalPrice}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      originalPrice: e.target.value === '' ? '' : Number(e.target.value) || 0,
                    })
                  }
                  fullWidth
                />
              </Box>
            </DialogContent>
            <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
              <Button type="button" variant="outlined" onClick={closeForm}>
                {t('cancel')}
              </Button>
              <Button type="submit" variant="contained" disableElevation disabled={isSubmitting || createMutation.isLoading || updateMutation.isLoading} startIcon={isSubmitting || createMutation.isLoading || updateMutation.isLoading ? <CircularProgress size={16} color="inherit" /> : null}>
                {editingProduct ? t('save') : t('create')}
              </Button>
            </DialogActions>
          </Box>
        </Dialog>
      </Container>
  )
}
