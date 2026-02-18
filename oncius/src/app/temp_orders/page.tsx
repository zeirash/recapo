"use client"

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Card, Container, Flex, Heading, Text } from 'theme-ui'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
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
    ['temp_orders', debouncedSearch],
    async () => {
      const res = await api.getTempOrders(debouncedSearch ? { search: debouncedSearch } : undefined)
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
          queryClient.invalidateQueries(['temp_orders', debouncedSearch]),
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
  }, [selectedTempOrder, tTemp, queryClient, debouncedSearch])

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
          queryClient.invalidateQueries(['temp_orders', debouncedSearch]),
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
  }, [selectedTempOrder, conflictData, tTemp, queryClient, debouncedSearch])

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
          queryClient.invalidateQueries(['temp_orders', debouncedSearch]),
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
  }, [selectedTempOrder, tTemp, queryClient, debouncedSearch])

  return (
    <Layout>
      <Container sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Flex sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden' }}>
          {isLoading && <Text>{t('loading')}</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>
              {(error as Error)?.message || tErrors('loadingError', { resource: tTemp('title') })}
            </Text>
          )}

          {!isLoading && !isError && (
            <Flex sx={{ overflow: 'hidden', bg: 'transparent', flex: 1, minHeight: 0 }}>
              {/* Left list */}
              <Box
                sx={{
                  width: ['100%', '300px'],
                  minHeight: 0,
                  display: 'flex',
                  flexDirection: 'column',
                  overflow: 'hidden',
                  borderRight: ['none', '1px solid'],
                  borderColor: 'border',
                }}
              >
                <Box sx={{ p: 4, flexShrink: 0 }}>
                  <SearchInput
                    value={searchInput}
                    onChange={(e) => setSearchInput(e.target.value)}
                    placeholder={tTemp('searchPlaceholder')}
                  />
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(tempOrdersRes || []).map((o) => {
                    const isActive = o.id === selectedTempOrderId
                    const statusStyle = getStatusStyle(o.status)
                    return (
                      <Box
                        key={o.id}
                        sx={{
                          py: 3,
                          px: 4,
                          cursor: 'pointer',
                          textAlign: 'left',
                          bg: isActive ? 'backgroundLight' : 'transparent',
                          borderRadius: 'medium',
                          '&:hover': {
                            bg: isActive ? 'backgroundLight' : 'background.secondary',
                          },
                        }}
                        onClick={() => setSelectedTempOrderId(o.id)}
                      >
                        <Flex sx={{ flexDirection: 'column', gap: 1 }}>
                          <Flex sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
                            <Text sx={{ fontWeight: 'bold', fontSize: 1 }}>
                              #{o.id}
                            </Text>
                            <Box
                              sx={{
                                px: 2,
                                py: '2px',
                                borderRadius: 'small',
                                fontSize: 0,
                                fontWeight: 'medium',
                                bg: statusStyle.bg,
                                color: statusStyle.color,
                                textTransform: 'capitalize',
                              }}
                            >
                              {toStatus(o.status) || o.status}
                            </Box>
                          </Flex>
                          <Text sx={{ fontSize: 0, color: 'text.secondary' }}>
                            {o.customer_name}
                          </Text>
                          <Text sx={{ fontSize: 0, fontWeight: 'medium' }}>
                            Rp {formatPrice(o.total_price)}
                          </Text>
                        </Flex>
                      </Box>
                    )
                  })}
                  {(tempOrdersRes || []).length === 0 && (
                    <Text sx={{ p: 3, color: 'text.secondary', textAlign: 'center' }}>
                      {tTemp('noOrders')}
                    </Text>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bg: 'background.secondary' }}>
                {selectedTempOrder ? (
                  <Box sx={{ maxWidth: 880, mx: 'auto', p: [4, 5] }}>
                    <Flex sx={{ alignItems: 'center', gap: 3, mb: 3, flexWrap: 'wrap' }}>
                      <Heading as="h2" sx={{ fontSize: 3 }}>
                        {tTemp('orderNumber', { id: selectedTempOrder.id })}
                      </Heading>
                      <Box
                        sx={{
                          px: 2,
                          py: '4px',
                          borderRadius: 'small',
                          fontSize: 0,
                          fontWeight: 'medium',
                          bg: getStatusStyle(selectedTempOrder.status).bg,
                          color: getStatusStyle(selectedTempOrder.status).color,
                          textTransform: 'capitalize',
                        }}
                      >
                        {toStatus(selectedTempOrder.status) || selectedTempOrder.status}
                      </Box>
                      <Flex sx={{ ml: 'auto', gap: 2 }}>
                        <Button
                          variant={selectedTempOrder.status === 'rejected' ? 'secondary' : 'primary'}
                          onClick={handleAccept}
                          disabled={acceptLoading || rejectLoading || selectedTempOrder.status !== 'pending'}
                          sx={{
                            ...(selectedTempOrder.status === 'rejected'
                              ? { bg: 'background.secondary', color: 'text', border: '1px solid', borderColor: 'border' }
                              : { bg: 'primary', color: 'white' }),
                            '&:disabled': { opacity: 0.7 },
                          }}
                        >
                          {acceptLoading ? tTemp('accepting') : tTemp('accept')}
                        </Button>
                        <Button
                          variant={selectedTempOrder.status === 'rejected' ? 'primary' : 'secondary'}
                          onClick={handleReject}
                          disabled={rejectLoading || selectedTempOrder.status !== 'pending'}
                          sx={{
                            ...(selectedTempOrder.status === 'rejected'
                              ? { bg: 'primary', color: 'white', border: 'none' }
                              : { bg: 'background.secondary', color: 'text', border: '1px solid', borderColor: 'border' }),
                            '&:disabled': { opacity: 0.7 },
                          }}
                        >
                          {rejectLoading ? tTemp('rejecting') : tTemp('reject')}
                        </Button>
                      </Flex>
                    </Flex>
                    {(acceptError || acceptSuccess || rejectError || rejectSuccess) && (
                      <Box
                        sx={{
                          mb: 3,
                          p: 2,
                          borderRadius: 'medium',
                          bg: acceptError || rejectError ? '#FFEBEE' : '#E8F5E9',
                          color: acceptError || rejectError ? '#C62828' : '#2E7D32',
                          fontSize: 1,
                        }}
                      >
                        {acceptError || rejectError || (acceptSuccess ? tTemp('acceptSuccess') : '') || (rejectSuccess ? tTemp('rejectSuccess') : '')}
                      </Box>
                    )}

                    {/* Temp order info card */}
                    <Card
                      sx={{
                        p: 4,
                        mb: 4,
                        borderRadius: 'large',
                        boxShadow: 'small',
                        border: '1px solid',
                        borderColor: 'border',
                        bg: 'white',
                      }}
                    >
                      <Flex sx={{ flexWrap: 'wrap', gap: [4, 5] }}>
                        <Box sx={{ minWidth: 140 }}>
                          <Text
                            sx={{
                              color: 'text.secondary',
                              fontSize: 1,
                              fontWeight: 700,
                              mb: 1,
                              display: 'block',
                            }}
                          >
                            {t('customer')}
                          </Text>
                          <Text sx={{ fontSize: 1, fontWeight: 'medium' }}>
                            {selectedTempOrder.customer_name}
                          </Text>
                        </Box>
                        <Box sx={{ minWidth: 140 }}>
                          <Text
                            sx={{
                              color: 'text.secondary',
                              fontSize: 1,
                              fontWeight: 700,
                              mb: 1,
                              display: 'block',
                            }}
                          >
                            {tTemp('phone')}
                          </Text>
                          <Text sx={{ fontSize: 1 }}>{selectedTempOrder.customer_phone}</Text>
                        </Box>
                        <Box sx={{ minWidth: 140 }}>
                          <Text
                            sx={{
                              color: 'text.secondary',
                              fontSize: 1,
                              fontWeight: 700,
                              mb: 1,
                              display: 'block',
                            }}
                          >
                            {to('created')}
                          </Text>
                          <Text sx={{ fontSize: 1 }}>
                            {formatDate(selectedTempOrder.created_at)}
                          </Text>
                        </Box>
                      </Flex>
                    </Card>

                    {/* Order items (read-only) */}
                    <Card
                      sx={{
                        py: 1,
                        px: 3,
                        borderRadius: 'large',
                        boxShadow: 'small',
                        border: '1px solid',
                        borderColor: 'border',
                        bg: 'white',
                        overflow: 'hidden',
                      }}
                    >
                      <Box sx={{ p: 2, borderBottom: '1px solid', borderColor: 'border', bg: 'background.secondary' }}>
                        <Heading as="h3" sx={{ fontSize: 2, fontWeight: 600 }}>
                          {t('items')}
                        </Heading>
                      </Box>
                      {selectedTempOrder.order_items && selectedTempOrder.order_items.length > 0 ? (
                        <Box as="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                          <Box as="thead">
                            <Box as="tr" sx={{ bg: 'background.secondary' }}>
                              <Box
                                as="th"
                                sx={{
                                  p: 3,
                                  textAlign: 'left',
                                  fontSize: 0,
                                  fontWeight: 600,
                                  color: 'text.secondary',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {t('product')}
                              </Box>
                              <Box
                                as="th"
                                sx={{
                                  p: 3,
                                  textAlign: 'right',
                                  fontSize: 0,
                                  fontWeight: 600,
                                  color: 'text.secondary',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {t('price')}
                              </Box>
                              <Box
                                as="th"
                                sx={{
                                  p: 3,
                                  textAlign: 'right',
                                  fontSize: 0,
                                  fontWeight: 600,
                                  color: 'text.secondary',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {t('quantity')}
                              </Box>
                              <Box
                                as="th"
                                sx={{
                                  p: 3,
                                  textAlign: 'right',
                                  fontSize: 0,
                                  fontWeight: 600,
                                  color: 'text.secondary',
                                  textTransform: 'uppercase',
                                  letterSpacing: '0.05em',
                                }}
                              >
                                {to('subtotal')}
                              </Box>
                            </Box>
                          </Box>
                          <Box as="tbody">
                            {selectedTempOrder.order_items.map((item) => (
                              <Box
                                as="tr"
                                key={item.id}
                                sx={{
                                  borderTop: '1px solid',
                                  borderColor: 'border',
                                  '&:hover': { bg: 'background.secondary' },
                                }}
                              >
                                <Box as="td" sx={{ py: 2, px: 3, fontSize: 1 }}>
                                  {item.product_name}
                                </Box>
                                <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontSize: 1 }}>
                                  Rp {formatPrice(item.price)}
                                </Box>
                                <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontSize: 1 }}>
                                  {item.qty}
                                </Box>
                                <Box
                                  as="td"
                                  sx={{
                                    py: 2,
                                    px: 3,
                                    textAlign: 'right',
                                    fontSize: 1,
                                    fontWeight: 'medium',
                                  }}
                                >
                                  Rp {formatPrice(item.price * item.qty)}
                                </Box>
                              </Box>
                            ))}
                          </Box>
                          <Box as="tfoot">
                            <Box
                              as="tr"
                              sx={{
                                borderTop: '2px solid',
                                borderColor: 'border',
                                bg: 'background.secondary',
                              }}
                            >
                              <Box
                                as="td"
                                sx={{
                                  py: 2,
                                  px: 3,
                                  textAlign: 'right',
                                  fontWeight: 700,
                                  fontSize: 2,
                                }}
                                {...({ colSpan: 3 } as object)}
                              >
                                {t('total')}
                              </Box>
                              <Box
                                as="td"
                                sx={{
                                  py: 2,
                                  px: 3,
                                  textAlign: 'right',
                                  fontWeight: 700,
                                  fontSize: 2,
                                  color: 'primary',
                                }}
                              >
                                Rp {formatPrice(selectedTempOrder.total_price)}
                              </Box>
                            </Box>
                          </Box>
                        </Box>
                      ) : (
                        <Box sx={{ p: 5, textAlign: 'center', color: 'text.secondary' }}>
                          <Text sx={{ fontSize: 2 }}>{to('noItems')}</Text>
                        </Box>
                      )}
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
                    <Box sx={{ fontSize: 6, opacity: 0.4 }}>📋</Box>
                    <Text sx={{ fontSize: 2 }}>{tTemp('selectOrder')}</Text>
                    <Text sx={{ fontSize: 1 }}>{to('chooseFromList')}</Text>
                  </Flex>
                )}
              </Box>
            </Flex>
          )}
        </Flex>
      </Container>

      {/* Active order conflict dialog */}
      {showActiveOrderConflictDialog && (
        <Box
          sx={{
            position: 'fixed',
            inset: 0,
            bg: 'rgba(0,0,0,0.4)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            p: 3,
            zIndex: 1000,
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowActiveOrderConflictDialog(false)
          }}
        >
          <Card sx={{ width: ['100%', '540px'], p: 4 }} onClick={(e) => e.stopPropagation()}>
            <Heading as="h3" sx={{ mb: 2 }}>
              {to('duplicateOrderTitle')}
            </Heading>
            <Text sx={{ mb: 4, color: 'text.secondary', display: 'block' }}>
              {to('duplicateOrderMessageInline')}
            </Text>
            <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
              <Button type="button" onClick={() => setShowActiveOrderConflictDialog(false)}>
                {t('cancel')}
              </Button>
              <Button type="button" onClick={handleConfirmMergeIntoActiveOrder} disabled={acceptLoading}>
                {acceptLoading ? tTemp('accepting') : tTemp('confirmMerge')}
              </Button>
            </Flex>
          </Card>
        </Box>
      )}
    </Layout>
  )
}
