"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, Dialog, DialogActions, DialogContent, DialogTitle, Drawer, IconButton, InputBase, ListSubheader, Menu, MenuItem, OutlinedInput, Paper, Select, Tooltip, Typography, useTheme } from '@mui/material'
import DateRangeFilter from '@/components/ui/DateRangeFilter'
import AddButton from '@/components/ui/AddButton'
import CustomerSearchSelect from '@/components/ui/CustomerSearchSelect'
import ProductSearchSelect from '@/components/ui/ProductSearchSelect'
import { api, ApiError } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { ArrowUpDown, Download, Info, ListFilter, Pencil, Search, Trash2, X } from 'lucide-react'

type OrderItem = {
  id: number
  order_id?: number
  product_name: string
  price: number
  qty: number
  created_at: string
  updated_at?: string | null
}

type OrderPayment = {
  id: number
  order_id?: number
  amount: number
  created_at: string
  updated_at?: string | null
}

type Order = {
  id: number
  customer_name: string
  is_customer_deleted?: boolean
  total_price: number
  status: string
  payment_status: string
  notes?: string
  order_items?: OrderItem[]
  order_payments?: OrderPayment[]
  created_at: string
  updated_at?: string | null
}

type Customer = {
  id: number
  name: string
  phone: string
  address: string
}


type CreateOrderForm = {
  customer_id: number | null
  customer_name?: string
  notes: string
}

type AddItemForm = {
  product_id: number | null
  product_name: string
  product_price: number
  qty: number
}

const emptyCreateForm: CreateOrderForm = { customer_id: null, notes: '' }
const emptyAddItemForm: AddItemForm = { product_id: null, product_name: '', product_price: 0, qty: 1 }

