"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Text, Textarea } from 'theme-ui'
import Layout from '@/components/Layout'
import { api } from '@/utils/api'

type Product = {
  id: number
  name: string
  price: number
  created_at?: string
  updated_at?: string | null
}

type FormState = {
  name: string
  price: number
}

const emptyForm: FormState = { name: '', price: 0 }

export default function ProductsPage() {
  const queryClient = useQueryClient()
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null)

  const { data: productsRes, isLoading, isError, error } = useQuery(
    ['products'],
    async () => {
      const res = await api.getProducts()
      if (!res.success) throw new Error(res.message || 'Failed to fetch products')
      return res.data as Product[]
    }
  )

  const createMutation = useMutation(
    async (payload: FormState) => {
      const res = await api.createProduct(payload)
      if (!res.success) throw new Error(res.message || 'Failed to create product')
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
      const res = await api.updateProduct(id, payload)
      if (!res.success) throw new Error(res.message || 'Failed to update product')
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
      if (!res.success) throw new Error(res.message || 'Failed to delete product')
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
    setForm({ name: product.name, price: product.price })
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
      <Container>
        <Flex sx={{ minHeight: '100vh', flexDirection: 'column' }}>
          {isLoading && <Text>Loading...</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>{(error as Error)?.message || 'Error loading products'}</Text>
          )}

          {!isLoading && !isError && (
            <Flex sx={{ overflow: 'hidden', bg: 'transparent', flex: 1 }}>
              {/* Left list (compact like side menu) */}
              <Box sx={{ width: ['100%', '300px'], display: 'flex', flexDirection: 'column', borderRight: ['none', '1px solid'], borderColor: 'border' }}>
                <Box sx={{ p: 4 }}>
                  <Button
                    onClick={openCreateForm}
                    sx={{ width: '100%', whiteSpace: 'nowrap', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}
                  >
                    + Product
                  </Button>
                </Box>
                <Box sx={{ overflowY: 'auto' }}>
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
                    <Text sx={{ p: 3, color: 'text.secondary', textAlign: 'center' }}>No products</Text>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, p: 4, bg: 'white' }}>
                {selectedProduct ? (
                  <>
                    <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
                      <Heading as="h2" sx={{ fontSize: 3 }}>{selectedProduct.name}</Heading>
                      <Flex sx={{ gap: 2 }}>
                        <Button variant="secondary" onClick={() => openEditForm(selectedProduct)}>Edit</Button>
                        <Button
                          variant="secondary"
                          onClick={() => {
                            if (confirm('Delete this product?')) deleteMutation.mutate(selectedProduct.id)
                          }}
                        >
                          Delete
                        </Button>
                      </Flex>
                    </Flex>
                    <Card sx={{ p: 3 }}>
                      <Box sx={{ mb: 2 }}>
                        <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Price</Text>
                        <Text>{selectedProduct.price.toLocaleString()}</Text>
                      </Box>
                    </Card>
                  </>
                ) : (
                  <Flex sx={{ height: '100%', alignItems: 'center', justifyContent: 'center', color: 'text.secondary' }}>
                    <Text>Select a product to view details</Text>
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
                {editingProduct ? 'Edit Product' : 'New Product'}
              </Heading>
              <Box as="form" onSubmit={submitForm}>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="name">Name</Label>
                  <Input id="name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="price">Price</Label>
                  <Input id="price" type="number" step="0.01" value={form.price || ''} onChange={(e) => setForm({ ...form, price: Number(e.target.value) || 0 })} required />
                </Box>
                <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
                  <Button type="button" variant="secondary" onClick={closeForm}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createMutation.isLoading || updateMutation.isLoading}>
                    {editingProduct ? 'Save Changes' : 'Create'}
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

