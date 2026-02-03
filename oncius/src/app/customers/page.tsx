"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Text, Textarea } from 'theme-ui'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import { api } from '@/utils/api'

type Customer = {
  id: number
  name: string
  phone: string
  address: string
  created_at?: string
  updated_at?: string | null
}

type FormState = {
  name: string
  phone: string
  address: string
}

const emptyForm: FormState = { name: '', phone: '', address: '' }

export default function CustomersPage() {
  const tc = useTranslations('customers')
  const queryClient = useQueryClient()
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [selectedCustomerId, setSelectedCustomerId] = useState<number | null>(null)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Debounce search: only trigger API after user stops typing for 300ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const { data: customersRes, isLoading, isError, error } = useQuery(
    ['customers', debouncedSearch],
    async () => {
      const res = await api.getCustomers(debouncedSearch || undefined)
      if (!res.success) throw new Error(res.message || tc('fetchFailed'))
      return res.data as Customer[]
    },
    { keepPreviousData: true }
  )

  const createMutation = useMutation(
    async (payload: FormState) => {
      const res = await api.createCustomer(payload)
      if (!res.success) throw new Error(res.message || 'Failed to create customer')
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['customers'])
        closeForm()
      },
    }
  )

  const updateMutation = useMutation(
    async ({ id, payload }: { id: number; payload: Partial<FormState> }) => {
      const res = await api.updateCustomer(id, payload)
      if (!res.success) throw new Error(res.message || 'Failed to update customer')
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['customers'])
        closeForm()
      },
    }
  )

  const deleteMutation = useMutation(
    async (id: number) => {
      const res = await api.deleteCustomer(id)
      if (!res.success) throw new Error(res.message || 'Failed to delete customer')
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['customers'])
      },
    }
  )

  // Set default selection when data loads
  useEffect(() => {
    if (!selectedCustomerId && customersRes && customersRes.length > 0) {
      setSelectedCustomerId(customersRes[0].id)
    }
  }, [customersRes, selectedCustomerId])

  const selectedCustomer: Customer | null = useMemo(() => {
    if (!customersRes) return null
    return customersRes.find((c) => c.id === selectedCustomerId) || null
  }, [customersRes, selectedCustomerId])

  function openCreateForm() {
    setEditingCustomer(null)
    setForm(emptyForm)
    setIsFormOpen(true)
  }

  function openEditForm(customer: Customer) {
    setEditingCustomer(customer)
    setForm({ name: customer.name, phone: customer.phone, address: customer.address })
    setIsFormOpen(true)
  }

  function closeForm() {
    setIsFormOpen(false)
    setForm(emptyForm)
    setEditingCustomer(null)
  }

  function submitForm(e: React.FormEvent) {
    e.preventDefault()
    if (editingCustomer) {
      updateMutation.mutate({ id: editingCustomer.id, payload: form })
    } else {
      createMutation.mutate(form)
    }
  }

  return (
    <Layout>
      <Container sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Flex sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden' }}>
          {isLoading && <Text>Loading...</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>{(error as Error)?.message || 'Error loading customers'}</Text>
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
                      placeholder={tc('searchPlaceholder')}
                    />
                    <Button
                      onClick={openCreateForm}
                      sx={{
                        width: 44,
                        minWidth: 44,
                        height: 44,
                        p: 0,
                        display: 'inline-flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        borderRadius: 'medium',
                        fontSize: 3,
                        fontWeight: 'bold',
                      }}
                      title={tc('addCustomer')}
                    >
                      +
                    </Button>
                  </Flex>
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(customersRes || []).map((c) => {
                    const isActive = c.id === selectedCustomerId
                    return (
                      <Box
                        key={c.id}
                        sx={{
                          py: 3,
                          px: 4,
                          cursor: 'pointer',
                          textAlign: 'left',
                          bg: isActive ? 'backgroundLight' : 'transparent',
                          borderRadius: 'medium',
                          '&:hover': { bg: isActive ? 'backgroundLight' : 'background.secondary' },
                        }}
                        onClick={() => setSelectedCustomerId(c.id)}
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
                            {c.name.charAt(0).toUpperCase()}
                          </Box>
                          <Text sx={{ fontSize: 0, lineHeight: 1, wordBreak: 'break-word' }}>{c.name}</Text>
                        </Flex>
                      </Box>
                    )
                  })}
                  {(customersRes || []).length === 0 && (
                    <Text sx={{ p: 3, color: 'text.secondary', textAlign: 'center' }}>No customers</Text>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bg: 'background.secondary' }}>
                {selectedCustomer ? (
                  <Box sx={{ maxWidth: 640, mx: 'auto', p: [4, 5] }}>
                    {/* Header card with avatar */}
                    <Card
                      sx={{
                        p: 4,
                        mb: 4,
                        borderRadius: 'large',
                        boxShadow: 'medium',
                        border: 'none',
                        background: 'linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%)',
                        color: 'white',
                      }}
                    >
                      <Flex sx={{ alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: 3 }}>
                        <Flex sx={{ alignItems: 'center', gap: 3 }}>
                          <Box
                            sx={{
                              width: 72,
                              height: 72,
                              borderRadius: 'round',
                              bg: 'rgba(255,255,255,0.25)',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 'bold',
                              fontSize: 5,
                              flexShrink: 0,
                            }}
                          >
                            {selectedCustomer.name.charAt(0).toUpperCase()}
                          </Box>
                          <Box>
                            <Heading as="h2" sx={{ fontSize: 4, fontWeight: 700, mb: 1, letterSpacing: '-0.02em' }}>
                              {selectedCustomer.name}
                            </Heading>
                            {selectedCustomer.created_at && (
                              <Text sx={{ fontSize: 0, opacity: 0.9 }}>
                                Customer since {new Date(selectedCustomer.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
                              </Text>
                            )}
                          </Box>
                        </Flex>
                        <Flex sx={{ gap: 2 }}>
                          <Button
                            variant="secondary"
                            onClick={() => openEditForm(selectedCustomer)}
                            sx={{
                              bg: 'rgba(255,255,255,0.2)',
                              border: '1px solid rgba(255,255,255,0.5)',
                              color: 'white',
                              '&:hover': { bg: 'rgba(255,255,255,0.3)', borderColor: 'white' },
                            }}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => {
                              if (confirm('Delete this customer?')) deleteMutation.mutate(selectedCustomer.id)
                            }}
                            sx={{
                              bg: 'rgba(239,68,68,0.3)',
                              border: '1px solid rgba(239,68,68,0.6)',
                              color: 'white',
                              '&:hover': { bg: 'rgba(239,68,68,0.5)', borderColor: '#ef4444' },
                            }}
                          >
                            Delete
                          </Button>
                        </Flex>
                      </Flex>
                    </Card>

                    {/* Contact info card */}
                    <Card
                      sx={{
                        p: 2,
                        borderRadius: 'large',
                        boxShadow: 'small',
                        border: '1px solid',
                        borderColor: 'border',
                        bg: 'white',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: 'medium' },
                      }}
                    >
                      <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                        <Flex sx={{ alignItems: 'center' }}>
                          <Box
                            sx={{
                              width: 44,
                              height: 44,
                              borderRadius: 'medium',
                              bg: 'background.secondary',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontSize: 3,
                              flexShrink: 0,
                            }}
                          >
                            📞
                          </Box>
                          <Box sx={{ flex: 1, minWidth: 0 }}>
                            <Text sx={{ fontSize: 1, lineHeight: 1.5, wordBreak: 'break-word' }}>
                              {selectedCustomer.phone || '—'}
                            </Text>
                          </Box>
                        </Flex>
                        <Flex sx={{ alignItems: 'center' }}>
                          <Box
                            sx={{
                              width: 44,
                              height: 44,
                              borderRadius: 'medium',
                              bg: 'background.secondary',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontSize: 3,
                              flexShrink: 0,
                            }}
                          >
                            📍
                          </Box>
                          <Box sx={{ flex: 1, minWidth: 0 }}>
                            <Text sx={{ fontSize: 1, lineHeight: 1.6, wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
                              {selectedCustomer.address || '—'}
                            </Text>
                          </Box>
                        </Flex>
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
                    <Box sx={{ fontSize: 6, opacity: 0.4 }}>👤</Box>
                    <Text sx={{ fontSize: 2 }}>Select a customer to view details</Text>
                    <Text sx={{ fontSize: 1 }}>Choose from the list on the left</Text>
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
                {editingCustomer ? 'Edit Customer' : 'New Customer'}
              </Heading>
              <Box as="form" onSubmit={submitForm}>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="name">Name</Label>
                  <Input id="name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="phone">Phone</Label>
                  <Input id="phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} required />
                </Box>
                <Box sx={{ mb: 4 }}>
                  <Label htmlFor="address">Address</Label>
                  <Textarea id="address" rows={3} value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} required />
                </Box>
                <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
                  <Button type="button" variant="secondary" onClick={closeForm}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createMutation.isLoading || updateMutation.isLoading}>
                    {editingCustomer ? 'Save Changes' : 'Create'}
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