const lightStatusColors: Record<string, { bg: string; color: string }> = {
  created:     { bg: '#E3F2FD', color: '#1565C0' },
  in_progress: { bg: '#FFF3E0', color: '#E65100' },
  in_delivery: { bg: '#F3E5F5', color: '#7B1FA2' },
  done:        { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled:   { bg: '#FFEBEE', color: '#C62828' },
}

const darkStatusColors: Record<string, { bg: string; color: string }> = {
  created:     { bg: '#1e3a5f', color: '#90caf9' },
  in_progress: { bg: '#3e2000', color: '#ffb74d' },
  in_delivery: { bg: '#2d1b4e', color: '#ce93d8' },
  done:        { bg: '#1b3a2d', color: '#81c784' },
  cancelled:   { bg: '#3e1a1a', color: '#ef9a9a' },
}

const lightPaymentStatusColors: Record<string, { bg: string; color: string }> = {
  outstanding: { bg: '#FFEBEE', color: '#C62828' },
  paid:        { bg: '#E8F5E9', color: '#2E7D32' },
}

const darkPaymentStatusColors: Record<string, { bg: string; color: string }> = {
  outstanding: { bg: '#3e1a1a', color: '#ef9a9a' },
  paid:        { bg: '#1b3a2d', color: '#81c784' },
}

function normalizePaymentStatus(status: string): string {
  if (status === 'unpaid' || status === 'partial') return 'outstanding'
  return status
}

const DEFAULT_STATUSES = ['created', 'in_progress', 'in_delivery']
const VALID_SORT_VALUES = ['date_desc', 'date_asc', 'total_asc', 'total_desc'] as const
type SortBy = typeof VALID_SORT_VALUES[number]

const FILTER_SESSION_KEY = 'orders_filter_state'

function getStoredFilterState() {
  if (typeof window === 'undefined') return null
  try {
    const s = sessionStorage.getItem(FILTER_SESSION_KEY)
    return s ? JSON.parse(s) : null
  } catch {
    return null
  }
}

export default function OrdersPage() {
  const theme = useTheme()
  const statusColors = theme.palette.mode === 'dark' ? darkStatusColors : lightStatusColors
  const paymentStatusColors = theme.palette.mode === 'dark' ? darkPaymentStatusColors : lightPaymentStatusColors
  const t = useTranslations('common')
  const to = useTranslations('orders')
  const toStatus = useTranslations('orderStatus')
  const toPaymentStatus = useTranslations('paymentStatus')
  const tErrors = useTranslations('errors')
  const queryClient = useQueryClient()
  const [isCreateFormOpen, setIsCreateFormOpen] = useState(false)
  const [isAddItemFormOpen, setIsAddItemFormOpen] = useState(false)
  const [isAddPaymentFormOpen, setIsAddPaymentFormOpen] = useState(false)
  const [addPaymentAmount, setAddPaymentAmount] = useState<number>(0)
  const [editingPayment, setEditingPayment] = useState<{ id: number; amount: number } | null>(null)
  const [createForm, setCreateForm] = useState<CreateOrderForm>(emptyCreateForm)
  const [addItemForm, setAddItemForm] = useState<AddItemForm>(emptyAddItemForm)
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null)
  const [searchInput, setSearchInput] = useState<string>(() => getStoredFilterState()?.searchInput ?? '')
  const [debouncedSearch, setDebouncedSearch] = useState<string>(() => getStoredFilterState()?.searchInput ?? '')
  const [statusFilter, setStatusFilter] = useState<string[]>(() => getStoredFilterState()?.statusFilter ?? [...DEFAULT_STATUSES])
  const [paymentStatusFilter, setPaymentStatusFilter] = useState<string>(() => getStoredFilterState()?.paymentStatusFilter ?? '')
  const [dateFrom, setDateFrom] = useState<string>(() => getStoredFilterState()?.dateFrom ?? '')
  const [dateTo, setDateTo] = useState<string>(() => getStoredFilterState()?.dateTo ?? '')
  const [createFormConflict, setCreateFormConflict] = useState(false)
  const [isExportDialogOpen, setIsExportDialogOpen] = useState(false)
  const [exportMessage, setExportMessage] = useState('')
  const [sortBy, setSortBy] = useState<SortBy>(() => getStoredFilterState()?.sortBy ?? 'date_desc')
  const [createCustomerDialog, setCreateCustomerDialog] = useState<{ open: boolean; name: string; phone: string }>({ open: false, name: '', phone: '' })
  const [createProductDialog, setCreateProductDialog] = useState<{ open: boolean; name: string; price: string }>({ open: false, name: '', price: '' })
  const [filtersVisible, setFiltersVisible] = useState<boolean>(() => getStoredFilterState()?.filtersVisible ?? false)
  const [sortMenuAnchor, setSortMenuAnchor] = useState<null | HTMLElement>(null)
  const [markUnpaidDialogOpen, setMarkUnpaidDialogOpen] = useState(false)

  useEffect(() => {
    sessionStorage.setItem(FILTER_SESSION_KEY, JSON.stringify({
      searchInput,
      statusFilter,
      paymentStatusFilter,
      dateFrom,
      dateTo,
      sortBy,
      filtersVisible,
    }))
  }, [searchInput, statusFilter, paymentStatusFilter, dateFrom, dateTo, sortBy, filtersVisible])

  const handleResetFilters = () => {
    setSearchInput('')
    setDebouncedSearch('')
    setStatusFilter([...DEFAULT_STATUSES])
    setPaymentStatusFilter('')
    setDateFrom('')
    setDateTo('')
  }

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const { data: ordersRes, isLoading, isError, error } = useQuery(
    ['orders', debouncedSearch, statusFilter, paymentStatusFilter, dateFrom, dateTo],
    async () => {
      const opts: { search?: string; status?: string; payment_status?: string; date_from?: string; date_to?: string } = {}
      if (debouncedSearch) opts.search = debouncedSearch
      if (statusFilter.length > 0) opts.status = statusFilter.join(',')
      if (paymentStatusFilter) opts.payment_status = paymentStatusFilter
      if (dateFrom) opts.date_from = dateFrom
      if (dateTo) opts.date_to = dateTo
      const res = await api.getOrders(opts)
      if (!res.success) throw new Error(res.message || to('fetchFailed'))
      return res.data as Order[]
    },
    { keepPreviousData: true }
  )

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

  const updatePaymentStatusMutation = useMutation(
    async ({ id, payment_status }: { id: number; payment_status: string }) => {
      const res = await api.updateOrder(id, { payment_status })
      if (!res.success) throw new Error(res.message || to('updatePaymentStatusFailed'))
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

  function derivePaymentStatus(totalPaid: number, totalPrice: number): string {
    if (totalPaid >= totalPrice && totalPaid > 0) return 'paid'
    return 'outstanding'
  }

  const addPaymentMutation = useMutation(
    async ({ orderId, amount }: { orderId: number; amount: number }) => {
      const res = await api.createOrderPayment(orderId, { amount })
      if (!res.success) throw new Error(res.message || to('addPaymentFailed'))

      const currentPaid = (selectedOrder?.order_payments || []).reduce((sum, p) => sum + p.amount, 0)
      const newTotalPaid = currentPaid + amount
      const newStatus = derivePaymentStatus(newTotalPaid, selectedOrder?.total_price || 0)
      await api.updateOrder(orderId, { payment_status: newStatus })

      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
        setIsAddPaymentFormOpen(false)
        setAddPaymentAmount(0)
      },
    }
  )

  const updatePaymentMutation = useMutation(
    async ({ orderId, paymentId, amount }: { orderId: number; paymentId: number; amount: number }) => {
      const res = await api.updateOrderPayment(orderId, paymentId, { amount })
      if (!res.success) throw new Error(res.message || to('updatePaymentFailed'))

      const currentPayments = selectedOrder?.order_payments || []
      const oldAmount = currentPayments.find((p) => p.id === paymentId)?.amount || 0
      const currentPaid = currentPayments.reduce((sum, p) => sum + p.amount, 0)
      const newTotalPaid = currentPaid - oldAmount + amount
      const newStatus = derivePaymentStatus(newTotalPaid, selectedOrder?.total_price || 0)
      await api.updateOrder(orderId, { payment_status: newStatus })

      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
      },
    }
  )

  const deletePaymentMutation = useMutation(
    async ({ orderId, paymentId }: { orderId: number; paymentId: number }) => {
      const res = await api.deleteOrderPayment(orderId, paymentId)
      if (!res.success) throw new Error(res.message || to('deletePaymentFailed'))

      const currentPayments = selectedOrder?.order_payments || []
      const deletedAmount = currentPayments.find((p) => p.id === paymentId)?.amount || 0
      const newTotalPaid = currentPayments.reduce((sum, p) => sum + p.amount, 0) - deletedAmount
      const newStatus = derivePaymentStatus(newTotalPaid, selectedOrder?.total_price || 0)
      await api.updateOrder(orderId, { payment_status: newStatus })

      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
      },
    }
  )

  const markUnpaidMutation = useMutation(
    async (orderId: number) => {
      await api.deleteOrderPayments(orderId)
      const res = await api.updateOrder(orderId, { payment_status: 'outstanding' })
      if (!res.success) throw new Error(res.message || to('updatePaymentStatusFailed'))
      return res
    },
    {
      onSuccess: () => {
        setMarkUnpaidDialogOpen(false)
        queryClient.invalidateQueries(['order', selectedOrderId])
        queryClient.invalidateQueries(['orders'])
      },
    }
  )

  const exportMutation = useMutation(
    async ({ orderId, message }: { orderId: number; message: string }) => {
      const blob = await api.exportOrderInvoice(orderId, message)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `invoice-${orderId}.pdf`
      a.click()
      URL.revokeObjectURL(url)
    },
    {
      onSuccess: () => {
        setIsExportDialogOpen(false)
        setExportMessage('')
      },
    }
  )

  const createCustomerMutation = useMutation(
    (data: { name: string; phone: string }) => api.createCustomer({ name: data.name, phone: data.phone, address: '' }),
    {
      onSuccess: (res) => {
        const customer = res.data as Customer
        setCreateForm((f) => ({ ...f, customer_id: customer.id, customer_name: customer.name }))
        setCreateCustomerDialog({ open: false, name: '', phone: '' })
        queryClient.invalidateQueries(['customers'])
      },
    }
  )

  const createProductMutation = useMutation(
    (data: { name: string; price: number }) => api.createProduct({ name: data.name, price: data.price }),
    {
      onSuccess: (res) => {
        const product = res.data as { id: number; name: string; price: number }
        setAddItemForm((f) => ({ ...f, product_id: product.id, product_name: product.name, product_price: product.price }))
        setCreateProductDialog({ open: false, name: '', price: '' })
        queryClient.invalidateQueries(['products'])
      },
    }
  )

  const sortedOrders = useMemo(() => {
    const list = ordersRes || []
    return [...list].sort((a, b) => {
      switch (sortBy) {
        case 'date_asc': return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
        case 'total_asc': return a.total_price - b.total_price
        case 'total_desc': return b.total_price - a.total_price
        case 'date_desc':
        default: return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      }
    })
  }, [ordersRes, sortBy])

  const selectedOrder: Order | null = useMemo(() => {
    if (!selectedOrderDetails) {
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
      addItemMutation.mutate({
        orderId: selectedOrderId,
        payload: { product_id: addItemForm.product_id, qty: addItemForm.qty },
        itemPrice: addItemForm.product_price,
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
    return statusColors[status] || (theme.palette.mode === 'dark' ? { bg: '#2a2a2a', color: '#bdbdbd' } : { bg: '#F5F5F5', color: '#616161' })
  }

  function getPaymentStatusStyle(status: string) {
    return paymentStatusColors[status] || (theme.palette.mode === 'dark' ? { bg: '#2a2a2a', color: '#bdbdbd' } : { bg: '#F5F5F5', color: '#616161' })
  }

  const drawerOpen = selectedOrderId !== null

  return (
    <Container disableGutters maxWidth={false} sx={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {isLoading && <PageLoadingSkeleton />}
        {isError && (
          <Box sx={{ p: '24px', color: 'error.main' }}>{(error as Error)?.message || tErrors('loadingError', { resource: to('title') })}</Box>
        )}

        {!isLoading && !isError && (
          <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            {/* Toolbar */}
            <Box sx={{ px: '24px', pt: '24px', pb: '16px', flexShrink: 0 }}>
              <Box sx={{ maxWidth: 1200, mx: 'auto' }}>
              {/* Action bar */}
              <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <Paper
                variant="outlined"
                sx={{ flex: 1, display: 'flex', alignItems: 'center', px: '8px', py: '4px', borderRadius: '10px', boxShadow: 'none', gap: '2px' }}
              >
                <Tooltip title={to('sortLabel')}>
                  <IconButton
                    size="small"
                    onClick={(e) => setSortMenuAnchor(e.currentTarget)}
                    sx={{ color: sortBy !== 'date_desc' ? 'primary.main' : 'text.secondary', '&:hover': { color: sortBy !== 'date_desc' ? 'primary.dark' : 'text.primary' } }}
                  >
                    <ArrowUpDown size={18} />
                  </IconButton>
                </Tooltip>
                <Tooltip title="Filters">
                  <IconButton
                    size="small"
                    onClick={() => setFiltersVisible((v) => !v)}
                    sx={{ color: filtersVisible ? 'primary.main' : 'text.secondary', '&:hover': { color: filtersVisible ? 'primary.dark' : 'text.primary' } }}
                  >
                    <ListFilter size={18} />
                  </IconButton>
                </Tooltip>

                <InputBase
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  placeholder={to('searchPlaceholder')}
                  sx={{ flex: 1, mx: '8px', fontSize: '13px', color: 'text.primary' }}
                  inputProps={{ 'aria-label': to('searchPlaceholder') }}
                />

                <Search size={18} style={{ color: '#9e9e9e', flexShrink: 0, marginRight: '4px' }} />
              </Paper>
              <AddButton onClick={openCreateForm} title={to('addOrder')} />
              </Box>

              {/* Sort menu */}
              <Menu
                anchorEl={sortMenuAnchor}
                open={!!sortMenuAnchor}
                onClose={() => setSortMenuAnchor(null)}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
                transformOrigin={{ vertical: 'top', horizontal: 'left' }}
              >
                <ListSubheader sx={{ fontSize: '12px', lineHeight: '28px' }}>{to('sortLabel')}</ListSubheader>
                {([
                  ['date_desc', to('sortDateDesc')],
                  ['date_asc', to('sortDateAsc')],
                  ['total_asc', to('sortTotalAsc')],
                  ['total_desc', to('sortTotalDesc')],
                ] as const).map(([val, label]) => (
                  <MenuItem
                    key={val}
                    selected={sortBy === val}
                    onClick={() => { setSortBy(val); setSortMenuAnchor(null) }}
                    sx={{ fontSize: '13px' }}
                  >
                    {label}
                  </MenuItem>
                ))}
              </Menu>

              {/* Filters row */}
              {filtersVisible && (
                <Box sx={{ display: 'flex', gap: '12px', mt: '12px', flexWrap: 'wrap', alignItems: 'flex-end' }}>
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                    <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>{t('status')}</Box>
                    <Select
                      size="small"
                      value={statusFilter.length === 1 ? statusFilter[0] : statusFilter.join(',')}
                      onChange={(e) => {
                        const val = e.target.value as string
                        setStatusFilter(val === '' ? [...DEFAULT_STATUSES] : val.split(','))
                      }}
                      sx={{ height: 34, fontSize: '13px', borderRadius: '6px', minWidth: 130 }}
                      MenuProps={{ anchorOrigin: { vertical: 'bottom', horizontal: 'left' }, transformOrigin: { vertical: 'top', horizontal: 'left' } }}
                    >
                      <MenuItem value="created,in_progress,in_delivery" sx={{ fontSize: '13px' }}>{toStatus('active')}</MenuItem>
                      <MenuItem value="created" sx={{ fontSize: '13px' }}>{toStatus('created')}</MenuItem>
                      <MenuItem value="in_progress" sx={{ fontSize: '13px' }}>{toStatus('in_progress')}</MenuItem>
                      <MenuItem value="in_delivery" sx={{ fontSize: '13px' }}>{toStatus('in_delivery')}</MenuItem>
                      <MenuItem value="done" sx={{ fontSize: '13px' }}>{toStatus('done')}</MenuItem>
                      <MenuItem value="cancelled" sx={{ fontSize: '13px' }}>{toStatus('cancelled')}</MenuItem>
                    </Select>
                  </Box>
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                    <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>{to('paymentStatus')}</Box>
                    <Select
                      size="small"
                      displayEmpty
                      value={paymentStatusFilter}
                      onChange={(e) => setPaymentStatusFilter(e.target.value as string)}
                      sx={{ height: 34, fontSize: '13px', borderRadius: '6px', minWidth: 130 }}
                      MenuProps={{ anchorOrigin: { vertical: 'bottom', horizontal: 'left' }, transformOrigin: { vertical: 'top', horizontal: 'left' } }}
                    >
                      <MenuItem value="" sx={{ fontSize: '13px' }}>{toStatus('all')}</MenuItem>
                      <MenuItem value="outstanding" sx={{ fontSize: '13px' }}>{toPaymentStatus('outstanding')}</MenuItem>
                      <MenuItem value="paid" sx={{ fontSize: '13px' }}>{toPaymentStatus('paid')}</MenuItem>
                    </Select>
                  </Box>
                  <DateRangeFilter
                    dateFrom={dateFrom}
                    dateTo={dateTo}
                    onDateFromChange={setDateFrom}
                    onDateToChange={setDateTo}
                  />
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={handleResetFilters}
                    sx={{ height: 36, fontSize: '13px', borderRadius: '6px', textTransform: 'none', alignSelf: 'flex-end', flexShrink: 0 }}
                  >
                    {to('resetFilters')}
                  </Button>
                </Box>
              )}
              </Box>
            </Box>

            {/* Table */}
            <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', px: '24px', pb: '24px' }}>
              <Box sx={{ maxWidth: 1200, mx: 'auto' }}>
              {sortedOrders.length === 0 ? (
                <Box sx={{ textAlign: 'center', py: '64px', color: 'text.secondary', fontSize: '14px' }}>{to('noOrders')}</Box>
              ) : (
                <Paper sx={{ borderRadius: '12px', border: '1px solid', borderColor: 'divider', boxShadow: 'none', overflow: 'hidden' }}>
                  <Box sx={{ overflowX: 'auto' }}>
                  <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 480 }}>
                    <Box component="thead">
                      <Box component="tr" sx={{ bgcolor: 'action.hover', borderBottom: '1px solid', borderColor: 'divider' }}>
                        <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', width: '90px' }}>
                          {to('orderNumber', { id: '' }).trim()}
                        </Box>
                        <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                          {t('customer')}
                        </Box>
                        <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', display: { xs: 'none', md: 'table-cell' } }}>
                          {t('status')}
                        </Box>
                        <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', display: { xs: 'none', sm: 'table-cell' } }}>
                          {to('paymentStatus')}
                        </Box>
                        <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                          {t('total')}
                        </Box>
                        <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', display: { xs: 'none', lg: 'table-cell' } }}>
                          {to('created')}
                        </Box>
                      </Box>
                    </Box>
                    <Box component="tbody">
                      {sortedOrders.map((o) => {
                        const statusStyle = getStatusStyle(o.status)
                        const payStyle = getPaymentStatusStyle(normalizePaymentStatus(o.payment_status))
                        return (
                          <Box
                            component="tr"
                            key={o.id}
                            onClick={() => setSelectedOrderId(o.id)}
                            sx={{
                              borderTop: '1px solid',
                              borderColor: 'divider',
                              cursor: 'pointer',
                              transition: 'background 0.1s',
                              bgcolor: selectedOrderId === o.id ? 'action.selected' : 'transparent',
                              '&:hover': { bgcolor: 'action.hover' },
                            }}
                          >
                            <Box component="td" sx={{ px: '16px', py: '14px', fontSize: '14px', fontWeight: 600, color: 'text.secondary', whiteSpace: 'nowrap' }}>
                              #{o.id}
                            </Box>
                            <Box component="td" sx={{ px: '16px', py: '14px' }}>
                              <Box sx={{ fontSize: '14px', fontWeight: 500, display: 'flex', alignItems: 'center', gap: '6px' }}>
                                {o.customer_name}
                                {o.is_customer_deleted && (
                                  <Box component="span" sx={{ fontSize: '10px', px: '5px', py: '1px', borderRadius: '3px', bgcolor: 'action.selected', color: 'text.secondary', fontWeight: 500, lineHeight: 1.5 }}>
                                    {to('customerDeleted')}
                                  </Box>
                                )}
                              </Box>
                            </Box>
                            <Box component="td" sx={{ px: '16px', py: '14px', display: { xs: 'none', md: 'table-cell' } }}>
                              <Box sx={{ display: 'inline-block', px: '8px', py: '3px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, bgcolor: statusStyle.bg, color: statusStyle.color }}>
                                {toStatus(o.status)}
                              </Box>
                            </Box>
                            <Box component="td" sx={{ px: '16px', py: '14px', display: { xs: 'none', sm: 'table-cell' } }}>
                              <Box sx={{ display: 'inline-block', px: '8px', py: '3px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, bgcolor: payStyle.bg, color: payStyle.color }}>
                                {toPaymentStatus(normalizePaymentStatus(o.payment_status || 'outstanding'))}
                              </Box>
                            </Box>
                            <Box component="td" sx={{ px: '16px', py: '14px', textAlign: 'right', fontSize: '14px', fontWeight: 600, whiteSpace: 'nowrap' }}>
                              {formatPrice(o.total_price)}
                            </Box>
                            <Box component="td" sx={{ px: '16px', py: '14px', textAlign: 'right', fontSize: '13px', color: 'text.secondary', whiteSpace: 'nowrap', display: { xs: 'none', lg: 'table-cell' } }}>
                              {formatDate(o.created_at)}
                            </Box>
                          </Box>
                        )
                      })}
                    </Box>
                  </Box>
                  </Box>
                </Paper>
              )}
              </Box>
            </Box>
          </Box>
        )}
      </Box>

      {/* Detail Drawer */}
      <Drawer
        anchor="right"
        open={drawerOpen}
        onClose={() => setSelectedOrderId(null)}
        slotProps={{
          paper: {
            sx: {
              width: { xs: '100%', sm: 680 },
              maxWidth: '100%',
            },
          },
        }}
      >
        {selectedOrder && (
          <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            {/* Drawer header */}
            <Box sx={{ px: '24px', py: '16px', borderBottom: '1px solid', borderColor: 'divider', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexShrink: 0 }}>
              <Box>
                <Typography sx={{ fontSize: '18px', fontWeight: 600 }}>{to('orderNumber', { id: selectedOrder.id })}</Typography>
                <Box sx={{ fontSize: '12px', color: 'text.secondary', mt: '2px' }}>{formatDate(selectedOrder.created_at)}</Box>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Tooltip title={to('exportInvoice')}>
                  <IconButton size="small" onClick={() => { setExportMessage(''); setIsExportDialogOpen(true) }} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: '8px', p: '6px' }}>
                    <Download size={18} />
                  </IconButton>
                </Tooltip>
                <Tooltip title={t('delete')}>
                  <IconButton size="small" onClick={() => { if (confirm(to('deleteConfirm'))) deleteMutation.mutate(selectedOrder.id) }} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: '8px', p: '6px' }}>
                    <Trash2 size={18} />
                  </IconButton>
                </Tooltip>
                <IconButton size="small" onClick={() => setSelectedOrderId(null)} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: '8px', p: '6px' }}>
                  <X size={18} />
                </IconButton>
              </Box>
            </Box>

            {/* Drawer body */}
            <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', p: '24px' }}>
              {/* Order info card */}
              <Paper sx={{ p: '20px', mb: '20px', borderRadius: '12px', boxShadow: 'none', border: '1px solid', borderColor: 'divider' }}>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: '20px' }}>
                  <Box sx={{ minWidth: 130 }}>
                    <Box sx={{ fontSize: '12px', fontWeight: 600, color: 'text.secondary', mb: '4px', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('customer')}</Box>
                    <Box sx={{ fontSize: '14px', fontWeight: 500, display: 'flex', alignItems: 'center', gap: '6px' }}>
                      {selectedOrder.customer_name}
                      {selectedOrder.is_customer_deleted && (
                        <Box component="span" sx={{ fontSize: '11px', px: '6px', py: '2px', borderRadius: '4px', bgcolor: 'action.selected', color: 'text.primary', fontWeight: 500 }}>
                          {to('customerDeleted')}
                        </Box>
                      )}
                    </Box>
                  </Box>
                  <Box>
                    <Box sx={{ fontSize: '12px', fontWeight: 600, color: 'text.secondary', mb: '4px', textTransform: 'uppercase', letterSpacing: '0.04em', display: 'flex', alignItems: 'center', gap: '4px' }}>
                      {to('paymentStatus')}
                      <Tooltip title={to('paymentStatusTooltip')} placement="top" arrow>
                        <Info size={12} style={{ flexShrink: 0 }} />
                      </Tooltip>
                    </Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                      <Box sx={{ display: 'inline-block', px: '8px', py: '3px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, bgcolor: getPaymentStatusStyle(normalizePaymentStatus(selectedOrder.payment_status)).bg, color: getPaymentStatusStyle(normalizePaymentStatus(selectedOrder.payment_status)).color }}>
                        {toPaymentStatus(normalizePaymentStatus(selectedOrder.payment_status || 'outstanding'))}
                      </Box>
                      <Button
                        size="small"
                        variant="outlined"
                        color={normalizePaymentStatus(selectedOrder.payment_status) === 'paid' ? 'error' : 'success'}
                        disabled={updatePaymentStatusMutation.isLoading || markUnpaidMutation.isLoading}
                        onClick={() => {
                          if (normalizePaymentStatus(selectedOrder.payment_status) === 'paid') {
                            setMarkUnpaidDialogOpen(true)
                          } else {
                            updatePaymentStatusMutation.mutate({ id: selectedOrder.id, payment_status: 'paid' })
                          }
                        }}
                        sx={{ fontSize: '12px', py: '2px', px: '8px', minWidth: 0, textTransform: 'none' }}
                      >
                        {normalizePaymentStatus(selectedOrder.payment_status) === 'paid' ? to('markUnpaid') : to('markPaid')}
                      </Button>
                    </Box>
                  </Box>
                  <Box sx={{ ml: { xs: 0, sm: 'auto' } }}>
                    <Box sx={{ fontSize: '12px', fontWeight: 600, color: 'text.secondary', mb: '4px', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('status')}</Box>
                    <Select
                      size="small"
                      value={selectedOrder.status}
                      onChange={(e) => updateStatusMutation.mutate({ id: selectedOrder.id, status: e.target.value })}
                      sx={{ fontSize: '14px', borderRadius: '8px', fontWeight: 500 }}
                      MenuProps={{ anchorOrigin: { vertical: 'bottom', horizontal: 'left' }, transformOrigin: { vertical: 'top', horizontal: 'left' } }}
                    >
                      <MenuItem value="created" sx={{ fontSize: '14px' }}>{toStatus('created')}</MenuItem>
                      <MenuItem value="in_progress" sx={{ fontSize: '14px' }}>{toStatus('in_progress')}</MenuItem>
                      <MenuItem value="in_delivery" sx={{ fontSize: '14px' }}>{toStatus('in_delivery')}</MenuItem>
                      <MenuItem value="done" sx={{ fontSize: '14px' }}>{toStatus('done')}</MenuItem>
                      <MenuItem value="cancelled" sx={{ fontSize: '14px' }}>{toStatus('cancelled')}</MenuItem>
                    </Select>
                  </Box>
                  <Box sx={{ flex: '1 1 100%', minWidth: 0 }}>
                    <Box sx={{ fontSize: '12px', fontWeight: 600, color: 'text.secondary', mb: '4px', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{to('notes')}</Box>
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
                      sx={{ width: '100%', fontSize: '14px', borderRadius: '8px' }}
                    />
                  </Box>
                </Box>
              </Paper>

              {/* Order items */}
              <Paper sx={{ borderRadius: '12px', boxShadow: 'none', border: '1px solid', borderColor: 'divider', overflow: 'hidden', mb: '20px' }}>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', px: '16px', py: '10px', borderBottom: '1px solid', borderColor: 'divider', bgcolor: 'action.hover' }}>
                  <Typography sx={{ fontSize: '14px', fontWeight: 600 }}>{t('items')}</Typography>
                  <Button variant="outlined" onClick={openAddItemForm} sx={{ fontSize: '12px', py: '3px', px: '10px', minWidth: 0 }}>
                    {to('addItem')}
                  </Button>
                </Box>
                {selectedOrder.order_items && selectedOrder.order_items.length > 0 ? (
                  <Box sx={{ overflowX: 'auto' }}>
                    <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 380 }}>
                      <Box component="thead">
                        <Box component="tr" sx={{ bgcolor: 'transparent' }}>
                          <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('product')}</Box>
                          <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('price')}</Box>
                          <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('quantity')}</Box>
                          <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{to('subtotal')}</Box>
                          <Box component="th" sx={{ width: '40px' }}></Box>
                        </Box>
                      </Box>
                      <Box component="tbody">
                        {selectedOrder.order_items.map((item) => (
                          <Box component="tr" key={item.id} sx={{ borderTop: '1px solid', borderColor: 'divider', '&:hover': { bgcolor: 'action.hover' } }}>
                            <Box component="td" sx={{ px: '16px', py: '8px', fontSize: '14px' }}>{item.product_name}</Box>
                            <Box component="td" sx={{ px: '16px', py: '8px', textAlign: 'right', fontSize: '14px' }}>{formatPrice(item.price)}</Box>
                            <Box component="td" sx={{ px: '16px', py: '8px' }}>
                              <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
                                <OutlinedInput
                                  type="number"
                                  inputProps={{ min: '1' }}
                                  defaultValue={item.qty}
                                  size="small"
                                  sx={{ width: '60px', fontSize: '14px', borderRadius: '8px' }}
                                  onBlur={(e) => {
                                    const newQty = Number(e.target.value) || 1
                                    if (newQty !== item.qty && newQty > 0) {
                                      updateItemMutation.mutate({ orderId: selectedOrder.id, itemId: item.id, newQty, itemPrice: item.price, oldQty: item.qty })
                                    }
                                  }}
                                  onKeyDown={(e) => { if (e.key === 'Enter') (e.target as HTMLInputElement).blur() }}
                                />
                              </Box>
                            </Box>
                            <Box component="td" sx={{ px: '16px', py: '8px', textAlign: 'right', fontSize: '14px', fontWeight: 500 }}>{formatPrice(item.price * item.qty)}</Box>
                            <Box component="td" sx={{ px: '8px', py: '8px', textAlign: 'center' }}>
                              <Box
                                component="span"
                                sx={{ color: 'error.main', cursor: 'pointer', display: 'inline-flex', '&:hover': { color: 'error.dark' } }}
                                onClick={() => { if (confirm(to('removeItemConfirm'))) deleteItemMutation.mutate({ orderId: selectedOrder.id, itemId: item.id, itemPrice: item.price, itemQty: item.qty }) }}
                              >
                                <Trash2 size={16} />
                              </Box>
                            </Box>
                          </Box>
                        ))}
                      </Box>
                      <Box component="tfoot">
                        <Box component="tr" sx={{ borderTop: '2px solid', borderColor: 'divider', bgcolor: 'action.hover' }}>
                          <Box component="td" sx={{ px: '16px', py: '10px', textAlign: 'right', fontWeight: 700, fontSize: '14px' }} {...({ colSpan: 3 } as object)}>{t('total')}</Box>
                          <Box component="td" sx={{ px: '16px', py: '10px', textAlign: 'right', fontWeight: 700, fontSize: '14px', color: 'primary.main' }}>{formatPrice(selectedOrder.total_price)}</Box>
                          <Box component="td"></Box>
                        </Box>
                        {selectedOrder.order_payments && selectedOrder.order_payments.length > 0 && (() => {
                          const totalPaid = selectedOrder.order_payments.reduce((sum, p) => sum + p.amount, 0)
                          const remaining = selectedOrder.total_price - totalPaid
                          return (
                            <>
                              <Box component="tr" sx={{ bgcolor: 'action.hover' }}>
                                <Box component="td" sx={{ px: '16px', py: '6px', textAlign: 'right', fontWeight: 600, fontSize: '13px', color: 'text.secondary' }} {...({ colSpan: 3 } as object)}>{to('totalPaid')}</Box>
                                <Box component="td" sx={{ px: '16px', py: '6px', textAlign: 'right', fontWeight: 600, fontSize: '13px', color: 'success.main' }}>{formatPrice(totalPaid)}</Box>
                                <Box component="td"></Box>
                              </Box>
                              <Box component="tr" sx={{ bgcolor: 'action.hover' }}>
                                <Box component="td" sx={{ px: '16px', py: '6px', textAlign: 'right', fontWeight: 600, fontSize: '13px', color: 'text.secondary' }} {...({ colSpan: 3 } as object)}>{to('remaining')}</Box>
                                <Box component="td" sx={{ px: '16px', py: '6px', textAlign: 'right', fontWeight: 700, fontSize: '13px', color: remaining > 0 ? 'error.main' : 'success.main' }}>{formatPrice(remaining)}</Box>
                                <Box component="td"></Box>
                              </Box>
                            </>
                          )
                        })()}
                      </Box>
                    </Box>
                  </Box>
                ) : (
                  <Box sx={{ p: '24px', textAlign: 'center', color: 'text.secondary' }}>
                    <Box sx={{ fontSize: '14px', mb: '8px' }}>{to('noItems')}</Box>
                    <Button variant="outlined" size="small" onClick={openAddItemForm}>{to('addFirstItem')}</Button>
                  </Box>
                )}
              </Paper>

              {/* Order payments */}
              <Paper sx={{ borderRadius: '12px', boxShadow: 'none', border: '1px solid', borderColor: 'divider', overflow: 'hidden' }}>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', px: '16px', py: '10px', borderBottom: '1px solid', borderColor: 'divider', bgcolor: 'action.hover' }}>
                  <Typography sx={{ fontSize: '14px', fontWeight: 600 }}>Payments</Typography>
                  <Button variant="outlined" onClick={() => { setAddPaymentAmount(0); setIsAddPaymentFormOpen(true) }} sx={{ fontSize: '12px', py: '3px', px: '10px', minWidth: 0 }}>
                    {to('addPayment')}
                  </Button>
                </Box>
                {selectedOrder.order_payments && selectedOrder.order_payments.length > 0 ? (
                  <Box sx={{ overflowX: 'auto' }}>
                    <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 280 }}>
                      <Box component="thead">
                        <Box component="tr">
                          <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{to('amount')}</Box>
                          <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{to('created')}</Box>
                          <Box component="th" sx={{ width: '60px' }}></Box>
                        </Box>
                      </Box>
                      <Box component="tbody">
                        {selectedOrder.order_payments.map((payment) => (
                          <Box component="tr" key={payment.id} sx={{ borderTop: '1px solid', borderColor: 'divider', '&:hover': { bgcolor: 'action.hover' } }}>
                            <Box component="td" sx={{ px: '16px', py: '8px', fontSize: '14px', fontWeight: 500 }}>{formatPrice(payment.amount)}</Box>
                            <Box component="td" sx={{ px: '16px', py: '8px', fontSize: '13px', color: 'text.secondary' }}>{formatDate(payment.created_at)}</Box>
                            <Box component="td" sx={{ px: '8px', py: '8px', textAlign: 'center' }}>
                              <Box sx={{ display: 'inline-flex', gap: '8px', alignItems: 'center' }}>
                                <Box component="span" sx={{ color: 'text.secondary', cursor: 'pointer', display: 'inline-flex', '&:hover': { color: 'text.primary' } }} onClick={() => setEditingPayment({ id: payment.id, amount: payment.amount })}>
                                  <Pencil size={15} />
                                </Box>
                                <Box component="span" sx={{ color: 'error.main', cursor: 'pointer', display: 'inline-flex', '&:hover': { color: 'error.dark' } }} onClick={() => { if (confirm(to('removePaymentConfirm'))) deletePaymentMutation.mutate({ orderId: selectedOrder.id, paymentId: payment.id }) }}>
                                  <Trash2 size={15} />
                                </Box>
                              </Box>
                            </Box>
                          </Box>
                        ))}
                      </Box>
                    </Box>
                  </Box>
                ) : (
                  <Box sx={{ p: '24px', textAlign: 'center', color: 'text.secondary' }}>
                    <Box sx={{ fontSize: '14px', mb: '8px' }}>{to('noPayments')}</Box>
                    <Button variant="outlined" size="small" onClick={() => { setAddPaymentAmount(0); setIsAddPaymentFormOpen(true) }}>{to('addFirstPayment')}</Button>
                  </Box>
                )}
              </Paper>
            </Box>
          </Box>
        )}
      </Drawer>

      {/* Create Order Modal */}
      <Dialog open={isCreateFormOpen} onClose={closeCreateForm} fullWidth maxWidth="sm">
        {createFormConflict ? (
          <>
            <DialogTitle sx={{ pb: '8px' }}>{to('duplicateOrderTitle')}</DialogTitle>
            <DialogContent sx={{ pt: '8px', pb: 0 }}>
              <Box sx={{ mb: '8px', color: 'text.secondary' }}>
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
                  onChange={(id) => setCreateForm({ ...createForm, customer_id: id, customer_name: undefined })}
                  placeholder={to('selectCustomer')}
                  searchPlaceholder={to('searchCustomer')}
                  noResultsText={to('noCustomersFound')}
                  selectedLabel={createForm.customer_name}
                  onCreateCustomer={(term) => {
                    const isPhone = /^\d+$/.test(term)
                    setCreateCustomerDialog({ open: true, name: isPhone ? '' : term, phone: isPhone ? term : '' })
                  }}
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
                  sx={{ width: '100%', py: '8px', px: '16px', fontSize: '14px', borderRadius: '8px', border: '1px solid', borderColor: 'divider', resize: 'vertical' }}
                />
              </Box>
            </DialogContent>
            <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
              <Button type="button" variant="outlined" onClick={closeCreateForm}>{t('cancel')}</Button>
              <Button type="submit" variant="contained" disableElevation disabled={createMutation.isLoading || !createForm.customer_id}>
                {to('createOrder')}
              </Button>
            </DialogActions>
          </Box>
        )}
      </Dialog>

      {/* Export Invoice Dialog */}
      <Dialog open={isExportDialogOpen} onClose={() => setIsExportDialogOpen(false)} fullWidth maxWidth="sm">
        <Box component="form" onSubmit={(e: React.FormEvent) => { e.preventDefault(); if (selectedOrder) exportMutation.mutate({ orderId: selectedOrder.id, message: exportMessage }) }}>
          <DialogTitle sx={{ pb: '8px' }}>{to('exportInvoiceTitle')}</DialogTitle>
          <DialogContent sx={{ pt: '8px', pb: 0 }}>
            <Box sx={{ mb: '8px' }}>
              <Box component="label" htmlFor="export-message" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{to('exportMessage')}</Box>
              <OutlinedInput
                id="export-message"
                value={exportMessage}
                onChange={(e) => setExportMessage(e.target.value)}
                placeholder={to('exportMessagePlaceholder')}
                multiline
                rows={4}
                size="small"
                sx={{ width: '100%', py: '8px', px: '16px', fontSize: '14px', borderRadius: '8px', border: '1px solid', borderColor: 'divider' }}
              />
            </Box>
          </DialogContent>
          <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
            <Button type="button" variant="outlined" onClick={() => setIsExportDialogOpen(false)}>{t('cancel')}</Button>
            <Button type="submit" variant="contained" disableElevation disabled={exportMutation.isLoading} startIcon={<Download size={16} />}>
              {to('exportDownload')}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Add Item Modal */}
      <Dialog open={isAddItemFormOpen} onClose={closeAddItemForm} fullWidth maxWidth="xs">
        <Box component="form" onSubmit={submitAddItemForm}>
          <DialogTitle sx={{ pb: '8px' }}>{to('addItemTitle')}</DialogTitle>
          <DialogContent sx={{ pt: '8px', pb: 0, minHeight: '180px' }}>
            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('product')}</Box>
              <ProductSearchSelect
                value={addItemForm.product_id}
                onChange={(id, price, name) => setAddItemForm((f) => ({ ...f, product_id: id, product_name: name ?? '', product_price: price ?? 0 }))}
                placeholder={to('selectProduct')}
                searchPlaceholder={to('searchProduct')}
                noResultsText={to('noProductsFound')}
                onCreateProduct={(term) => setCreateProductDialog({ open: true, name: term, price: '' })}
                selectedLabel={addItemForm.product_name}
              />
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
            <Button type="button" variant="outlined" onClick={closeAddItemForm}>{t('cancel')}</Button>
            <Button type="submit" variant="contained" disableElevation disabled={addItemMutation.isLoading || !addItemForm.product_id}>
              {to('addItemButton')}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Edit Payment Modal */}
      <Dialog open={!!editingPayment} onClose={() => setEditingPayment(null)} fullWidth maxWidth="xs">
        <Box component="form" onSubmit={(e: React.FormEvent) => { e.preventDefault(); if (selectedOrderId && editingPayment && editingPayment.amount > 0) updatePaymentMutation.mutate({ orderId: selectedOrderId, paymentId: editingPayment.id, amount: editingPayment.amount }, { onSuccess: () => setEditingPayment(null) }) }}>
          <DialogTitle sx={{ pb: '8px' }}>{to('editPaymentTitle')}</DialogTitle>
          <DialogContent sx={{ pt: '8px', pb: 0 }}>
            <Box sx={{ mb: '8px' }}>
              <Box component="label" htmlFor="edit-payment-amount" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{to('amount')}</Box>
              <OutlinedInput
                id="edit-payment-amount"
                type="number"
                inputProps={{ min: '1' }}
                value={editingPayment?.amount || ''}
                onChange={(e) => setEditingPayment(editingPayment ? { ...editingPayment, amount: Number(e.target.value) || 0 } : null)}
                size="small"
                sx={{ width: '100%', borderRadius: '8px' }}
                required
                autoFocus
              />
            </Box>
          </DialogContent>
          <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
            <Button type="button" variant="outlined" onClick={() => setEditingPayment(null)}>{t('cancel')}</Button>
            <Button type="submit" variant="contained" disableElevation disabled={updatePaymentMutation.isLoading || !editingPayment || editingPayment.amount <= 0}>
              {t('save')}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Add Payment Modal */}
      <Dialog open={isAddPaymentFormOpen} onClose={() => setIsAddPaymentFormOpen(false)} fullWidth maxWidth="xs">
        <Box component="form" onSubmit={(e: React.FormEvent) => { e.preventDefault(); if (selectedOrderId && addPaymentAmount > 0) addPaymentMutation.mutate({ orderId: selectedOrderId, amount: addPaymentAmount }) }}>
          <DialogTitle sx={{ pb: '8px' }}>{to('addPaymentTitle')}</DialogTitle>
          <DialogContent sx={{ pt: '8px', pb: 0 }}>
            <Box sx={{ mb: '8px' }}>
              <Box component="label" htmlFor="payment-amount" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{to('amount')}</Box>
              <OutlinedInput
                id="payment-amount"
                type="number"
                inputProps={{ min: '1' }}
                value={addPaymentAmount || ''}
                onChange={(e) => setAddPaymentAmount(Number(e.target.value) || 0)}
                size="small"
                sx={{ width: '100%', borderRadius: '8px' }}
                required
              />
            </Box>
          </DialogContent>
          <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
            <Button type="button" variant="outlined" onClick={() => setIsAddPaymentFormOpen(false)}>{t('cancel')}</Button>
            <Button type="submit" variant="contained" disableElevation disabled={addPaymentMutation.isLoading || addPaymentAmount <= 0}>
              {to('addPaymentButton')}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Create Customer Dialog */}
      <Dialog open={createCustomerDialog.open} onClose={() => setCreateCustomerDialog({ open: false, name: '', phone: '' })} fullWidth maxWidth="xs">
        <Box component="form" onSubmit={(e: React.FormEvent) => { e.preventDefault(); createCustomerMutation.mutate({ name: createCustomerDialog.name, phone: createCustomerDialog.phone }) }}>
          <DialogTitle sx={{ pb: '8px' }}>{t('create')} {t('customer')}</DialogTitle>
          <DialogContent sx={{ pt: '8px !important', pb: 0 }}>
            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('name')}</Box>
              <OutlinedInput value={createCustomerDialog.name} onChange={(e) => setCreateCustomerDialog((d) => ({ ...d, name: e.target.value }))} size="small" fullWidth required />
            </Box>
            <Box sx={{ mb: '8px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('phone')}</Box>
              <OutlinedInput value={createCustomerDialog.phone} onChange={(e) => setCreateCustomerDialog((d) => ({ ...d, phone: e.target.value }))} size="small" fullWidth required />
            </Box>
          </DialogContent>
          <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
            <Button type="button" variant="outlined" onClick={() => setCreateCustomerDialog({ open: false, name: '', phone: '' })}>{t('cancel')}</Button>
            <Button type="submit" variant="contained" disableElevation disabled={createCustomerMutation.isLoading}>{t('create')}</Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Create Product Dialog */}
      <Dialog open={createProductDialog.open} onClose={() => setCreateProductDialog({ open: false, name: '', price: '' })} fullWidth maxWidth="xs">
        <Box component="form" onSubmit={(e: React.FormEvent) => { e.preventDefault(); const price = parseFloat(createProductDialog.price); if (createProductDialog.name && !isNaN(price) && price > 0) createProductMutation.mutate({ name: createProductDialog.name, price }) }}>
          <DialogTitle sx={{ pb: '8px' }}>{t('create')} {t('product')}</DialogTitle>
          <DialogContent sx={{ pt: '8px !important', pb: 0 }}>
            <Box sx={{ mb: '16px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('name')}</Box>
              <OutlinedInput value={createProductDialog.name} onChange={(e) => setCreateProductDialog((d) => ({ ...d, name: e.target.value }))} size="small" fullWidth required autoFocus />
            </Box>
            <Box sx={{ mb: '8px' }}>
              <Box component="label" sx={{ display: 'block', mb: '4px', fontSize: '14px', fontWeight: 600 }}>{t('price')}</Box>
              <OutlinedInput type="number" inputProps={{ min: '0', step: 'any' }} value={createProductDialog.price} onChange={(e) => setCreateProductDialog((d) => ({ ...d, price: e.target.value }))} size="small" fullWidth required />
            </Box>
          </DialogContent>
          <DialogActions sx={{ px: '24px', pt: '16px', pb: '24px', gap: '8px' }}>
            <Button type="button" variant="outlined" onClick={() => setCreateProductDialog({ open: false, name: '', price: '' })}>{t('cancel')}</Button>
            <Button type="submit" variant="contained" disableElevation disabled={createProductMutation.isLoading}>{t('create')}</Button>
          </DialogActions>
        </Box>
      </Dialog>
      {/* Mark Unpaid Confirmation Dialog */}
      <Dialog open={markUnpaidDialogOpen} onClose={() => setMarkUnpaidDialogOpen(false)} fullWidth maxWidth="xs">
        <DialogTitle sx={{ pb: '8px' }}>{to('markUnpaidConfirmTitle')}</DialogTitle>
        <DialogContent sx={{ pt: '8px !important' }}>
          <Typography sx={{ fontSize: '14px', color: 'text.secondary' }}>
            {to('markUnpaidConfirmMessage')}
          </Typography>
        </DialogContent>
        <DialogActions sx={{ px: '24px', pt: '8px', pb: '24px', gap: '8px' }}>
          <Button variant="outlined" onClick={() => setMarkUnpaidDialogOpen(false)} disabled={markUnpaidMutation.isLoading}>
            {t('cancel')}
          </Button>
          <Button
            variant="contained"
            color="error"
            disableElevation
            disabled={markUnpaidMutation.isLoading}
            onClick={() => { if (selectedOrder) markUnpaidMutation.mutate(selectedOrder.id) }}
          >
            {to('markUnpaid')}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  )
}
