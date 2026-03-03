"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, NativeSelect, OutlinedInput, Paper, Typography } from '@mui/material'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import AddButton from '@/components/AddButton'
import CustomerSearchSelect from '@/components/CustomerSearchSelect'
import { api, ApiError } from '@/utils/api'
import { ClipboardList, Trash2 } from 'lucide-react'

type OrderItem = {
  id: number
  order_id?: number
  product_name: string
  price: number
  qty: number
  created_at: string
  updated_at?: string | null
}

type Order = {
  id: number
  customer_name: string
  total_price: number
  status: string
  notes?: string
  order_items?: OrderItem[]
  created_at: string
  updated_at?: string | null
}

type Customer = {
  id: number
  name: string
  phone: string
  address: string
}

type Product = {
  id: number
  name: string
  price: number
}

type CreateOrderForm = {
  customer_id: number | null
  notes: string
}

type AddItemForm = {
  product_id: number | null
  qty: number
}

const emptyCreateForm: CreateOrderForm = { customer_id: null, notes: '' }
const emptyAddItemForm: AddItemForm = { product_id: null, qty: 1 }

const statusColors: Record<string, { bg: string; color: string }> = {
  created: { bg: '#E3F2FD', color: '#1565C0' },
  in_progress: { bg: '#FFF3E0', color: '#E65100' },
  in_delivery: { bg: '#F3E5F5', color: '#7B1FA2' },
  done: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FFEBEE', color: '#C62828' },
}

