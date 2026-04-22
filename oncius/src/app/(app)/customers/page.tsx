"use client"

import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, IconButton, InputBase, Menu, MenuItem, OutlinedInput, Paper, Tooltip, Typography } from '@mui/material'
import AddButton from '@/components/ui/AddButton'
import { api } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { ArrowUpDown, MapPin, Pencil, Phone, Search, Trash2, User } from 'lucide-react'

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

const FILTER_SESSION_KEY = 'customers_filter_state'

function getStoredFilterState() {
  if (typeof window === 'undefined') return null
  try {
    const s = sessionStorage.getItem(FILTER_SESSION_KEY)
    return s ? JSON.parse(s) : null
  } catch {
    return null
  }
}

export default function CustomersPage() {
  const tc = useTranslations('customers')
  const t = useTranslations('common')
  const queryClient = useQueryClient()
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [searchInput, setSearchInput] = useState<string>(() => getStoredFilterState()?.searchInput ?? '')
  const [debouncedSearch, setDebouncedSearch] = useState<string>(() => getStoredFilterState()?.searchInput ?? '')
  const [sortValue, setSortValue] = useState<string>(() => getStoredFilterState()?.sortValue ?? 'updated_at,desc')
  const [sortMenuAnchor, setSortMenuAnchor] = useState<null | HTMLElement>(null)
  const [customerToDelete, setCustomerToDelete] = useState<Customer | null>(null)
  const [deleteActiveOrderId, setDeleteActiveOrderId] = useState<number | null>(null)
  const [checkingActiveOrder, setCheckingActiveOrder] = useState(false)

  useEffect(() => {
    sessionStorage.setItem(FILTER_SESSION_KEY, JSON.stringify({ searchInput, sortValue }))
  }, [searchInput, sortValue])

  // Debounce search: only trigger API after user stops typing for 300ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const sortParam = sortValue || undefined

  const { data: customersRes, isLoading, isError, error } = useQuery(
    ['customers', debouncedSearch, sortParam],
    async () => {
      const res = await api.getCustomers(debouncedSearch || undefined, sortParam)
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

  async function handleDeleteClick(customer: Customer) {
    setCheckingActiveOrder(true)
    setCustomerToDelete(customer)
    try {
      const res = await api.checkActiveOrder({ phone: customer.phone, name: customer.name })
      setDeleteActiveOrderId(res.data?.active_order_id ?? null)
    } catch {
      setDeleteActiveOrderId(null)
    } finally {
      setCheckingActiveOrder(false)
    }
  }

  function closeDeleteDialog() {
    setCustomerToDelete(null)
    setDeleteActiveOrderId(null)
  }

  function confirmDelete() {
    if (customerToDelete) deleteMutation.mutate(customerToDelete.id)
    closeDeleteDialog()
  }

  const customers = customersRes || []

  return (
    <Container disableGutters maxWidth={false} sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        {/* Top bar */}
        <Box sx={{ p: '24px', flexShrink: 0 }}>
          <Box sx={{ maxWidth: 960, mx: 'auto' }}>
            <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
            <Paper
              variant="outlined"
              sx={{ flex: 1, display: 'flex', alignItems: 'center', px: '8px', py: '4px', borderRadius: '10px', boxShadow: 'none', gap: '2px' }}
            >
              <Tooltip title={tc('sortLabel') || 'Sort'}>
                <IconButton
                  size="small"
                  onClick={(e) => setSortMenuAnchor(e.currentTarget)}
                  sx={{ color: sortValue !== 'updated_at,desc' ? 'primary.main' : 'text.secondary', '&:hover': { color: sortValue !== 'updated_at,desc' ? 'primary.dark' : 'text.primary' } }}
                >
                  <ArrowUpDown size={18} />
                </IconButton>
              </Tooltip>

              <InputBase
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                placeholder={tc('searchPlaceholder')}
                sx={{ flex: 1, mx: '8px', fontSize: '13px', color: 'text.primary' }}
                inputProps={{ 'aria-label': tc('searchPlaceholder') }}
              />

              <Search size={18} style={{ color: '#9e9e9e', flexShrink: 0, marginRight: '4px' }} />
            </Paper>
            <AddButton onClick={openCreateForm} title={tc('addCustomer')} />
            </Box>

            <Menu
              anchorEl={sortMenuAnchor}
              open={!!sortMenuAnchor}
              onClose={() => setSortMenuAnchor(null)}
              anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
              transformOrigin={{ vertical: 'top', horizontal: 'left' }}
            >
              {([
                ['updated_at,desc', tc('sortUpdatedAtDesc')],
                ['updated_at,asc', tc('sortUpdatedAtAsc')],
                ['name,asc', tc('sortNameAsc')],
                ['name,desc', tc('sortNameDesc')],
              ] as const).map(([val, label]) => (
                <MenuItem
                  key={val}
                  selected={sortValue === val}
                  onClick={() => { setSortValue(val); setSortMenuAnchor(null) }}
                  sx={{ fontSize: '13px' }}
                >
                  {label}
                </MenuItem>
              ))}
            </Menu>
          </Box>
        </Box>

        {/* Scrollable body */}
        <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', px: '24px', pb: '24px' }}>
          {isLoading && <PageLoadingSkeleton />}
          {isError && (
            <Box sx={{ color: 'error.main' }}>{(error as Error)?.message || 'Error loading customers'}</Box>
          )}

          {/* Empty state */}
          {!isLoading && !isError && customers.length === 0 && (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: '8px', color: 'text.secondary', minHeight: 320 }}>
              <User size={48} opacity={0.4} />
              <Typography>{tc('noCustomers')}</Typography>
            </Box>
          )}

          {/* Customer list */}
          {!isLoading && !isError && customers.length > 0 && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: '8px', maxWidth: 960, mx: 'auto' }}>
              {customers.map((c) => (
                <Paper
                  key={c.id}
                  elevation={0}
                  sx={{
                    border: '1px solid',
                    borderColor: 'divider',
                    borderRadius: '10px',
                    bgcolor: 'background.paper',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '16px',
                    p: '12px 16px',
                    '&:hover': { borderColor: 'divider', bgcolor: 'action.hover' },
                  }}
                >
                  {/* Avatar */}
                  <Box sx={{
                    width: 48,
                    height: 48,
                    borderRadius: '50%',
                    background: 'linear-gradient(135deg,rgb(92, 151, 245) 0%,rgb(26, 94, 239) 100%)',
                    color: 'white',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}>
                    <User size={24} />
                  </Box>

                  {/* Info */}
                  <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Typography sx={{ fontWeight: 600, fontSize: '14px', lineHeight: 1.3, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {c.name}
                    </Typography>
                    {c.phone && (
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px', mt: '2px' }}>
                        <Phone size={11} style={{ color: '#9ca3af', flexShrink: 0 }} />
                        <Box sx={{ fontSize: '12px', color: 'text.secondary', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                          {c.phone}
                        </Box>
                      </Box>
                    )}
                    {c.address && (
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px', mt: '2px' }}>
                        <MapPin size={11} style={{ color: '#9ca3af', flexShrink: 0 }} />
                        <Box sx={{ fontSize: '12px', color: 'text.secondary', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                          {c.address}
                        </Box>
                      </Box>
                    )}
                  </Box>

                  {/* Actions */}
                  <Box sx={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                    <Tooltip title="Edit">
                      <IconButton size="small" onClick={() => openEditForm(c)}>
                        <Pencil size={16} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('delete')}>
                      <IconButton
                        size="small"
                        onClick={() => handleDeleteClick(c)}
                        disabled={checkingActiveOrder && customerToDelete?.id === c.id}
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

        <Dialog open={!!customerToDelete && !checkingActiveOrder} onClose={closeDeleteDialog}>
          <DialogTitle>{tc('deleteConfirm')}</DialogTitle>
          {deleteActiveOrderId ? (
            <DialogContent>
              <DialogContentText sx={{ color: 'warning.dark', fontSize: '14px' }}>
                {tc('deleteActiveOrderWarning', { id: deleteActiveOrderId })}
              </DialogContentText>
            </DialogContent>
          ) : null}
          <DialogActions sx={{ px: '24px', pb: '16px', gap: '8px' }}>
            <Button variant="outlined" onClick={closeDeleteDialog}>
              {t('cancel')}
            </Button>
            <Button
              variant="contained"
              color="error"
              disableElevation
              disabled={deleteMutation.isLoading}
              onClick={confirmDelete}
            >
              {t('delete')}
            </Button>
          </DialogActions>
        </Dialog>

        <Dialog open={isFormOpen} onClose={closeForm} fullWidth maxWidth="sm">
          <Box component="form" onSubmit={submitForm}>
            <DialogTitle sx={{ pb: '8px' }}>{editingCustomer ? 'Edit Customer' : 'New Customer'}</DialogTitle>
            <DialogContent sx={{ pt: '8px', pb: 0 }}>
              <Box sx={{ mb: '16px' }}>
                <Box component="label" htmlFor="name" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>Name</Box>
                <OutlinedInput size="small" fullWidth id="name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
              </Box>
              <Box sx={{ mb: '16px' }}>
                <Box component="label" htmlFor="phone" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>Phone</Box>
                <OutlinedInput size="small" fullWidth id="phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} required />
              </Box>
              <Box sx={{ mb: '8px' }}>
                <Box component="label" htmlFor="address" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>Address</Box>
                <OutlinedInput size="small" fullWidth multiline rows={3} id="address" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
              </Box>
            </DialogContent>
            <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
              <Button type="button" variant="outlined" onClick={closeForm}>
                Cancel
              </Button>
              <Button type="submit" variant="contained" disableElevation disabled={createMutation.isLoading || updateMutation.isLoading}>
                {editingCustomer ? 'Save Changes' : 'Create'}
              </Button>
            </DialogActions>
          </Box>
        </Dialog>
      </Container>
  )
}
