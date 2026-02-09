"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Text, Textarea } from 'theme-ui'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import AddButton from '@/components/AddButton'
import { api } from '@/utils/api'

type Product = {
  id: number
  name: string
  description: string
  price: number
  original_price?: number | null
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

  // Debounce search: only trigger API after user stops typing for 300ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

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
    async (payload: FormState) => {
      const data: Parameters<typeof api.createProduct>[0] = {
        name: payload.name,
        description: payload.description || undefined,
        price: payload.price,
      }
      if (payload.originalPrice !== '') data.original_price = payload.originalPrice
      const res = await api.createProduct(data)
      if (!res.success) throw new Error(res.message || tp('createFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['products'])
        closeForm()
      },
    }
  )

  const updateMutation = useMutation(
    async ({ id, payload }: { id: number; payload: Partial<FormState> }) => {
      const data: Parameters<typeof api.updateProduct>[1] = {}
      if (payload.name !== undefined) data.name = payload.name
      if (payload.description !== undefined) data.description = payload.description
      if (payload.price !== undefined) data.price = payload.price
      if (payload.originalPrice !== undefined && payload.originalPrice !== '') data.original_price = payload.originalPrice
      const res = await api.updateProduct(id, data)
      if (!res.success) throw new Error(res.message || tp('updateFailed'))
      return res
    },
    {
      onSuccess: () => {
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
    setIsFormOpen(true)
  }

  function closeForm() {
    setIsFormOpen(false)
    setForm(emptyForm)
    setEditingProduct(null)
  }

  function submitForm(e: React.FormEvent) {
    e.preventDefault()
    if (editingProduct) {
      updateMutation.mutate({ id: editingProduct.id, payload: form })
    } else {
      createMutation.mutate(form)
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
                  <Flex sx={{ gap: 2, alignItems: 'center' }}>
                    <SearchInput
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      placeholder={tp('searchPlaceholder')}
                    />
                    <AddButton onClick={openCreateForm} title={tp('addProduct')} />
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
                          }}>
                            {p.name.charAt(0).toUpperCase()}
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
                      {/* Header with avatar */}
                      <Flex sx={{ alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: 3, mb: 4, pb: 4, borderBottom: '1px solid', borderColor: 'border' }}>
                        <Flex sx={{ alignItems: 'center', gap: 3 }}>
                          <Box
                            sx={{
                              width: 72,
                              height: 72,
                              borderRadius: 'round',
                              background: 'linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%)',
                              color: 'white',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 'bold',
                              fontSize: 5,
                              flexShrink: 0,
                            }}
                          >
                            {selectedProduct.name.charAt(0).toUpperCase()}
                          </Box>
                          <Box>
                            <Heading as="h2" sx={{ fontSize: 4, fontWeight: 700, mb: 1, letterSpacing: '-0.02em' }}>
                              {selectedProduct.name}
                            </Heading>
                            <Text sx={{ fontSize: 2, fontWeight: 600, color: 'primary' }}>
                              Rp. {selectedProduct.price.toLocaleString()}
                            </Text>
                          </Box>
                        </Flex>
                        <Flex sx={{ gap: 2 }}>
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
                    <Box sx={{ fontSize: 6, opacity: 0.4 }}>📦</Box>
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

