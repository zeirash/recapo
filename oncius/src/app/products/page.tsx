"use client"

import { useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Text, Textarea } from 'theme-ui'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import AddButton from '@/components/AddButton'
import { api, resolveImageURL } from '@/utils/api'
import { Share2, Check, Package } from 'lucide-react'

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
      <Container sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Flex sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden' }}>
          {isLoading && <Text>{t('loading')}</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>{(error as Error)?.message || tErrors('loadingError', { resource: tp('title') })}</Text>
          )}

          {!isLoading && !isError && (
            <Flex sx={{ overflow: 'hidden', bg: 'transparent', flex: 1, minHeight: 0 }}>
              {/* Left list (compact like side menu) */}
              <Box sx={{ width: ['100%', '300px'], minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRight: ['none', '1px solid'], borderColor: 'border' }}>
                <Box sx={{ p: 4, flexShrink: 0 }}>
                  <Flex sx={{ gap: 2, alignItems: 'center', flexWrap: 'wrap' }}>
                    <SearchInput
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      placeholder={tp('searchPlaceholder')}
                    />
                    <AddButton onClick={openCreateForm} title={tp('addProduct')} />
                    <Button
                      variant="secondary"
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
                        borderRadius: 'medium',
                        fontSize: 1,
                      }}
                    >
                      {shareCopied ? <Check size={16} /> : <Share2 size={16} />}
                    </Button>
                  </Flex>
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(productsRes || []).map((p) => {
                    const isActive = p.id === selectedProductId
                    return (
                      <Box
                        key={p.id}
                        sx={{
                          py: 3,
                          px: 4,
                          cursor: 'pointer',
                          textAlign: 'left',
                          bg: isActive ? 'backgroundLight' : 'transparent',
                          borderRadius: 'medium',
                          '&:hover': { bg: isActive ? 'backgroundLight' : 'background.secondary' },
                        }}
                        onClick={() => setSelectedProductId(p.id)}
                      >
                        <Flex sx={{ flexDirection: 'row', alignItems: 'center', gap: 2 }}>
                          <Box sx={{
                            width: 36,
                            height: 36,
                            borderRadius: '50%',
                            bg: 'primary',
                            color: 'white',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontWeight: 'bold',
                            fontSize: 1,
                            flexShrink: 0,
                            overflow: 'hidden',
                          }}>
                            {p.image_url
                              ? <img src={resolveImageURL(p.image_url)} alt={p.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                              : p.name.charAt(0).toUpperCase()
                            }
                          </Box>
                          <Text sx={{ fontSize: 0, lineHeight: 1, wordBreak: 'break-word' }}>{p.name}</Text>
                        </Flex>
                      </Box>
                    )
                  })}
                  {(productsRes || []).length === 0 && (
                    <Text sx={{ p: 3, color: 'text.secondary', textAlign: 'center' }}>{tp('noProducts')}</Text>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bg: 'background.secondary' }}>
                {selectedProduct ? (
                  <Box sx={{ maxWidth: 640, mx: 'auto', p: [4, 5] }}>
                    <Card
                      sx={{
                        p: 4,
                        borderRadius: 'large',
                        boxShadow: 'medium',
                        border: '1px solid',
                        borderColor: 'border',
                        bg: 'white',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: 'large' },
                      }}
                    >
                      {/* Header */}
                      <Flex sx={{ alignItems: 'flex-start', justifyContent: 'space-between', gap: 3, mb: 4, pb: 4, borderBottom: '1px solid', borderColor: 'border' }}>
                        {/* Image + name/price stacked below */}
                        <Box>
                          <Box
                            sx={{
                              width: 300,
                              height: 300,
                              borderRadius: 'medium',
                              background: 'linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%)',
                              color: 'white',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 'bold',
                              fontSize: 5,
                              overflow: 'hidden',
                              mb: 3,
                            }}
                          >
                            {selectedProduct.image_url
                              ? <img src={resolveImageURL(selectedProduct.image_url)} alt={selectedProduct.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                              : selectedProduct.name.charAt(0).toUpperCase()
                            }
                          </Box>
                          <Heading as="h2" sx={{ fontSize: 4, fontWeight: 700, mb: 1, letterSpacing: '-0.02em' }}>
                            {selectedProduct.name}
                          </Heading>
                          <Text sx={{ fontSize: 2, fontWeight: 600, color: 'primary' }}>
                            Rp. {selectedProduct.price.toLocaleString()}
                          </Text>
                        </Box>
                        {/* Action buttons on the right */}
                        <Flex sx={{ gap: 2, flexShrink: 0 }}>
                          <Button variant="secondary" onClick={() => openEditForm(selectedProduct)}>
                            {t('edit')}
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => {
                              if (confirm(tp('deleteConfirm'))) deleteMutation.mutate(selectedProduct.id)
                            }}
                            sx={{
                              bg: 'transparent',
                              color: 'error',
                              border: '2px solid',
                              borderColor: 'error',
                              '&:hover': { bg: '#fef2f2' },
                            }}
                          >
                            {t('delete')}
                          </Button>
                        </Flex>
                      </Flex>

                      {/* Description */}
                      <Box>
                        <Text sx={{ fontWeight: 600, fontSize: 2, color: 'text.secondary', mb: 1, display: 'block' }}>
                          {t('description')}
                        </Text>
                        <Text sx={{ fontSize: 1, lineHeight: 1.6, wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
                          {selectedProduct.description || '—'}
                        </Text>
                      </Box>

                      {/* Original price */}
                      <Box sx={{ mt: 4 }}>
                        <Text sx={{ fontWeight: 600, fontSize: 2, color: 'text.secondary', mb: 1, display: 'block' }}>
                          {tp('originalPrice')}
                        </Text>
                        <Text sx={{ fontSize: 1, color: 'text.secondary' }}>
                          Rp. {(selectedProduct.original_price ?? selectedProduct.price).toLocaleString()}
                        </Text>
                      </Box>
                    </Card>
                  </Box>
                ) : (
                  <Flex
                    sx={{
                      height: '100%',
                      minHeight: 320,
                      alignItems: 'center',
                      justifyContent: 'center',
                      flexDirection: 'column',
                      gap: 2,
                      color: 'text.secondary',
                    }}
                  >
                    <Package size={48} opacity={0.4} />
                    <Text sx={{ fontSize: 2 }}>{tp('selectProduct')}</Text>
                    <Text sx={{ fontSize: 1 }}>{tp('chooseFromList')}</Text>
                  </Flex>
                )}
              </Box>
            </Flex>
          )}
        </Flex>

        {isFormOpen && (
          <Box
            sx={{
              position: 'fixed',
              inset: 0,
              bg: 'rgba(0,0,0,0.4)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              p: 3,
            }}
            onClick={(e) => {
              if (e.target === e.currentTarget) closeForm()
            }}
          >
            <Card sx={{ width: ['100%', '540px'] }}>
              <Heading as="h3" sx={{ mb: 3 }}>
                {editingProduct ? tp('editProduct') : tp('newProduct')}
              </Heading>
              <Box as="form" onSubmit={submitForm}>
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
                <Box sx={{ mb: 3 }}>
                  <Label>{tp('image')}</Label>
                  {imagePreviewURL && (
                    <Box sx={{ mb: 2, width: 120, height: 120, borderRadius: 'medium', overflow: 'hidden', border: '1px solid', borderColor: 'border' }}>
                      <img src={imagePreviewURL} alt="preview" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                    </Box>
                  )}
                  <Flex sx={{ gap: 2, alignItems: 'center' }}>
                    <Button type="button" variant="secondary" onClick={() => fileInputRef.current?.click()}>
                      {tp('chooseImage')}
                    </Button>
                    {(imagePreviewURL || imageFile) && (
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={handleRemoveImage}
                        sx={{ color: 'error', borderColor: 'error', '&:hover': { bg: '#fef2f2' } }}
                      >
                        {tp('removeImage')}
                      </Button>
                    )}
                  </Flex>
                  {uploadError && <Text sx={{ color: 'error', fontSize: 1, mt: 1, display: 'block' }}>{uploadError}</Text>}
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="name">{t('name')}</Label>
                  <Input id="name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="description">{t('description')}</Label>
                  <Textarea id="description" rows={3} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="price">{t('price')}</Label>
                  <Input id="price" type="number" step="1" min={0} value={form.price ?? ''} onChange={(e) => setForm({ ...form, price: Number(e.target.value) || 0 })} required />
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="originalPrice">{tp('originalPrice')}</Label>
                  <Input
                    id="originalPrice"
                    type="number"
                    step="1"
                    min={0}
                    placeholder={tp('originalPricePlaceholder')}
                    value={form.originalPrice === '' ? '' : form.originalPrice}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        originalPrice: e.target.value === '' ? '' : Number(e.target.value) || 0,
                      })
                    }
                  />
                </Box>
                <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
                  <Button type="button" variant="secondary" onClick={closeForm}>
                    {t('cancel')}
                  </Button>
                  <Button type="submit" disabled={createMutation.isLoading || updateMutation.isLoading}>
                    {editingProduct ? t('save') : t('create')}
                  </Button>
                </Flex>
              </Box>
            </Card>
          </Box>
        )}
      </Container>
    </Layout>
  )
}