export default function OrdersPage() {
  const DEFAULT_STATUSES = ['created', 'in_progress'] as const
  const t = useTranslations('common')
  const to = useTranslations('orders')
  const toStatus = useTranslations('orderStatus')
  const tErrors = useTranslations('errors')
  const queryClient = useQueryClient()
  const [isCreateFormOpen, setIsCreateFormOpen] = useState(false)
  const [isAddItemFormOpen, setIsAddItemFormOpen] = useState(false)
  const [createForm, setCreateForm] = useState<CreateOrderForm>(emptyCreateForm)
  const [addItemForm, setAddItemForm] = useState<AddItemForm>(emptyAddItemForm)
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<string[]>([...DEFAULT_STATUSES])
  const [createFormConflict, setCreateFormConflict] = useState(false)

  // Debounce search: only trigger API after user stops typing for 300ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  // Fetch orders
  const { data: ordersRes, isLoading, isError, error } = useQuery(
    ['orders', debouncedSearch, statusFilter],
    async () => {
      const opts: { search?: string; status?: string } = {}
      if (debouncedSearch) opts.search = debouncedSearch
      if (statusFilter.length > 0) opts.status = statusFilter.join(',')
      const res = await api.getOrders(opts)
      if (!res.success) throw new Error(res.message || to('fetchFailed'))
      return res.data as Order[]
    },
    { keepPreviousData: true }
  )

  // Fetch selected order details (includes order items)
  const { data: selectedOrderDetails } = useQuery(
    ['order', selectedOrderId],
    async () => {
      if (!selectedOrderId) return null
      const res = await api.getOrder(selectedOrderId)
      if (!res.success) throw new Error(res.message || to('fetchDetailsFailed'))
      return res.data as Order
    },
    { enabled: !!selectedOrderId }
  )

  // Fetch products for add item form
  const { data: productsRes } = useQuery(
    ['products'],
    async () => {
      const res = await api.getProducts()
      if (!res.success) throw new Error(res.message || 'Failed to fetch products')
      return res.data as Product[]
    },
    { enabled: isAddItemFormOpen }
  )

  const createMutation = useMutation(
    async (payload: { customer_id: number; notes?: string }) => {
      const res = await api.createOrder(payload)
      if (!res.success) throw new Error(res.message || to('createFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['orders'])
        closeCreateForm()
      },
      onError: (error) => {
        if (
          error instanceof ApiError &&
          error.status === 409 &&
          (error.data as { code?: string })?.code === 'duplicate_customer_order'
        ) {
          setCreateFormConflict(true)
        }
      },
    }
  )

  const deleteMutation = useMutation(
    async (id: number) => {
      const res = await api.deleteOrder(id)
      if (!res.success) throw new Error(res.message || to('deleteFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['orders'])
        setSelectedOrderId(null)
      },
    }
  )

  const updateStatusMutation = useMutation(
    async ({ id, status }: { id: number; status: string }) => {
      const res = await api.updateOrder(id, { status })
      if (!res.success) throw new Error(res.message || to('updateStatusFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['orders'])
        queryClient.invalidateQueries(['order', selectedOrderId])
      },
    }
  )

  const updateNotesMutation = useMutation(
    async ({ id, notes }: { id: number; notes: string }) => {
      const res = await api.updateOrder(id, { notes })
      if (!res.success) throw new Error(res.message || to('updateNotesFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['orders'])
        queryClient.invalidateQueries(['order', selectedOrderId])
      },
    }
  )

  const addItemMutation = useMutation(
    async ({ orderId, payload, itemPrice }: { orderId: number; payload: { product_id: number; qty: number }; itemPrice: number }) => {
      const res = await api.createOrderItem(orderId, payload)
      if (!res.success) throw new Error(res.message || to('addItemFailed'))

      // Update order total price
      const currentTotal = selectedOrder?.total_price || 0
      const newTotal = currentTotal + (itemPrice * payload.qty)
      const updateRes = await api.updateOrder(orderId, { total_price: newTotal })
      if (!updateRes.success) throw new Error(updateRes.message || to('updateTotalFailed'))

      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
        closeAddItemForm()
      },
    }
  )

  const deleteItemMutation = useMutation(
    async ({ orderId, itemId, itemPrice, itemQty }: { orderId: number; itemId: number; itemPrice: number; itemQty: number }) => {
      const res = await api.deleteOrderItem(orderId, itemId)
      if (!res.success) throw new Error(res.message || to('deleteItemFailed'))

      // Update order total price
      const currentTotal = selectedOrder?.total_price || 0
      const newTotal = currentTotal - (itemPrice * itemQty)
      const updateRes = await api.updateOrder(orderId, { total_price: Math.max(0, newTotal) })
      if (!updateRes.success) throw new Error(updateRes.message || to('updateTotalFailed'))

      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
      },
    }
  )

  const updateItemMutation = useMutation(
    async ({ orderId, itemId, newQty, itemPrice, oldQty }: { orderId: number; itemId: number; newQty: number; itemPrice: number; oldQty: number }) => {
      const res = await api.updateOrderItem(orderId, itemId, { qty: newQty })
      if (!res.success) throw new Error(res.message || to('updateItemFailed'))

      // Update order total price
      const currentTotal = selectedOrder?.total_price || 0
      const priceDiff = itemPrice * (newQty - oldQty)
      const newTotal = currentTotal + priceDiff
      const updateRes = await api.updateOrder(orderId, { total_price: Math.max(0, newTotal) })
      if (!updateRes.success) throw new Error(updateRes.message || to('updateTotalFailed'))

      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
      },
    }
  )

  // Set default selection when data loads
  useEffect(() => {
    if (!selectedOrderId && ordersRes && ordersRes.length > 0) {
      setSelectedOrderId(ordersRes[0].id)
    }
  }, [ordersRes, selectedOrderId])

  const selectedOrder: Order | null = useMemo(() => {
    if (!selectedOrderDetails) {
      // Fallback to basic order data from list
      if (!ordersRes) return null
      return ordersRes.find((o) => o.id === selectedOrderId) || null
    }
    return selectedOrderDetails
  }, [selectedOrderDetails, ordersRes, selectedOrderId])

  function openCreateForm() {
    setCreateForm(emptyCreateForm)
    setIsCreateFormOpen(true)
  }

  function closeCreateForm() {
    setIsCreateFormOpen(false)
    setCreateForm(emptyCreateForm)
    setCreateFormConflict(false)
  }

  function openAddItemForm() {
    setAddItemForm(emptyAddItemForm)
    setIsAddItemFormOpen(true)
  }

  function closeAddItemForm() {
    setIsAddItemFormOpen(false)
    setAddItemForm(emptyAddItemForm)
  }

  function submitCreateForm(e: React.FormEvent) {
    e.preventDefault()
    if (createForm.customer_id) {
      createMutation.mutate({
        customer_id: createForm.customer_id,
        notes: createForm.notes.trim() || undefined
      })
    }
  }

  function submitAddItemForm(e: React.FormEvent) {
    e.preventDefault()
    if (selectedOrderId && addItemForm.product_id && addItemForm.qty > 0) {
      // Find the product price
      const product = productsRes?.find((p) => p.id === addItemForm.product_id)
      const itemPrice = product?.price || 0

      addItemMutation.mutate({
        orderId: selectedOrderId,
        payload: { product_id: addItemForm.product_id, qty: addItemForm.qty },
        itemPrice
      })
    }
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  function formatPrice(price: number) {
    return price.toLocaleString()
  }

  function getStatusStyle(status: string) {
    return statusColors[status] || { bg: '#F5F5F5', color: '#616161' }
  }

  return (
    <Layout>
      <Container disableGutters maxWidth={false} sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden', display: 'flex' }}>
          {isLoading && <Box>{t('loading')}</Box>}
          {isError && (
            <Box sx={{ color: 'error.main' }}>{(error as Error)?.message || tErrors('loadingError', { resource: to('title') })}</Box>
          )}

          {!isLoading && !isError && (
            <Box sx={{ overflow: 'hidden', bgcolor: 'transparent', flex: 1, minHeight: 0, display: 'flex' }}>
              {/* Left list */}
              <Box sx={{ width: { xs: '100%', sm: '300px' }, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRight: { xs: 'none', sm: '1px solid' }, borderColor: 'grey.200' }}>
                <Box sx={{ p: '24px', flexShrink: 0 }}>
                  <Box sx={{ gap: '8px', alignItems: 'center', display: 'flex' }}>
                    <SearchInput
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      placeholder={to('searchPlaceholder')}
                    />
                    <AddButton onClick={openCreateForm} title={to('addOrder')} />
                  </Box>
                  <Box sx={{ mt: '16px' }}>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(Array.from(e.target.selectedOptions).map((o) => o.value))}
                      style={{
                        width: '100px',
                        padding: '6px',
                        fontSize: 12,
                        borderRadius: 6,
                        border: '1px solid var(--theme-ui-colors-border, #e0e0e0)',
                        backgroundColor: 'white',
                        color: 'var(--theme-ui-colors-text, #333)',
                        cursor: 'pointer',
                      }}
                    >
                      <option value="all">{toStatus('active')}</option>
                      <option value="created">{toStatus('created')}</option>
                      <option value="in_progress">{toStatus('in_progress')}</option>
                      <option value="in_delivery">{toStatus('in_delivery')}</option>
                      <option value="done">{toStatus('done')}</option>
                      <option value="cancelled">{toStatus('cancelled')}</option>
                    </select>
                  </Box>
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(ordersRes || []).map((o) => {
                    const isActive = o.id === selectedOrderId
                    const statusStyle = getStatusStyle(o.status)
                    return (
                      <Box
                        key={o.id}
                        sx={{
                          py: '16px',
                          px: '24px',
                          cursor: 'pointer',
                          textAlign: 'left',
                          bgcolor: isActive ? 'grey.100' : 'transparent',
                          borderRadius: '8px',
                          '&:hover': { bgcolor: isActive ? 'grey.100' : 'grey.50' },
                        }}
                        onClick={() => setSelectedOrderId(o.id)}
                      >
                        <Box sx={{ flexDirection: 'column', gap: '4px', display: 'flex' }}>
                          <Box sx={{ justifyContent: 'space-between', alignItems: 'center', display: 'flex' }}>
                            <Box sx={{ fontWeight: 700, fontSize: '14px' }}>#{o.id}</Box>
                            <Box
                              sx={{
                                px: '8px',
                                py: '2px',
                                borderRadius: '4px',
                                fontSize: '12px',
                                fontWeight: 500,
                                bgcolor: statusStyle.bg,
                                color: statusStyle.color,
                                textTransform: 'capitalize',
                              }}
                            >
                              {toStatus(o.status)}
                            </Box>
                          </Box>
                          <Box sx={{ fontSize: '12px', color: 'grey.500' }}>{o.customer_name}</Box>
                          <Box sx={{ fontSize: '12px', fontWeight: 500 }}>{formatPrice(o.total_price)}</Box>
                        </Box>
                      </Box>
                    )
                  })}
                  {(ordersRes || []).length === 0 && (
                    <Box sx={{ p: '16px', color: 'grey.500', textAlign: 'center' }}>{to('noOrders')}</Box>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bgcolor: 'grey.50' }}>
                {selectedOrder ? (
                  <Box sx={{ maxWidth: 880, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
                    <Box sx={{ alignItems: 'center', justifyContent: 'space-between', mb: '16px', display: 'flex' }}>
                      <Box sx={{ alignItems: 'center', gap: '16px', display: 'flex' }}>
                        <Typography component="h2" sx={{ fontSize: '18px' }}>{to('orderNumber', { id: selectedOrder.id })}</Typography>
                        <Box
                          sx={{
                            px: '8px',
                            py: '4px',
                            borderRadius: '4px',
                            fontSize: '12px',
                            fontWeight: 500,
                            bgcolor: getStatusStyle(selectedOrder.status).bg,
                            color: getStatusStyle(selectedOrder.status).color,
                            textTransform: 'capitalize',
                          }}
                        >
                          {toStatus(selectedOrder.status)}
                        </Box>
                      </Box>
                      <Button
                        variant="outlined"
                        onClick={() => {
                          if (confirm(to('deleteConfirm'))) deleteMutation.mutate(selectedOrder.id)
                        }}
                      >
                        {t('delete')}
                      </Button>
                    </Box>

                    {/* Order info card */}
                    <Paper
                      sx={{
                        p: '24px',
                        mb: '24px',
                        borderRadius: '12px',
                        boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'white',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' },
                      }}
                    >
                      <Box sx={{ flexWrap: 'wrap', gap: { xs: '24px', sm: '32px' }, display: 'flex' }}>
                        <Box sx={{ minWidth: 140 }}>
                          <Box sx={{ color: 'grey.500', fontSize: '14px', fontWeight: 700, mb: '4px', display: 'block' }}>{t('customer')}</Box>
                          <Box sx={{ fontSize: '14px', fontWeight: 500 }}>{selectedOrder.customer_name}</Box>
                        </Box>
                        <Box sx={{ minWidth: 140 }}>
                          <Box sx={{ color: 'grey.500', fontSize: '14px', fontWeight: 700, mb: '4px', display: 'block' }}>{to('created')}</Box>
                          <Box sx={{ fontSize: '14px' }}>{formatDate(selectedOrder.created_at)}</Box>
                        </Box>
                        <Box sx={{ minWidth: 140, ml: 'auto' }}>
                          <Box sx={{ color: 'grey.500', fontSize: '14px', fontWeight: 700, mb: '4px', display: 'block' }}>{t('status')}</Box>
                          <NativeSelect
                            value={selectedOrder.status}
                            onChange={(e) => updateStatusMutation.mutate({ id: selectedOrder.id, status: e.target.value })}
                            disableUnderline
                            sx={{
                              py: '4px',
                              px: '8px',
                              fontSize: '14px',
                              borderRadius: '8px',
                              border: '1px solid',
                              borderColor: 'grey.200',
                              fontWeight: 500,
                              cursor: 'pointer',
                            }}
                          >
                            <option value="created">{toStatus('created')}</option>
                            <option value="in_progress">{toStatus('in_progress')}</option>
                            <option value="in_delivery">{toStatus('in_delivery')}</option>
                            <option value="done">{toStatus('done')}</option>
                            <option value="cancelled">{toStatus('cancelled')}</option>
                          </NativeSelect>
                        </Box>
                        <Box sx={{ flex: '1 1 100%', minWidth: 0 }}>
                          <Box sx={{ color: 'grey.500', fontSize: '14px', fontWeight: 700, mb: '4px', display: 'block' }}>{to('notes')}</Box>
                          <OutlinedInput
                            key={selectedOrder.id}
                            defaultValue={selectedOrder.notes || ''}
                            onBlur={(e) => {
                              const notes = e.target.value
                              if (notes !== (selectedOrder.notes || '')) {
                                updateNotesMutation.mutate({ id: selectedOrder.id, notes })
                              }
                            }}
                            multiline
                            rows={2}
                            size="small"
                            sx={{
                              width: '100%',
                              py: '8px',
                              px: '16px',
                              fontSize: '14px',
                              borderRadius: '8px',
                              border: '1px solid',
                              borderColor: 'grey.200',
                              resize: 'vertical',
                              minHeight: '60px',
                            }}
                          />
                        </Box>
                      </Box>
                    </Paper>

                    {/* Order items */}
                    <Paper
                      sx={{
                        borderRadius: '12px',
                        boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'white',
                        overflow: 'hidden',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' },
                      }}
                    >
                      <Box sx={{ alignItems: 'center', justifyContent: 'space-between', p: '8px', borderBottom: '1px solid', borderColor: 'grey.200', bgcolor: 'grey.50', display: 'flex' }}>
                        <Typography component="h3" sx={{ fontSize: '16px', fontWeight: 600 }}>{t('items')}</Typography>
                        <Button variant="outlined" onClick={openAddItemForm} sx={{ fontSize: '12px', py: '4px', px: '8px' }}>
                          {to('addItem')}
                        </Button>
                      </Box>
                      {selectedOrder.order_items && selectedOrder.order_items.length > 0 ? (
                        <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                          <Box component="thead">
                            <Box component="tr" sx={{ bgcolor: 'grey.50' }}>
                              <Box component="th" sx={{ p: '16px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'grey.500', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('product')}</Box>
                              <Box component="th" sx={{ p: '16px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'grey.500', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('price')}</Box>
                              <Box component="th" sx={{ p: '16px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'grey.500', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('quantity')}</Box>
                              <Box component="th" sx={{ p: '16px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'grey.500', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{to('subtotal')}</Box>
                              <Box component="th" sx={{ p: '16px', width: '56px' }}></Box>
                            </Box>
                          </Box>
                          <Box component="tbody">
                            {selectedOrder.order_items.map((item) => (
                              <Box component="tr" key={item.id} sx={{ borderTop: '1px solid', borderColor: 'grey.200', '&:hover': { bgcolor: 'grey.50' } }}>
                                <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px' }}>{item.product_name}</Box>
                                <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontSize: '14px' }}>{formatPrice(item.price)}</Box>
                                <Box component="td" sx={{ py: '8px', pl: '16px', pr: '8px' }}>
                                  <Box sx={{ justifyContent: 'flex-end', display: 'flex' }}>
                                  <OutlinedInput
                                    type="number"
                                    inputProps={{ min: '1' }}
                                    defaultValue={item.qty}
                                    size="small"
                                    sx={{ width: '60px', textAlign: 'right', fontSize: '14px', borderRadius: '8px' }}
                                    onBlur={(e) => {
                                      const newQty = Number(e.target.value) || 1
                                      if (newQty !== item.qty && newQty > 0) {
                                        updateItemMutation.mutate({
                                          orderId: selectedOrder.id,
                                          itemId: item.id,
                                          newQty,
                                          itemPrice: item.price,
                                          oldQty: item.qty
                                        })
                                      }
                                    }}
                                    onKeyDown={(e) => {
                                      if (e.key === 'Enter') {
                                        (e.target as HTMLInputElement).blur()
                                      }
                                    }}
                                  />
                                  </Box>
                                </Box>
                                <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontSize: '14px', fontWeight: 500 }}>{formatPrice(item.price * item.qty)}</Box>
                                <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'center' }}>
                                  <Box
                                    component="span"
                                    sx={{ color: 'error.main', cursor: 'pointer', display: 'inline-flex', '&:hover': { color: 'error.dark' } }}
                                    onClick={() => {
                                      if (confirm(to('removeItemConfirm'))) {
                                        deleteItemMutation.mutate({ orderId: selectedOrder.id, itemId: item.id, itemPrice: item.price, itemQty: item.qty })
                                      }
                                    }}
                                  >
                                    <Trash2 size={20} />
                                  </Box>
                                </Box>
                              </Box>
                            ))}
                          </Box>
                          <Box component="tfoot">
                            <Box component="tr" sx={{ borderTop: '2px solid', borderColor: 'grey.200', bgcolor: 'grey.50' }}>
                              <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontWeight: 700, fontSize: '16px' }} {...({ colSpan: 3 } as object)}>{t('total')}</Box>
                              <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontWeight: 700, fontSize: '16px', color: 'primary.main' }}>{formatPrice(selectedOrder.total_price)}</Box>
                              <Box component="td" sx={{ py: '8px', px: '16px' }}></Box>
                            </Box>
                          </Box>
                        </Box>
                      ) : (
                        <Box sx={{ p: '32px', textAlign: 'center', color: 'grey.500' }}>
                          <Box sx={{ fontSize: '16px', display: 'block', mb: '8px' }}>{to('noItems')}</Box>
                          <Button variant="outlined" onClick={openAddItemForm}>
                            {to('addFirstItem')}
                          </Button>
                        </Box>
                      )}
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
                    <ClipboardList size={48} opacity={0.4} />
                    <Box sx={{ fontSize: '16px' }}>{to('selectOrder')}</Box>
                    <Box sx={{ fontSize: '14px' }}>{to('chooseFromList')}</Box>
                  </Box>
                )}
              </Box>
            </Box>
          )}
        </Box>

        {/* Create Order Modal */}
        <Dialog open={isCreateFormOpen} onClose={closeCreateForm} fullWidth maxWidth="sm">
          {createFormConflict ? (
            <>
              <DialogTitle sx={{ pb: '8px' }}>{to('duplicateOrderTitle')}</DialogTitle>
              <DialogContent sx={{ pt: '8px', pb: 0 }}>
                <Box sx={{ mb: '8px', color: 'grey.500' }}>
                  {to('duplicateOrderMessageInline')}
                </Box>
              </DialogContent>
              <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
                <Button type="button" variant="outlined" onClick={() => setCreateFormConflict(false)}>
                  {to('chooseDifferentCustomer')}
                </Button>
                <Button type="button" variant="contained" disableElevation onClick={closeCreateForm}>
                  {t('cancel')}
                </Button>
              </DialogActions>
            </>
          ) : (
            <Box component="form" onSubmit={submitCreateForm}>
              <DialogTitle sx={{ pb: '8px' }}>{to('newOrder')}</DialogTitle>
              <DialogContent sx={{ pt: '8px', pb: 0 }}>
                <Box sx={{ mb: '16px' }}>
                  <Box component="label" htmlFor="customer" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('customer')}</Box>
                  <CustomerSearchSelect
                    value={createForm.customer_id}
                    onChange={(id) => setCreateForm({ ...createForm, customer_id: id })}
                    placeholder={to('selectCustomer')}
                    searchPlaceholder={to('searchCustomer')}
                  />
                </Box>
                <Box sx={{ mb: '8px' }}>
                  <Box component="label" htmlFor="notes" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{to('notes')}</Box>
                  <OutlinedInput
                    id="notes"
                    value={createForm.notes}
                    onChange={(e) => setCreateForm({ ...createForm, notes: e.target.value })}
                    multiline
                    rows={3}
                    size="small"
                    sx={{
                      width: '100%',
                      py: '8px',
                      px: '16px',
                      fontSize: '14px',
                      borderRadius: '8px',
                      border: '1px solid',
                      borderColor: 'grey.200',
                      resize: 'vertical',
                    }}
                  />
                </Box>
              </DialogContent>
              <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
                <Button type="button" variant="outlined" onClick={closeCreateForm}>
                  {t('cancel')}
                </Button>
                <Button type="submit" variant="contained" disableElevation disabled={createMutation.isLoading || !createForm.customer_id}>
                  {to('createOrder')}
                </Button>
              </DialogActions>
            </Box>
          )}
        </Dialog>

        {/* Add Item Modal */}
        <Dialog open={isAddItemFormOpen} onClose={closeAddItemForm} fullWidth maxWidth="xs">
          <Box component="form" onSubmit={submitAddItemForm}>
            <DialogTitle sx={{ pb: '8px' }}>{to('addItemTitle')}</DialogTitle>
            <DialogContent sx={{ pt: '8px', pb: 0 }}>
              <Box sx={{ mb: '16px' }}>
                <Box component="label" htmlFor="product" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('product')}</Box>
                <NativeSelect
                  id="product"
                  value={addItemForm.product_id || ''}
                  onChange={(e) => setAddItemForm({ ...addItemForm, product_id: Number(e.target.value) || null })}
                  disableUnderline
                  required
                  sx={{ width: '100%' }}
                >
                  <option value="">{to('selectProduct')}</option>
                  {(productsRes || []).map((p) => (
                    <option key={p.id} value={p.id}>{p.name}</option>
                  ))}
                </NativeSelect>
              </Box>
              <Box sx={{ mb: '8px' }}>
                <Box component="label" htmlFor="qty" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('quantity')}</Box>
                <OutlinedInput
                  id="qty"
                  type="number"
                  inputProps={{ min: '1' }}
                  value={addItemForm.qty}
                  onChange={(e) => setAddItemForm({ ...addItemForm, qty: Number(e.target.value) || 1 })}
                  size="small"
                  sx={{ width: '60px', borderRadius: '8px' }}
                  required
                />
              </Box>
            </DialogContent>
            <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
              <Button type="button" variant="outlined" onClick={closeAddItemForm}>
                {t('cancel')}
              </Button>
              <Button type="submit" variant="contained" disableElevation disabled={addItemMutation.isLoading || !addItemForm.product_id}>
                {to('addItemButton')}
              </Button>
            </DialogActions>
          </Box>
        </Dialog>

      </Container>
    </Layout>
  )
}
