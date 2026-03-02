"use client"

import { useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, OutlinedInput, Paper, Typography } from '@mui/material'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import AddButton from '@/components/AddButton'
import { api, resolveImageURL } from '@/utils/api'
import { Share2, Check, Package, X } from 'lucide-react'

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
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [shareCopied, setShareCopied] = useState(false)
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [imagePreviewURL, setImagePreviewURL] = useState<string | null>(null)
  const [imageRemoved, setImageRemoved] = useState(false)
  const [uploadError, setUploadError] = useState<string | null>(null)
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

  const { data: productsRes, isLoading, isError, error } = useQuery(
    ['products', debouncedSearch],
    async () => {
      const res = await api.getProducts(debouncedSearch || undefined)
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

  // Set default selection when data loads
  useEffect(() => {
    if (!selectedProductId && productsRes && productsRes.length > 0) {
      setSelectedProductId(productsRes[0].id)
    }
  }, [productsRes, selectedProductId])

  const selectedProduct: Product | null = useMemo(() => {
    if (!productsRes) return null
    return productsRes.find((p) => p.id === selectedProductId) || null
  }, [productsRes, selectedProductId])

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

  return (
    <Layout>
      <Container disableGutters sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden', display: 'flex' }}>
          {isLoading && <Box>{t('loading')}</Box>}
          {isError && (
            <Box sx={{ color: 'error.main' }}>{(error as Error)?.message || tErrors('loadingError', { resource: tp('title') })}</Box>
          )}

          {!isLoading && !isError && (
            <Box sx={{ overflow: 'hidden', bgcolor: 'transparent', flex: 1, minHeight: 0, display: 'flex' }}>
              {/* Left list (compact like side menu) */}
              <Box sx={{ width: { xs: '100%', sm: '300px' }, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRight: { xs: 'none', sm: '1px solid' }, borderColor: 'grey.200' }}>
                <Box sx={{ p: '24px', flexShrink: 0 }}>
                  <Box sx={{ gap: '8px', alignItems: 'center', flexWrap: 'wrap', display: 'flex' }}>
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
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(productsRes || []).map((p) => {
                    const isActive = p.id === selectedProductId
                    return (
                      <Box
                        key={p.id}
                        sx={{
                          py: '16px',
                          px: '24px',
                          cursor: 'pointer',
                          textAlign: 'left',
                          bgcolor: isActive ? 'grey.100' : 'transparent',
                          borderRadius: '8px',
                          '&:hover': { bgcolor: isActive ? 'grey.100' : 'grey.50' },
                        }}
                        onClick={() => setSelectedProductId(p.id)}
                      >
                        <Box sx={{ flexDirection: 'row', alignItems: 'center', gap: '8px', display: 'flex' }}>
                          <Box sx={{
                            width: 36,
                            height: 36,
                            borderRadius: '50%',
                            bgcolor: 'primary.main',
                            color: 'white',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontWeight: 700,
                            fontSize: '14px',
                            flexShrink: 0,
                            overflow: 'hidden',
                          }}>
                            {p.image_url
                              ? <img src={resolveImageURL(p.image_url)} alt={p.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                              : p.name.charAt(0).toUpperCase()
                            }
                          </Box>
                          <Box sx={{ fontSize: '12px', lineHeight: 1, wordBreak: 'break-word' }}>{p.name}</Box>
                        </Box>
                      </Box>
                    )
                  })}
                  {(productsRes || []).length === 0 && (
                    <Box sx={{ p: '16px', color: 'grey.500', textAlign: 'center' }}>{tp('noProducts')}</Box>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bgcolor: 'grey.50' }}>
                {selectedProduct ? (
                  <Box sx={{ maxWidth: 640, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
                    <Paper
                      sx={{
                        p: '24px',
                        borderRadius: '12px',
                        boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'white',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)' },
                      }}
                    >
                      {/* Header */}
                      <Box sx={{ alignItems: 'flex-start', justifyContent: 'space-between', gap: '16px', mb: '24px', pb: '24px', borderBottom: '1px solid', borderColor: 'grey.200', display: 'flex' }}>
                        {/* Image + name/price stacked below */}
                        <Box>
                          <Box
                            sx={{
                              width: 300,
                              height: 300,
                              borderRadius: '8px',
                              background: 'linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%)',
                              color: 'white',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 700,
                              fontSize: '24px',
                              overflow: 'hidden',
                              mb: '16px',
                            }}
                          >
                            {selectedProduct.image_url
                              ? <img src={resolveImageURL(selectedProduct.image_url)} alt={selectedProduct.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                              : selectedProduct.name.charAt(0).toUpperCase()
                            }
                          </Box>
                          <Typography component="h2" sx={{ fontSize: '20px', fontWeight: 700, mb: '4px', letterSpacing: '-0.02em' }}>
                            {selectedProduct.name}
                          </Typography>
                          <Box sx={{ fontSize: '16px', fontWeight: 600, color: 'primary.main' }}>
                            Rp. {selectedProduct.price.toLocaleString()}
                          </Box>
                        </Box>
                        {/* Action buttons on the right */}
                        <Box sx={{ gap: '8px', flexShrink: 0, display: 'flex' }}>
                          <Button variant="outlined" onClick={() => openEditForm(selectedProduct)}>
                            {t('edit')}
                          </Button>
                          <Button
                            variant="outlined"
                            onClick={() => {
                              if (confirm(tp('deleteConfirm'))) deleteMutation.mutate(selectedProduct.id)
                            }}
                            sx={{
                              bgcolor: 'transparent',
                              color: 'error.main',
                              border: '2px solid',
                              borderColor: 'error.main',
                              '&:hover': { bgcolor: 'error.light' },
                            }}
                          >
                            {t('delete')}
                          </Button>
                        </Box>
                      </Box>

                      {/* Description */}
                      <Box>
                        <Box sx={{ fontWeight: 600, fontSize: '16px', color: 'grey.500', mb: '4px', display: 'block' }}>
                          {t('description')}
                        </Box>
                        <Box sx={{ fontSize: '14px', lineHeight: 1.6, wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
                          {selectedProduct.description || '—'}
                        </Box>
                      </Box>

                      {/* Original price */}
                      <Box sx={{ mt: '24px' }}>
                        <Box sx={{ fontWeight: 600, fontSize: '16px', color: 'grey.500', mb: '4px', display: 'block' }}>
                          {tp('originalPrice')}
                        </Box>
                        <Box sx={{ fontSize: '14px', color: 'grey.500' }}>
                          Rp. {(selectedProduct.original_price ?? selectedProduct.price).toLocaleString()}
                        </Box>
                      </Box>
                    </Paper>
                  </Box>
                ) : (
                  <Box
                    sx={{
                      height: '100%',
                      minHeight: 320,
                      alignItems: 'center',
                      justifyContent: 'center',
                      flexDirection: 'column',
                      gap: '8px',
                      color: 'grey.500',
                      display: 'flex',
                    }}
                  >
                    <Package size={48} opacity={0.4} />
                    <Box sx={{ fontSize: '16px' }}>{tp('selectProduct')}</Box>
                    <Box sx={{ fontSize: '14px' }}>{tp('chooseFromList')}</Box>
                  </Box>
                )}
              </Box>
            </Box>
          )}
        </Box>

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
              <Button type="submit" variant="contained" disableElevation disabled={createMutation.isLoading || updateMutation.isLoading}>
                {editingProduct ? t('save') : t('create')}
              </Button>
            </DialogActions>
          </Box>
        </Dialog>
      </Container>
    </Layout>
  )
}
