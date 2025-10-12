"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Text, Textarea } from 'theme-ui'
import Layout from '@/components/Layout'
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
  const queryClient = useQueryClient()
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [selectedCustomerId, setSelectedCustomerId] = useState<number | null>(null)

  const { data: customersRes, isLoading, isError, error } = useQuery(
    ['customers'],
    async () => {
      const res = await api.getCustomers()
      if (!res.success) throw new Error(res.message || 'Failed to fetch customers')
      return res.data as Customer[]
    }
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

  // Search removed temporarily; show all customers

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
      <Container>
        <Flex sx={{ minHeight: '100vh', flexDirection: 'column' }}>
          {isLoading && <Text>Loading...</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>{(error as Error)?.message || 'Error loading customers'}</Text>
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
                    + Customer
                  </Button>
                </Box>
                <Box sx={{ overflowY: 'auto' }}>
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
                          bg: isActive ? 'background.light' : 'transparent',
                          borderRadius: 'medium',
                          '&:hover': { bg: isActive ? 'background.light' : 'background.secondary' },
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
              <Box sx={{ flex: 1, p: 4, bg: 'white' }}>
                {selectedCustomer ? (
                  <>
                    <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
                      <Heading as="h2" sx={{ fontSize: 3 }}>{selectedCustomer.name}</Heading>
                      <Flex sx={{ gap: 2 }}>
                        <Button variant="secondary" onClick={() => openEditForm(selectedCustomer)}>Edit</Button>
                        <Button
                          variant="secondary"
                          onClick={() => {
                            if (confirm('Delete this customer?')) deleteMutation.mutate(selectedCustomer.id)
                          }}
                        >
                          Delete
                        </Button>
                      </Flex>
                    </Flex>
                    <Card sx={{ p: 3 }}>
                      <Box sx={{ mb: 2 }}>
                        <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Phone</Text>
                        <Text>{selectedCustomer.phone || '-'}</Text>
                      </Box>
                      <Box>
                        <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Address</Text>
                        <Text>{selectedCustomer.address || '-'}</Text>
                      </Box>
                    </Card>
                  </>
                ) : (
                  <Flex sx={{ height: '100%', alignItems: 'center', justifyContent: 'center', color: 'text.secondary' }}>
                    <Text>Select a customer to view details</Text>
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

