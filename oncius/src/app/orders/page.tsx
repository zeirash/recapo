"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Select, Text } from 'theme-ui'
import Layout from '@/components/Layout'
import SearchInput from '@/components/SearchInput'
import { api } from '@/utils/api'

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
}

type AddItemForm = {
  product_id: number | null
  qty: number
}

const emptyCreateForm: CreateOrderForm = { customer_id: null }
const emptyAddItemForm: AddItemForm = { product_id: null, qty: 1 }

const statusColors: Record<string, { bg: string; color: string }> = {
  created: { bg: '#E3F2FD', color: '#1565C0' },
  pending: { bg: '#FFF3E0', color: '#E65100' },
  processing: { bg: '#F3E5F5', color: '#7B1FA2' },
  completed: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FFEBEE', color: '#C62828' },
}

export default function OrdersPage() {
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

  // Debounce search: only trigger API after user stops typing for 300ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  // Fetch orders
  const { data: ordersRes, isLoading, isError, error } = useQuery(
    ['orders', debouncedSearch],
    async () => {
      const res = await api.getOrders(debouncedSearch || undefined)
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

  // Fetch customers for create form
  const { data: customersRes } = useQuery(
    ['customers'],
    async () => {
      const res = await api.getCustomers()
      if (!res.success) throw new Error(res.message || 'Failed to fetch customers')
      return res.data as Customer[]
    },
    { enabled: isCreateFormOpen }
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
    async (payload: { customer_id: number }) => {
      const res = await api.createOrder(payload)
      if (!res.success) throw new Error(res.message || to('createFailed'))
      return res
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['orders'])
        closeCreateForm()
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
      createMutation.mutate({ customer_id: createForm.customer_id })
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
      <Container sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Flex sx={{ height: '100%', minHeight: 0, flex: 1, flexDirection: 'column', overflow: 'hidden' }}>
          {isLoading && <Text>{t('loading')}</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>{(error as Error)?.message || tErrors('loadingError', { resource: to('title') })}</Text>
          )}

          {!isLoading && !isError && (
            <Flex sx={{ overflow: 'hidden', bg: 'transparent', flex: 1, minHeight: 0 }}>
              {/* Left list */}
              <Box sx={{ width: ['100%', '300px'], minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRight: ['none', '1px solid'], borderColor: 'border' }}>
                <Box sx={{ p: 4, flexShrink: 0 }}>
                  <Flex sx={{ gap: 2, alignItems: 'center' }}>
                    <SearchInput
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      placeholder={to('searchPlaceholder')}
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
                      title={to('addOrder')}
                    >
                      +
                    </Button>
                  </Flex>
                </Box>
                <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
                  {(ordersRes || []).map((o) => {
                    const isActive = o.id === selectedOrderId
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
                          '&:hover': { bg: isActive ? 'backgroundLight' : 'background.secondary' },
                        }}
                        onClick={() => setSelectedOrderId(o.id)}
                      >
                        <Flex sx={{ flexDirection: 'column', gap: 1 }}>
                          <Flex sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
                            <Text sx={{ fontWeight: 'bold', fontSize: 1 }}>#{o.id}</Text>
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
                              {o.status}
                            </Box>
                          </Flex>
                          <Text sx={{ fontSize: 0, color: 'text.secondary' }}>{o.customer_name}</Text>
                          <Text sx={{ fontSize: 0, fontWeight: 'medium' }}>{formatPrice(o.total_price)}</Text>
                        </Flex>
                      </Box>
                    )
                  })}
                  {(ordersRes || []).length === 0 && (
                    <Text sx={{ p: 3, color: 'text.secondary', textAlign: 'center' }}>{to('noOrders')}</Text>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', bg: 'background.secondary' }}>
                {selectedOrder ? (
                  <Box sx={{ maxWidth: 880, mx: 'auto', p: [4, 5] }}>
                    <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
                      <Flex sx={{ alignItems: 'center', gap: 3 }}>
                        <Heading as="h2" sx={{ fontSize: 3 }}>{to('orderNumber', { id: selectedOrder.id })}</Heading>
                        <Box
                          sx={{
                            px: 2,
                            py: '4px',
                            borderRadius: 'small',
                            fontSize: 0,
                            fontWeight: 'medium',
                            bg: getStatusStyle(selectedOrder.status).bg,
                            color: getStatusStyle(selectedOrder.status).color,
                            textTransform: 'capitalize',
                          }}
                        >
                          {selectedOrder.status}
                        </Box>
                      </Flex>
                      <Button
                        variant="secondary"
                        onClick={() => {
                          if (confirm(to('deleteConfirm'))) deleteMutation.mutate(selectedOrder.id)
                        }}
                      >
                        {t('delete')}
                      </Button>
                    </Flex>

                    {/* Order info card */}
                    <Card
                      sx={{
                        p: 4,
                        mb: 4,
                        borderRadius: 'large',
                        boxShadow: 'small',
                        border: '1px solid',
                        borderColor: 'border',
                        bg: 'white',
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: 'medium' },
                      }}
                    >
                      <Flex sx={{ flexWrap: 'wrap', gap: [4, 5] }}>
                        <Box sx={{ minWidth: 140 }}>
                          <Text sx={{ color: 'text.secondary', fontSize: 1, fontWeight: 700, mb: 1, display: 'block' }}>Customer</Text>
                          <Text sx={{ fontSize: 1, fontWeight: 'medium' }}>{selectedOrder.customer_name}</Text>
                        </Box>
                        <Box sx={{ minWidth: 140 }}>
                          <Text sx={{ color: 'text.secondary', fontSize: 1, fontWeight: 700, mb: 1, display: 'block' }}>{to('created')}</Text>
                          <Text sx={{ fontSize: 1 }}>{formatDate(selectedOrder.created_at)}</Text>
                        </Box>
                        <Box sx={{ minWidth: 140, ml: 'auto' }}>
                          <Text sx={{ color: 'text.secondary', fontSize: 1, fontWeight: 700, mb: 1, display: 'block' }}>{t('status')}</Text>
                          <Select
                            value={selectedOrder.status}
                            onChange={(e) => updateStatusMutation.mutate({ id: selectedOrder.id, status: e.target.value })}
                            sx={{
                              py: 1,
                              px: 2,
                              fontSize: 1,
                              borderRadius: 'medium',
                              border: '1px solid',
                              borderColor: 'border',
                              fontWeight: 'medium',
                              cursor: 'pointer',
                            }}
                          >
                            <option value="created">{toStatus('created')}</option>
                            <option value="pending">{toStatus('pending')}</option>
                            <option value="processing">{toStatus('processing')}</option>
                            <option value="completed">{toStatus('completed')}</option>
                            <option value="cancelled">{toStatus('cancelled')}</option>
                          </Select>
                        </Box>
                      </Flex>
                    </Card>

                    {/* Order items */}
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
                        transition: 'box-shadow 0.2s ease',
                        '&:hover': { boxShadow: 'medium' },
                      }}
                    >
                      <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', p: 2, borderBottom: '1px solid', borderColor: 'border', bg: 'background.secondary' }}>
                        <Heading as="h3" sx={{ fontSize: 2, fontWeight: 600 }}>{t('items')}</Heading>
                        <Button variant="secondary" onClick={openAddItemForm} sx={{ fontSize: 0, py: 1, px: 2 }}>
                          {to('addItem')}
                        </Button>
                      </Flex>
                      {selectedOrder.order_items && selectedOrder.order_items.length > 0 ? (
                        <Box as="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                          <Box as="thead">
                            <Box as="tr" sx={{ bg: 'background.secondary' }}>
                              <Box as="th" sx={{ p: 3, textAlign: 'left', fontSize: 0, fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('product')}</Box>
                              <Box as="th" sx={{ p: 3, textAlign: 'right', fontSize: 0, fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('price')}</Box>
                              <Box as="th" sx={{ p: 3, textAlign: 'right', fontSize: 0, fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('quantity')}</Box>
                              <Box as="th" sx={{ p: 3, textAlign: 'right', fontSize: 0, fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{to('subtotal')}</Box>
                              <Box as="th" sx={{ p: 3, width: '56px' }}></Box>
                            </Box>
                          </Box>
                          <Box as="tbody">
                            {selectedOrder.order_items.map((item) => (
                              <Box as="tr" key={item.id} sx={{ borderTop: '1px solid', borderColor: 'border', '&:hover': { bg: 'background.secondary' } }}>
                                <Box as="td" sx={{ py: 2, px: 3, fontSize: 1 }}>{item.product_name}</Box>
                                <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontSize: 1 }}>{formatPrice(item.price)}</Box>
                                <Box as="td" sx={{ py: 2, pl: 3, pr: 2 }}>
                                  <Flex sx={{ justifyContent: 'flex-end' }}>
                                  <Input
                                    type="number"
                                    min="1"
                                    defaultValue={item.qty}
                                    sx={{ width: '64px', textAlign: 'right', py: 2, px: 2, fontSize: 1, borderRadius: 'medium' }}
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
                                  </Flex>
                                </Box>
                                <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontSize: 1, fontWeight: 'medium' }}>{formatPrice(item.price * item.qty)}</Box>
                                <Box as="td" sx={{ py: 2, px: 3, textAlign: 'center' }}>
                                  <Button
                                    variant="secondary"
                                    sx={{ fontSize: 1, py: 1, px: 2, bg: 'transparent', color: 'error', borderRadius: 'medium', '&:hover': { bg: '#fef2f2' } }}
                                    onClick={() => {
                                      if (confirm(to('removeItemConfirm'))) {
                                        deleteItemMutation.mutate({ orderId: selectedOrder.id, itemId: item.id, itemPrice: item.price, itemQty: item.qty })
                                      }
                                    }}
                                  >
                                    ×
                                  </Button>
                                </Box>
                              </Box>
                            ))}
                          </Box>
                          <Box as="tfoot">
                            <Box as="tr" sx={{ borderTop: '2px solid', borderColor: 'border', bg: 'background.secondary' }}>
                              <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontWeight: 700, fontSize: 2 }} {...({ colSpan: 3 } as object)}>{t('total')}</Box>
                              <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontWeight: 700, fontSize: 2, color: 'primary' }}>{formatPrice(selectedOrder.total_price)}</Box>
                              <Box as="td" sx={{ py: 2, px: 3 }}></Box>
                            </Box>
                          </Box>
                        </Box>
                      ) : (
                        <Box sx={{ p: 5, textAlign: 'center', color: 'text.secondary' }}>
                          <Text sx={{ fontSize: 2, display: 'block', mb: 2 }}>{to('noItems')}</Text>
                          <Button variant="secondary" onClick={openAddItemForm}>
                            {to('addFirstItem')}
                          </Button>
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
                    <Text sx={{ fontSize: 2 }}>{to('selectOrder')}</Text>
                    <Text sx={{ fontSize: 1 }}>{to('chooseFromList')}</Text>
                  </Flex>
                )}
              </Box>
            </Flex>
          )}
        </Flex>

        {/* Create Order Modal */}
        {isCreateFormOpen && (
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
              if (e.target === e.currentTarget) closeCreateForm()
            }}
          >
            <Card sx={{ width: ['100%', '540px'] }}>
              <Heading as="h3" sx={{ mb: 3 }}>{to('newOrder')}</Heading>
              <Box as="form" onSubmit={submitCreateForm}>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="customer">{t('customer')}</Label>
                  <Select
                    id="customer"
                    value={createForm.customer_id || ''}
                    onChange={(e) => setCreateForm({ ...createForm, customer_id: Number(e.target.value) || null })}
                    required
                  >
                    <option value="">{to('selectCustomer')}</option>
                    {(customersRes || []).map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </Select>
                </Box>
                <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
                  <Button type="button" variant="secondary" onClick={closeCreateForm}>
                    {t('cancel')}
                  </Button>
                  <Button type="submit" disabled={createMutation.isLoading || !createForm.customer_id}>
                    {to('createOrder')}
                  </Button>
                </Flex>
              </Box>
            </Card>
          </Box>
        )}

        {/* Add Item Modal */}
        {isAddItemFormOpen && (
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
              if (e.target === e.currentTarget) closeAddItemForm()
            }}
          >
            <Card sx={{ width: ['100%', '540px'] }}>
              <Heading as="h3" sx={{ mb: 3 }}>{to('addItemTitle')}</Heading>
              <Box as="form" onSubmit={submitAddItemForm}>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="product">{t('product')}</Label>
                  <Select
                    id="product"
                    value={addItemForm.product_id || ''}
                    onChange={(e) => setAddItemForm({ ...addItemForm, product_id: Number(e.target.value) || null })}
                    required
                  >
                    <option value="">{to('selectProduct')}</option>
                    {(productsRes || []).map((p) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </Select>
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="qty">{t('quantity')}</Label>
                  <Input
                    id="qty"
                    type="number"
                    min="1"
                    value={addItemForm.qty}
                    onChange={(e) => setAddItemForm({ ...addItemForm, qty: Number(e.target.value) || 1 })}
                    required
                  />
                </Box>
                <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
                  <Button type="button" variant="secondary" onClick={closeAddItemForm}>
                    {t('cancel')}
                  </Button>
                  <Button type="submit" disabled={addItemMutation.isLoading || !addItemForm.product_id}>
                    {to('addItemButton')}
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
