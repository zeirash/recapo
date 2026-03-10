"use client"

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Container, Paper, Typography } from '@mui/material'
import Layout from '@/components/Layout'
import { ClipboardList } from 'lucide-react'
import SearchInput from '@/components/ui/SearchInput'
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

const statusColors: Record<string, { bg: string; color: string }> = {
  pending: { bg: '#FFF3E0', color: '#E65100' },
  accepted: { bg: '#E8F5E9', color: '#2E7D32' },
  rejected: { bg: '#FFEBEE', color: '#C62828' },
}

export default function TempOrdersPage() {
  const t = useTranslations('common')
  const to = useTranslations('orders')
  const tTemp = useTranslations('tempOrders')
  const tErrors = useTranslations('errors')
  const toStatus = useTranslations('orderStatus')
  const queryClient = useQueryClient()
  const [selectedTempOrderId, setSelectedTempOrderId] = useState<number | null>(null)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<string[]>(['pending'])
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
    ['temp_orders', debouncedSearch, statusFilter],
    async () => {
      const opts: { search?: string; status?: string } = {}
      if (debouncedSearch) opts.search = debouncedSearch
      if (statusFilter.length > 0) opts.status = statusFilter.join(',')
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
    if (!selectedTempOrderId && tempOrdersRes && tempOrdersRes.length > 0) {
      setSelectedTempOrderId(tempOrdersRes[0].id)
    }
  }, [tempOrdersRes, selectedTempOrderId])

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
    return statusColors[status] || { bg: '#F5F5F5', color: '#616161' }
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
      const hasActiveOrders = activeOrderId > 0

      if (hasActiveOrders) {
        setConflictData({ customerId, activeOrderId })
        setShowActiveOrderConflictDialog(true)
        return
      }

      const mergeRes = await api.mergeTempOrder({
        temp_order_id: selectedTempOrder.id,
        customer_id: customerId,
      })
      if (mergeRes.success) {
        setAcceptSuccess(true)
        await Promise.all([
          queryClient.invalidateQueries(['temp_orders', debouncedSearch, statusFilter]),
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
          queryClient.invalidateQueries(['temp_orders', debouncedSearch, statusFilter]),
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
          queryClient.invalidateQueries(['temp_orders', debouncedSearch, statusFilter]),
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

  return (
    <Layout>
      <Container disableGutters maxWidth={false} sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden', display: 'flex' }}>
          {isLoading && <Box>{t('loading')}</Box>}
          {isError && (
            <Box sx={{ color: 'error.main' }}>
              {(error as Error)?.message || tErrors('loadingError', { resource: tTemp('title') })}
            </Box>
          )}

          {!isLoading && !isError && (
            <Box sx={{ overflow: 'hidden', bgcolor: 'transparent', flex: 1, minHeight: 0, display: 'flex' }}>
              {/* Left list */}
              <Box
                sx={{
                  width: { xs: '100%', sm: '300px' },
                  minHeight: 0,
                  display: 'flex',
                  flexDirection: 'column',
                  overflow: 'hidden',
                  borderRight: { xs: 'none', sm: '1px solid' },
                  borderColor: 'grey.200',
                }}
              >
                <Box sx={{ p: '24px', flexShrink: 0 }}>
                  <SearchInput
                    value={searchInput}
                    onChange={(e) => setSearchInput(e.target.value)}
                    placeholder={tTemp('searchPlaceholder')}
                  />
                  <Box sx={{ mt: '16px' }}>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(Array.from(e.target.selectedOptions).map(o => o.value))}
                      style={{
                        width: '100px',
                        padding: '6px',
                        fontSize: 12,
                        borderRadius: 6,
                        border: '1px solid #e5e7eb',
                        backgroundColor: 'white',
                        color: '#1f2937',
                        cursor: 'pointer',
                      }}
                    >
                      <option value="all">{toStatus('all')}</option>
                      <option value="pending">{toStatus('pending')}</option>
                      <option value="accepted">{toStatus('accepted')}</option>
                      <option value="rejected">{toStatus('rejected')}</option>
                    </select>
                  </Box>
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(tempOrdersRes || []).map((o) => {
                    const isActive = o.id === selectedTempOrderId
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
                          '&:hover': {
                            bgcolor: isActive ? 'grey.100' : 'grey.50',
                          },
                        }}
                        onClick={() => setSelectedTempOrderId(o.id)}
                      >
                        <Box sx={{ flexDirection: 'column', gap: '4px', display: 'flex' }}>
                          <Box sx={{ justifyContent: 'space-between', alignItems: 'center', display: 'flex' }}>
                            <Box sx={{ fontWeight: 700, fontSize: '14px' }}>
                              #{o.id}
                            </Box>
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
                              {toStatus(o.status) || o.status}
                            </Box>
                          </Box>
                          <Box sx={{ fontSize: '12px', color: 'grey.500' }}>
                            {o.customer_name}
                          </Box>
                          <Box sx={{ fontSize: '12px', fontWeight: 500 }}>
                            Rp {formatPrice(o.total_price)}
                          </Box>
                        </Box>
                      </Box>
                    )
                  })}
                  {(tempOrdersRes || []).length === 0 && (
                    <Box sx={{ p: '16px', color: 'grey.500', textAlign: 'center' }}>
                      {tTemp('noOrders')}
                    </Box>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bgcolor: 'grey.50' }}>
                {selectedTempOrder ? (
                  <Box sx={{ maxWidth: 880, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
                    <Box sx={{ alignItems: 'center', gap: '16px', mb: '16px', flexWrap: 'wrap', display: 'flex' }}>
                      <Typography component="h2" sx={{ fontSize: '18px' }}>
                        {tTemp('orderNumber', { id: selectedTempOrder.id })}
                      </Typography>
                      <Box
                        sx={{
                          px: '8px',
                          py: '4px',
                          borderRadius: '4px',
                          fontSize: '12px',
                          fontWeight: 500,
                          bgcolor: getStatusStyle(selectedTempOrder.status).bg,
                          color: getStatusStyle(selectedTempOrder.status).color,
                          textTransform: 'capitalize',
                        }}
                      >
                        {toStatus(selectedTempOrder.status) || selectedTempOrder.status}
                      </Box>
                      <Box sx={{ ml: 'auto', gap: '8px', display: 'flex' }}>
                        <Button
                          variant={selectedTempOrder.status === 'rejected' ? 'outlined' : 'contained'}
                          disableElevation
                          onClick={handleAccept}
                          disabled={acceptLoading || rejectLoading || selectedTempOrder.status !== 'pending'}
                          sx={{
                            ...(selectedTempOrder.status === 'rejected'
                              ? { bgcolor: 'grey.50', color: 'grey.800', border: '1px solid', borderColor: 'grey.200' }
                              : { bgcolor: 'primary.main', color: 'white' }),
                            '&:disabled': { opacity: 0.7 },
                          }}
                        >
                          {acceptLoading ? tTemp('accepting') : tTemp('accept')}
                        </Button>
                        <Button
                          variant={selectedTempOrder.status === 'rejected' ? 'contained' : 'outlined'}
                          disableElevation
                          onClick={handleReject}
                          disabled={rejectLoading || selectedTempOrder.status !== 'pending'}
                          sx={{
                            ...(selectedTempOrder.status === 'rejected'
                              ? { bgcolor: 'primary.main', color: 'white', border: 'none' }
                              : { bgcolor: 'grey.50', color: 'grey.800', border: '1px solid', borderColor: 'grey.200' }),
                            '&:disabled': { opacity: 0.7 },
                          }}
                        >
                          {rejectLoading ? tTemp('rejecting') : tTemp('reject')}
                        </Button>
                      </Box>
                    </Box>
                    {(acceptError || acceptSuccess || rejectError || rejectSuccess) && (
                      <Box
                        sx={{
                          mb: '16px',
                          p: '8px',
                          borderRadius: '8px',
                          bgcolor: acceptError || rejectError ? '#FFEBEE' : '#E8F5E9',
                          color: acceptError || rejectError ? '#C62828' : '#2E7D32',
                          fontSize: '14px',
                        }}
                      >
                        {acceptError || rejectError || (acceptSuccess ? tTemp('acceptSuccess') : '') || (rejectSuccess ? tTemp('rejectSuccess') : '')}
                      </Box>
                    )}

                    {/* Temp order info card */}
                    <Paper
                      sx={{
                        p: '24px',
                        mb: '24px',
                        borderRadius: '12px',
                        boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'white',
                      }}
                    >
                      <Box sx={{ flexWrap: 'wrap', gap: { xs: '24px', sm: '32px' }, display: 'flex' }}>
                        <Box sx={{ minWidth: 140 }}>
                          <Box
                            sx={{
                              color: 'grey.500',
                              fontSize: '14px',
                              fontWeight: 700,
                              mb: '4px',
                              display: 'block',
                            }}
                          >
                            {t('customer')}
                          </Box>
                          <Box sx={{ fontSize: '14px', fontWeight: 500 }}>
                            {selectedTempOrder.customer_name}
                          </Box>
                        </Box>
                        <Box sx={{ minWidth: 140 }}>
                          <Box
                            sx={{
                              color: 'grey.500',
                              fontSize: '14px',
                              fontWeight: 700,
                              mb: '4px',
                              display: 'block',
                            }}
                          >
                            {tTemp('phone')}
                          </Box>
                          <Box sx={{ fontSize: '14px' }}>{selectedTempOrder.customer_phone}</Box>
                        </Box>
                        <Box sx={{ minWidth: 140 }}>
                          <Box
                            sx={{
                              color: 'grey.500',
                              fontSize: '14px',
                              fontWeight: 700,
                              mb: '4px',
                              display: 'block',
                            }}
                          >
                            {to('created')}
                          </Box>
                          <Box sx={{ fontSize: '14px' }}>
                            {formatDate(selectedTempOrder.created_at)}
                          </Box>
                        </Box>
                      </Box>
                    </Paper>

                    {/* Order items (read-only) */}
                    <Paper
                      sx={{
                        borderRadius: '12px',
                        boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'white',
                        overflow: 'hidden',
                      }}
                    >
                      <Box sx={{ p: '8px', borderBottom: '1px solid', borderColor: 'grey.200', bgcolor: 'grey.50' }}>
                        <Typography component="h3" sx={{ fontSize: '16px', fontWeight: 600 }}>
                          {t('items')}
                        </Typography>
                      </Box>
                      {selectedTempOrder.order_items && selectedTempOrder.order_items.length > 0 ? (
                        <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                          <Box component="thead">
                            <Box component="tr" sx={{ bgcolor: 'grey.50' }}>
                              <Box
                                component="th"
                                sx={{
                                  p: '16px',
                                  textAlign: 'left',
                                  fontSize: '12px',
                                  fontWeight: 600,
                                  color: 'grey.500',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {t('product')}
                              </Box>
                              <Box
                                component="th"
                                sx={{
                                  p: '16px',
                                  textAlign: 'right',
                                  fontSize: '12px',
                                  fontWeight: 600,
                                  color: 'grey.500',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {t('price')}
                              </Box>
                              <Box
                                component="th"
                                sx={{
                                  p: '16px',
                                  textAlign: 'right',
                                  fontSize: '12px',
                                  fontWeight: 600,
                                  color: 'grey.500',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {t('quantity')}
                              </Box>
                              <Box
                                component="th"
                                sx={{
                                  p: '16px',
                                  textAlign: 'right',
                                  fontSize: '12px',
                                  fontWeight: 600,
                                  color: 'grey.500',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {to('subtotal')}
                              </Box>
                            </Box>
                          </Box>
                          <Box component="tbody">
                            {selectedTempOrder.order_items.map((item) => (
                              <Box
                                component="tr"
                                key={item.id}
                                sx={{
                                  borderTop: '1px solid',
                                  borderColor: 'grey.200',
                                  '&:hover': { bgcolor: 'grey.50' },
                                }}
                              >
                                <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px' }}>
                                  {item.product_name}
                                </Box>
                                <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontSize: '14px' }}>
                                  Rp {formatPrice(item.price)}
                                </Box>
                                <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontSize: '14px' }}>
                                  {item.qty}
                                </Box>
                                <Box
                                  component="td"
                                  sx={{
                                    py: '8px',
                                    px: '16px',
                                    textAlign: 'right',
                                    fontSize: '14px',
                                    fontWeight: 500,
                                  }}
                                >
                                  Rp {formatPrice(item.price * item.qty)}
                                </Box>
                              </Box>
                            ))}
                          </Box>
                          <Box component="tfoot">
                            <Box
                              component="tr"
                              sx={{
                                borderTop: '2px solid',
                                borderColor: 'grey.200',
                                bgcolor: 'grey.50',
                              }}
                            >
                              <Box
                                component="td"
                                sx={{
                                  py: '8px',
                                  px: '16px',
                                  textAlign: 'right',
                                  fontWeight: 700,
                                  fontSize: '16px',
                                }}
                                {...({ colSpan: 3 } as object)}
                              >
                                {t('total')}
                              </Box>
                              <Box
                                component="td"
                                sx={{
                                  py: '8px',
                                  px: '16px',
                                  textAlign: 'right',
                                  fontWeight: 700,
                                  fontSize: '16px',
                                  color: 'primary.main',
                                }}
                              >
                                Rp {formatPrice(selectedTempOrder.total_price)}
                              </Box>
                            </Box>
                          </Box>
                        </Box>
                      ) : (
                        <Box sx={{ p: '32px', textAlign: 'center', color: 'grey.500' }}>
                          <Box sx={{ fontSize: '16px' }}>{to('noItems')}</Box>
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
                    <ClipboardList size={40} opacity={0.4} />
                    <Box sx={{ fontSize: '16px' }}>{tTemp('selectOrder')}</Box>
                    <Box sx={{ fontSize: '14px' }}>{to('chooseFromList')}</Box>
                  </Box>
                )}
              </Box>
            </Box>
          )}
        </Box>
      </Container>

      {/* Active order conflict dialog */}
      {showActiveOrderConflictDialog && (
        <Box
          sx={{
            position: 'fixed',
            inset: 0,
            bgcolor: 'rgba(0,0,0,0.4)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            p: '16px',
            zIndex: 1000,
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowActiveOrderConflictDialog(false)
          }}
        >
          <Paper sx={{ width: { xs: '100%', sm: '540px' }, p: '24px' }} onClick={(e) => e.stopPropagation()}>
            <Typography component="h3" sx={{ mb: '8px' }}>
              {to('duplicateOrderTitle')}
            </Typography>
            <Box sx={{ mb: '24px', color: 'grey.500', display: 'block' }}>
              {to('duplicateOrderMessageInline')}
            </Box>
            <Box sx={{ gap: '8px', justifyContent: 'flex-end', display: 'flex' }}>
              <Button variant="contained" disableElevation type="button" onClick={() => setShowActiveOrderConflictDialog(false)}>
                {t('cancel')}
              </Button>
              <Button variant="contained" disableElevation type="button" onClick={handleConfirmMergeIntoActiveOrder} disabled={acceptLoading}>
                {acceptLoading ? tTemp('accepting') : tTemp('confirmMerge')}
              </Button>
            </Box>
          </Paper>
        </Box>
      )}
    </Layout>
  )
}
