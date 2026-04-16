"use client"

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, Drawer, IconButton, InputBase, MenuItem, Paper, Select, Tooltip, Typography, useTheme } from '@mui/material'
import { ListFilter, Search, X } from 'lucide-react'
import DateRangeFilter from '@/components/ui/DateRangeFilter'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { api } from '@/utils/api'

type TempOrderItem = {
  id: number
  product_name: string
  price: number
  qty: number
  created_at?: string
}

type TempOrder = {
  id: number
  customer_name: string
  customer_phone: string
  total_price: number
  status: string
  order_items?: TempOrderItem[]
  created_at: string
  updated_at?: string | null
}

const lightStatusColors: Record<string, { bg: string; color: string }> = {
  pending:  { bg: '#FFF3E0', color: '#E65100' },
  accepted: { bg: '#E8F5E9', color: '#2E7D32' },
  rejected: { bg: '#FFEBEE', color: '#C62828' },
}

const darkStatusColors: Record<string, { bg: string; color: string }> = {
  pending:  { bg: '#3e2000', color: '#ffb74d' },
  accepted: { bg: '#1b3a2d', color: '#81c784' },
  rejected: { bg: '#3e1a1a', color: '#ef9a9a' },
}

export default function TempOrdersPage() {
  const t = useTranslations('common')
  const to = useTranslations('orders')
  const tTemp = useTranslations('tempOrders')
  const tErrors = useTranslations('errors')
  const toStatus = useTranslations('orderStatus')
  const theme = useTheme()
  const statusColors = theme.palette.mode === 'dark' ? darkStatusColors : lightStatusColors
  const queryClient = useQueryClient()
  const [selectedTempOrderId, setSelectedTempOrderId] = useState<number | null>(null)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('pending')
  const [dateFrom, setDateFrom] = useState<string>('')
  const [dateTo, setDateTo] = useState<string>('')
  const [filtersVisible, setFiltersVisible] = useState(false)
  const [acceptLoading, setAcceptLoading] = useState(false)
  const [acceptError, setAcceptError] = useState<string | null>(null)
  const [acceptSuccess, setAcceptSuccess] = useState(false)
  const [rejectLoading, setRejectLoading] = useState(false)
  const [rejectSuccess, setRejectSuccess] = useState(false)
  const [rejectError, setRejectError] = useState<string | null>(null)
  const [showActiveOrderConflictDialog, setShowActiveOrderConflictDialog] = useState(false)
  const [conflictData, setConflictData] = useState<{ customerId: number; activeOrderId: number } | null>(null)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const { data: tempOrdersRes, isLoading, isError, error } = useQuery(
    ['temp_orders', debouncedSearch, statusFilter, dateFrom, dateTo],
    async () => {
      const opts: { search?: string; status?: string; date_from?: string; date_to?: string } = {}
      if (debouncedSearch) opts.search = debouncedSearch
      if (statusFilter && statusFilter !== 'all') opts.status = statusFilter
      if (dateFrom) opts.date_from = dateFrom
      if (dateTo) opts.date_to = dateTo
      const res = await api.getTempOrders(opts)
      if (!res.success) throw new Error(res.message || tTemp('fetchFailed'))
      return res.data as TempOrder[]
    },
    { keepPreviousData: true }
  )

  const { data: selectedTempOrderDetails } = useQuery(
    ['temp_order', selectedTempOrderId],
    async () => {
      if (!selectedTempOrderId) return null
      const res = await api.getTempOrder(selectedTempOrderId)
      if (!res.success) throw new Error(res.message || tTemp('fetchDetailsFailed'))
      return res.data as TempOrder
    },
    { enabled: !!selectedTempOrderId }
  )

  useEffect(() => {
    setAcceptError(null)
    setAcceptSuccess(false)
    setRejectSuccess(false)
    setRejectError(null)
    setShowActiveOrderConflictDialog(false)
    setConflictData(null)
  }, [selectedTempOrderId])

  const selectedTempOrder: TempOrder | null = useMemo(() => {
    if (!selectedTempOrderDetails) {
      if (!tempOrdersRes) return null
      return tempOrdersRes.find((o) => o.id === selectedTempOrderId) || null
    }
    return selectedTempOrderDetails
  }, [selectedTempOrderDetails, tempOrdersRes, selectedTempOrderId])

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

  const handleAccept = useCallback(async () => {
    if (!selectedTempOrder) return
    setAcceptError(null)
    setAcceptSuccess(false)
    setAcceptLoading(true)
    try {
      const res = await api.checkActiveOrder({
        phone: selectedTempOrder.customer_phone,
        name: selectedTempOrder.customer_name,
      })
      if (!res.success) {
        setAcceptError(res.message || tTemp('acceptError'))
        return
      }
      const customerId = res.data!.customer_id
      const activeOrderId = res.data!.active_order_id ?? 0

      if (activeOrderId > 0) {
        setConflictData({ customerId, activeOrderId })
        setShowActiveOrderConflictDialog(true)
        return
      }

      const mergeRes = await api.mergeTempOrder({ temp_order_id: selectedTempOrder.id, customer_id: customerId })
      if (mergeRes.success) {
        setAcceptSuccess(true)
        await Promise.all([
          queryClient.invalidateQueries(['temp_orders', debouncedSearch, statusFilter, dateFrom, dateTo]),
          queryClient.invalidateQueries(['temp_order', selectedTempOrder.id]),
        ])
      } else {
        setAcceptError(mergeRes.message || tTemp('acceptError'))
      }
    } catch (e: unknown) {
      setAcceptError((e as Error)?.message || tTemp('acceptError'))
    } finally {
      setAcceptLoading(false)
    }
  }, [selectedTempOrder, tTemp, queryClient, debouncedSearch, statusFilter])

  const handleConfirmMergeIntoActiveOrder = useCallback(async () => {
    if (!selectedTempOrder || !conflictData) return
    setAcceptError(null)
    setAcceptLoading(true)
    try {
      const mergeRes = await api.mergeTempOrder({
        temp_order_id: selectedTempOrder.id,
        customer_id: conflictData.customerId,
        active_order_id: conflictData.activeOrderId,
      })
      if (mergeRes.success) {
        setAcceptSuccess(true)
        setShowActiveOrderConflictDialog(false)
        setConflictData(null)
        await Promise.all([
          queryClient.invalidateQueries(['temp_orders', debouncedSearch, statusFilter, dateFrom, dateTo]),
          queryClient.invalidateQueries(['temp_order', selectedTempOrder.id]),
        ])
      } else {
        setAcceptError(mergeRes.message || tTemp('acceptError'))
      }
    } catch (e: unknown) {
      setAcceptError((e as Error)?.message || tTemp('acceptError'))
    } finally {
      setAcceptLoading(false)
    }
  }, [selectedTempOrder, conflictData, tTemp, queryClient, debouncedSearch, statusFilter])

  const handleReject = useCallback(async () => {
    if (!selectedTempOrder) return
    setRejectError(null)
    setRejectSuccess(false)
    setAcceptError(null)
    setAcceptSuccess(false)
    setRejectLoading(true)
    try {
      const res = await api.rejectTempOrder(selectedTempOrder.id)
      if (res.success) {
        setRejectSuccess(true)
        await Promise.all([
          queryClient.invalidateQueries(['temp_orders', debouncedSearch, statusFilter, dateFrom, dateTo]),
          queryClient.invalidateQueries(['temp_order', selectedTempOrder.id]),
        ])
      } else {
        setRejectError(res.message || tTemp('rejectError'))
      }
    } catch (e: unknown) {
      setRejectError((e as Error)?.message || tTemp('rejectError'))
    } finally {
      setRejectLoading(false)
    }
  }, [selectedTempOrder, tTemp, queryClient, debouncedSearch, statusFilter])

  const drawerOpen = selectedTempOrderId !== null

  return (
    <>
      <Container disableGutters maxWidth={false} sx={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          {isLoading && <PageLoadingSkeleton />}
          {isError && (
            <Box sx={{ p: '24px', color: 'error.main' }}>
              {(error as Error)?.message || tErrors('loadingError', { resource: tTemp('title') })}
            </Box>
          )}

          {!isLoading && !isError && (
            <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
              {/* Toolbar */}
              <Box sx={{ px: '24px', pt: '24px', pb: '16px', flexShrink: 0 }}>
                <Box sx={{ maxWidth: 1200, mx: 'auto' }}>
                <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                  <Paper
                    variant="outlined"
                    sx={{ flex: 1, display: 'flex', alignItems: 'center', px: '8px', py: '4px', borderRadius: '10px', boxShadow: 'none', gap: '2px' }}
                  >
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
                      placeholder={tTemp('searchPlaceholder')}
                      sx={{ flex: 1, mx: '8px', fontSize: '13px', color: 'text.primary' }}
                      inputProps={{ 'aria-label': tTemp('searchPlaceholder') }}
                    />

                    <Search size={18} style={{ color: '#9e9e9e', flexShrink: 0, marginRight: '4px' }} />
                  </Paper>
                </Box>

                {/* Filters row */}
                {filtersVisible && (
                  <Box sx={{ display: 'flex', gap: '12px', mt: '12px', flexWrap: 'wrap', alignItems: 'flex-end' }}>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                      <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>{t('status')}</Box>
                      <Select
                        size="small"
                        value={statusFilter}
                        onChange={(e) => setStatusFilter(e.target.value)}
                        sx={{ fontSize: '13px', minWidth: 130 }}
                      >
                        {(['all', 'pending', 'accepted', 'rejected'] as const).map((s) => (
                          <MenuItem key={s} value={s} sx={{ fontSize: '13px' }}>{toStatus(s)}</MenuItem>
                        ))}
                      </Select>
                    </Box>
                    <DateRangeFilter
                      dateFrom={dateFrom}
                      dateTo={dateTo}
                      onDateFromChange={setDateFrom}
                      onDateToChange={setDateTo}
                    />
                  </Box>
                )}
                </Box>
              </Box>

              {/* Table */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', px: '24px', pb: '24px' }}>
                <Box sx={{ maxWidth: 1200, mx: 'auto' }}>
                {(tempOrdersRes || []).length === 0 ? (
                  <Box sx={{ textAlign: 'center', py: '64px', color: 'text.secondary', fontSize: '14px' }}>{tTemp('noOrders')}</Box>
                ) : (
                  <Paper sx={{ borderRadius: '12px', border: '1px solid', borderColor: 'divider', boxShadow: 'none', overflow: 'hidden' }}>
                    <Box sx={{ overflowX: 'auto' }}>
                      <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 480 }}>
                        <Box component="thead">
                          <Box component="tr" sx={{ bgcolor: 'action.hover', borderBottom: '1px solid', borderColor: 'divider' }}>
                            <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', width: '80px' }}>
                              #
                            </Box>
                            <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                              {t('customer')}
                            </Box>
                            <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', display: { xs: 'none', sm: 'table-cell' } }}>
                              {tTemp('phone')}
                            </Box>
                            <Box component="th" sx={{ px: '16px', py: '12px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', display: { xs: 'none', md: 'table-cell' } }}>
                              {t('status')}
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
                          {(tempOrdersRes || []).map((o) => {
                            const statusStyle = getStatusStyle(o.status)
                            return (
                              <Box
                                component="tr"
                                key={o.id}
                                onClick={() => setSelectedTempOrderId(o.id)}
                                sx={{
                                  borderTop: '1px solid',
                                  borderColor: 'divider',
                                  cursor: 'pointer',
                                  transition: 'background 0.1s',
                                  bgcolor: selectedTempOrderId === o.id ? 'action.selected' : 'transparent',
                                  '&:hover': { bgcolor: 'action.hover' },
                                }}
                              >
                                <Box component="td" sx={{ px: '16px', py: '14px', fontSize: '14px', fontWeight: 600, color: 'text.secondary', whiteSpace: 'nowrap' }}>
                                  #{o.id}
                                </Box>
                                <Box component="td" sx={{ px: '16px', py: '14px', fontSize: '14px', fontWeight: 500 }}>
                                  {o.customer_name}
                                </Box>
                                <Box component="td" sx={{ px: '16px', py: '14px', fontSize: '13px', color: 'text.secondary', display: { xs: 'none', sm: 'table-cell' } }}>
                                  {o.customer_phone}
                                </Box>
                                <Box component="td" sx={{ px: '16px', py: '14px', display: { xs: 'none', md: 'table-cell' } }}>
                                  <Box sx={{ display: 'inline-block', px: '8px', py: '3px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, bgcolor: statusStyle.bg, color: statusStyle.color }}>
                                    {toStatus(o.status) || o.status}
                                  </Box>
                                </Box>
                                <Box component="td" sx={{ px: '16px', py: '14px', textAlign: 'right', fontSize: '14px', fontWeight: 600, whiteSpace: 'nowrap' }}>
                                  Rp {formatPrice(o.total_price)}
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
          onClose={() => setSelectedTempOrderId(null)}
          slotProps={{
            paper: {
              sx: {
                width: { xs: '100%', sm: 560 },
                maxWidth: '100%',
              },
            },
          }}
        >
          {selectedTempOrder && (
            <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              {/* Drawer header */}
              <Box sx={{ px: '24px', py: '16px', borderBottom: '1px solid', borderColor: 'divider', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexShrink: 0 }}>
                <Box>
                  <Typography sx={{ fontSize: '18px', fontWeight: 600 }}>{tTemp('orderNumber', { id: selectedTempOrder.id })}</Typography>
                  <Box sx={{ fontSize: '12px', color: 'text.secondary', mt: '2px' }}>{formatDate(selectedTempOrder.created_at)}</Box>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <Box sx={{ display: 'inline-block', px: '8px', py: '3px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, bgcolor: getStatusStyle(selectedTempOrder.status).bg, color: getStatusStyle(selectedTempOrder.status).color }}>
                    {toStatus(selectedTempOrder.status) || selectedTempOrder.status}
                  </Box>
                  <IconButton size="small" onClick={() => setSelectedTempOrderId(null)} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: '8px', p: '6px' }}>
                    <X size={18} />
                  </IconButton>
                </Box>
              </Box>

              {/* Drawer body */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', p: '24px' }}>
                {/* Accept / Reject */}
                <Box sx={{ display: 'flex', gap: '8px', mb: '20px' }}>
                  <Button
                    variant={selectedTempOrder.status === 'rejected' ? 'outlined' : 'contained'}
                    disableElevation
                    fullWidth
                    onClick={handleAccept}
                    disabled={acceptLoading || rejectLoading || selectedTempOrder.status !== 'pending'}
                    sx={{
                      ...(selectedTempOrder.status === 'rejected'
                        ? { bgcolor: 'action.hover', color: 'text.primary', border: '1px solid', borderColor: 'grey.200' }
                        : { bgcolor: 'primary.main', color: 'white' }),
                      '&.Mui-disabled': selectedTempOrder.status === 'accepted'
                        ? { bgcolor: 'success.main', color: 'white', borderColor: 'transparent' }
                        : { bgcolor: 'action.disabledBackground', color: 'action.disabled', borderColor: 'transparent' },
                    }}
                  >
                    {acceptLoading ? tTemp('accepting') : tTemp('accept')}
                  </Button>
                  <Button
                    variant={selectedTempOrder.status === 'rejected' ? 'contained' : 'outlined'}
                    disableElevation
                    fullWidth
                    onClick={handleReject}
                    disabled={rejectLoading || selectedTempOrder.status !== 'pending'}
                    sx={{
                      ...(selectedTempOrder.status === 'rejected'
                        ? { bgcolor: 'primary.main', color: 'white', border: 'none' }
                        : { bgcolor: 'action.hover', color: 'text.primary', border: '1px solid', borderColor: 'grey.200' }),
                      '&.Mui-disabled': selectedTempOrder.status === 'rejected'
                        ? { bgcolor: 'error.main', color: 'white', borderColor: 'transparent' }
                        : { bgcolor: 'action.disabledBackground', color: 'action.disabled', borderColor: 'transparent' },
                    }}
                  >
                    {rejectLoading ? tTemp('rejecting') : tTemp('reject')}
                  </Button>
                </Box>

                {(acceptError || acceptSuccess || rejectError || rejectSuccess) && (
                  <Box sx={{ mb: '20px', p: '10px 14px', borderRadius: '8px', bgcolor: acceptError || rejectError ? 'error.light' : 'success.light', color: acceptError || rejectError ? 'error.dark' : 'success.dark', fontSize: '14px' }}>
                    {acceptError || rejectError || (acceptSuccess ? tTemp('acceptSuccess') : '') || (rejectSuccess ? tTemp('rejectSuccess') : '')}
                  </Box>
                )}

                {/* Info card */}
                <Paper sx={{ p: '20px', mb: '20px', borderRadius: '12px', boxShadow: 'none', border: '1px solid', borderColor: 'divider' }}>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: '20px' }}>
                    <Box>
                      <Box sx={{ fontSize: '12px', color: 'text.secondary', mb: '4px', px: '2px' }}>{t('customer')}</Box>
                      <Box sx={{ fontSize: '14px', fontWeight: 500 }}>{selectedTempOrder.customer_name}</Box>
                    </Box>
                    <Box>
                      <Box sx={{ fontSize: '12px', color: 'text.secondary', mb: '4px', px: '2px' }}>{tTemp('phone')}</Box>
                      <Box sx={{ fontSize: '14px' }}>{selectedTempOrder.customer_phone}</Box>
                    </Box>
                    <Box>
                      <Box sx={{ fontSize: '12px', color: 'text.secondary', mb: '4px', px: '2px' }}>{to('created')}</Box>
                      <Box sx={{ fontSize: '14px' }}>{formatDate(selectedTempOrder.created_at)}</Box>
                    </Box>
                  </Box>
                </Paper>

                {/* Order items */}
                <Paper sx={{ borderRadius: '12px', boxShadow: 'none', border: '1px solid', borderColor: 'divider', overflow: 'hidden' }}>
                  <Box sx={{ px: '16px', py: '10px', borderBottom: '1px solid', borderColor: 'divider', bgcolor: 'action.hover' }}>
                    <Typography sx={{ fontSize: '14px', fontWeight: 600 }}>{t('items')}</Typography>
                  </Box>
                  {selectedTempOrder.order_items && selectedTempOrder.order_items.length > 0 ? (
                    <Box sx={{ overflowX: 'auto' }}>
                      <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 340 }}>
                        <Box component="thead">
                          <Box component="tr">
                            <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('product')}</Box>
                            <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('price')}</Box>
                            <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{t('quantity')}</Box>
                            <Box component="th" sx={{ px: '16px', py: '10px', textAlign: 'right', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.04em' }}>{to('subtotal')}</Box>
                          </Box>
                        </Box>
                        <Box component="tbody">
                          {selectedTempOrder.order_items.map((item) => (
                            <Box component="tr" key={item.id} sx={{ borderTop: '1px solid', borderColor: 'divider', '&:hover': { bgcolor: 'action.hover' } }}>
                              <Box component="td" sx={{ px: '16px', py: '8px', fontSize: '14px' }}>{item.product_name}</Box>
                              <Box component="td" sx={{ px: '16px', py: '8px', textAlign: 'right', fontSize: '14px' }}>Rp {formatPrice(item.price)}</Box>
                              <Box component="td" sx={{ px: '16px', py: '8px', textAlign: 'right', fontSize: '14px' }}>{item.qty}</Box>
                              <Box component="td" sx={{ px: '16px', py: '8px', textAlign: 'right', fontSize: '14px', fontWeight: 500 }}>Rp {formatPrice(item.price * item.qty)}</Box>
                            </Box>
                          ))}
                        </Box>
                        <Box component="tfoot">
                          <Box component="tr" sx={{ borderTop: '2px solid', borderColor: 'divider', bgcolor: 'action.hover' }}>
                            <Box component="td" sx={{ px: '16px', py: '10px', textAlign: 'right', fontWeight: 700, fontSize: '14px' }} {...({ colSpan: 3 } as object)}>{t('total')}</Box>
                            <Box component="td" sx={{ px: '16px', py: '10px', textAlign: 'right', fontWeight: 700, fontSize: '14px', color: 'primary.main' }}>Rp {formatPrice(selectedTempOrder.total_price)}</Box>
                          </Box>
                        </Box>
                      </Box>
                    </Box>
                  ) : (
                    <Box sx={{ p: '24px', textAlign: 'center', color: 'text.secondary', fontSize: '14px' }}>{to('noItems')}</Box>
                  )}
                </Paper>
              </Box>
            </Box>
          )}
        </Drawer>
      </Container>

      {/* Active order conflict dialog */}
      {showActiveOrderConflictDialog && (
        <Box
          sx={{ position: 'fixed', inset: 0, bgcolor: 'rgba(0,0,0,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center', p: '16px', zIndex: 1300 }}
          onClick={(e) => { if (e.target === e.currentTarget) setShowActiveOrderConflictDialog(false) }}
        >
          <Paper sx={{ width: { xs: '100%', sm: '540px' }, p: '24px' }} onClick={(e) => e.stopPropagation()}>
            <Typography component="h3" sx={{ mb: '8px', fontWeight: 600 }}>{to('duplicateOrderTitle')}</Typography>
            <Box sx={{ mb: '24px', color: 'text.secondary', fontSize: '14px' }}>{to('duplicateOrderMessageInline')}</Box>
            <Box sx={{ gap: '8px', justifyContent: 'flex-end', display: 'flex' }}>
              <Button variant="outlined" type="button" onClick={() => setShowActiveOrderConflictDialog(false)}>{t('cancel')}</Button>
              <Button variant="contained" disableElevation type="button" onClick={handleConfirmMergeIntoActiveOrder} disabled={acceptLoading}>
                {acceptLoading ? tTemp('accepting') : tTemp('confirmMerge')}
              </Button>
            </Box>
          </Paper>
        </Box>
      )}
    </>
  )
}
