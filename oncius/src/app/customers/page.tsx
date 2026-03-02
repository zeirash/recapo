"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, OutlinedInput, Paper, Typography } from '@mui/material'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import AddButton from '@/components/AddButton'
import { api } from '@/utils/api'
import { Phone, MapPin, User } from 'lucide-react'

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
      <Container disableGutters sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden', display: 'flex' }}>
          {isLoading && <Box>Loading...</Box>}
          {isError && (
            <Box sx={{ color: 'error.main' }}>{(error as Error)?.message || 'Error loading customers'}</Box>
          )}

          {!isLoading && !isError && (
            <Box sx={{ overflow: 'hidden', bgcolor: 'transparent', flex: 1, minHeight: 0, display: 'flex' }}>
              {/* Left list (compact like side menu) */}
              <Box sx={{ width: { xs: '100%', sm: '300px' }, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRight: { xs: 'none', sm: '1px solid' }, borderColor: 'grey.200' }}>
                <Box sx={{ p: '24px', flexShrink: 0 }}>
                  <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                    <SearchInput
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      placeholder={tc('searchPlaceholder')}
                    />
                    <AddButton onClick={openCreateForm} title={tc('addCustomer')} />
                  </Box>
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(customersRes || []).map((c) => {
                    const isActive = c.id === selectedCustomerId
                    return (
                      <Box
                        key={c.id}
                        sx={{
                          py: '16px',
                          px: '24px',
                          cursor: 'pointer',
                          textAlign: 'left',
                          bgcolor: isActive ? 'grey.100' : 'transparent',
                          borderRadius: '8px',
                          '&:hover': { bgcolor: isActive ? 'grey.100' : 'grey.50' },
                        }}
                        onClick={() => setSelectedCustomerId(c.id)}
                      >
                        <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center', gap: '8px' }}>
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
                          }}>
                            {c.name.charAt(0).toUpperCase()}
                          </Box>
                          <Box sx={{ fontSize: '12px', lineHeight: 1, wordBreak: 'break-word' }}>{c.name}</Box>
                        </Box>
                      </Box>
                    )
                  })}
                  {(customersRes || []).length === 0 && (
                    <Box sx={{ p: '16px', color: 'grey.500', textAlign: 'center' }}>No customers</Box>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bgcolor: 'grey.50' }}>
                {selectedCustomer ? (
                  <Box sx={{ maxWidth: 640, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
                    {/* Header card with avatar */}
                    <Paper
                      sx={{
                        p: '24px',
                        mb: '24px',
                        borderRadius: '12px',
                        boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)',
                        border: 'none',
                        background: 'linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%)',
                        color: 'white',
                      }}
                    >
                      <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: '16px' }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                          <Box
                            sx={{
                              width: 72,
                              height: 72,
                              borderRadius: '50%',
                              bgcolor: 'rgba(255,255,255,0.25)',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 700,
                              fontSize: '24px',
                              flexShrink: 0,
                            }}
                          >
                            {selectedCustomer.name.charAt(0).toUpperCase()}
                          </Box>
                          <Box>
                            <Typography component="h2" sx={{ fontSize: '20px', fontWeight: 700, mb: '4px', letterSpacing: '-0.02em' }}>
                              {selectedCustomer.name}
                            </Typography>
                            {selectedCustomer.created_at && (
                              <Box sx={{ fontSize: '12px', opacity: 0.9 }}>
                                Customer since {new Date(selectedCustomer.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
                              </Box>
                            )}
                          </Box>
                        </Box>
                        <Box sx={{ display: 'flex', gap: '8px' }}>
                          <Button
                            variant="outlined"
                            onClick={() => openEditForm(selectedCustomer)}
                            sx={{
                              bgcolor: 'rgba(255,255,255,0.2)',
                              border: '1px solid rgba(255,255,255,0.5)',
                              color: 'white',
                              '&:hover': { bgcolor: 'rgba(255,255,255,0.3)', borderColor: 'white' },
                            }}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="outlined"
                            onClick={() => {
                              if (confirm('Delete this customer?')) deleteMutation.mutate(selectedCustomer.id)
                            }}
                            sx={{
                              bgcolor: 'rgba(239,68,68,0.3)',
                              border: '1px solid rgba(239,68,68,0.6)',
                              color: 'white',
                              '&:hover': { bgcolor: 'rgba(239,68,68,0.5)', borderColor: 'error.main' },
                            }}
                          >
                            Delete
                          </Button>
                        </Box>
                      </Box>
                    </Paper>

                    {/* Contact info card */}
                    <Paper
                      sx={{
                        p: '8px',
                        borderRadius: '12px',
                        boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'white',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' },
                      }}
                    >
                      <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', px: '8px', py: '10px' }}>
                          <Phone size={14} style={{ flexShrink: 0, color: '#6b7280' }} />
                          <Box sx={{ flex: 1, minWidth: 0 }}>
                            <Box sx={{ fontSize: '14px', lineHeight: 1.5, wordBreak: 'break-word' }}>
                              {selectedCustomer.phone || '—'}
                            </Box>
                          </Box>
                        </Box>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', px: '8px', py: '10px' }}>
                          <MapPin size={14} style={{ flexShrink: 0, color: '#6b7280' }} />
                          <Box sx={{ flex: 1, minWidth: 0 }}>
                            <Box sx={{ fontSize: '14px', lineHeight: 1.6, wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
                              {selectedCustomer.address || '—'}
                            </Box>
                          </Box>
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
                    <User size={48} opacity={0.4} />
                    <Box sx={{ fontSize: '16px' }}>Select a customer to view details</Box>
                    <Box sx={{ fontSize: '14px' }}>Choose from the list on the left</Box>
                  </Box>
                )}
              </Box>
            </Box>
          )}
        </Box>

        {isFormOpen && (
          <Box
            sx={{
              position: 'fixed',
              inset: 0,
              bgcolor: 'rgba(0,0,0,0.4)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              p: '16px',
            }}
            onClick={(e) => {
              if (e.target === e.currentTarget) closeForm()
            }}
          >
            <Paper sx={{ width: { xs: '100%', sm: '540px' }, p: '24px' }}>
              <Typography component="h3" sx={{ mb: '16px' }}>
                {editingCustomer ? 'Edit Customer' : 'New Customer'}
              </Typography>
              <Box component="form" onSubmit={submitForm}>
                <Box sx={{ mb: '16px' }}>
                  <Box component="label" htmlFor="name" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>Name</Box>
                  <OutlinedInput size="small" fullWidth id="name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
                </Box>
                <Box sx={{ mb: '16px' }}>
                  <Box component="label" htmlFor="phone" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>Phone</Box>
                  <OutlinedInput size="small" fullWidth id="phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} required />
                </Box>
                <Box sx={{ mb: '24px' }}>
                  <Box component="label" htmlFor="address" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>Address</Box>
                  <OutlinedInput size="small" fullWidth multiline rows={3} id="address" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} required />
                </Box>
                <Box sx={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                  <Button type="button" variant="outlined" onClick={closeForm}>
                    Cancel
                  </Button>
                  <Button type="submit" variant="contained" disableElevation disabled={createMutation.isLoading || updateMutation.isLoading}>
                    {editingCustomer ? 'Save Changes' : 'Create'}
                  </Button>
                </Box>
              </Box>
            </Paper>
          </Box>
        )}
      </Container>
    </Layout>
  )
}
