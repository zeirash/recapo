"use client"

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { Box, Button, Card, Container, Flex, Heading, Input, Label, Select, Text } from 'theme-ui'
import Layout from '@/components/Layout'
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
  const queryClient = useQueryClient()
  const [isCreateFormOpen, setIsCreateFormOpen] = useState(false)
  const [isAddItemFormOpen, setIsAddItemFormOpen] = useState(false)
  const [createForm, setCreateForm] = useState<CreateOrderForm>(emptyCreateForm)
  const [addItemForm, setAddItemForm] = useState<AddItemForm>(emptyAddItemForm)
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null)

  // Fetch orders
  const { data: ordersRes, isLoading, isError, error } = useQuery(
    ['orders'],
    async () => {
      const res = await api.getOrders()
      if (!res.success) throw new Error(res.message || 'Failed to fetch orders')
      return res.data as Order[]
    }
  )

  // Fetch selected order details (includes order items)
  const { data: selectedOrderDetails } = useQuery(
    ['order', selectedOrderId],
    async () => {
      if (!selectedOrderId) return null
      const res = await api.getOrder(selectedOrderId)
      if (!res.success) throw new Error(res.message || 'Failed to fetch order details')
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
      if (!res.success) throw new Error(res.message || 'Failed to create order')
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
      if (!res.success) throw new Error(res.message || 'Failed to delete order')
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
      if (!res.success) throw new Error(res.message || 'Failed to update order status')
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
      if (!res.success) throw new Error(res.message || 'Failed to add item')

      // Update order total price
      const currentTotal = selectedOrder?.total_price || 0
      const newTotal = currentTotal + (itemPrice * payload.qty)
      const updateRes = await api.updateOrder(orderId, { total_price: newTotal })
      if (!updateRes.success) throw new Error(updateRes.message || 'Failed to update order total')

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
      if (!res.success) throw new Error(res.message || 'Failed to delete item')

      // Update order total price
      const currentTotal = selectedOrder?.total_price || 0
      const newTotal = currentTotal - (itemPrice * itemQty)
      const updateRes = await api.updateOrder(orderId, { total_price: Math.max(0, newTotal) })
      if (!updateRes.success) throw new Error(updateRes.message || 'Failed to update order total')

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
      if (!res.success) throw new Error(res.message || 'Failed to update item')

      // Update order total price
      const currentTotal = selectedOrder?.total_price || 0
      const priceDiff = itemPrice * (newQty - oldQty)
      const newTotal = currentTotal + priceDiff
      const updateRes = await api.updateOrder(orderId, { total_price: Math.max(0, newTotal) })
      if (!updateRes.success) throw new Error(updateRes.message || 'Failed to update order total')

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
      console.log('product', product)
      const itemPrice = product?.price || 0
      console.log('itemPrice', itemPrice)

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
          {isLoading && <Text>Loading...</Text>}
          {isError && (
            <Text sx={{ color: 'error' }}>{(error as Error)?.message || 'Error loading orders'}</Text>
          )}

          {!isLoading && !isError && (
            <Flex sx={{ overflow: 'hidden', bg: 'transparent', flex: 1, minHeight: 0 }}>
              {/* Left list */}
              <Box sx={{ width: ['100%', '300px'], minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', borderRight: ['none', '1px solid'], borderColor: 'border' }}>
                <Box sx={{ p: 4, flexShrink: 0 }}>
                  <Button
                    onClick={openCreateForm}
                    sx={{ width: '100%', whiteSpace: 'nowrap', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}
                  >
                    + Order
                  </Button>
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
                    <Text sx={{ p: 3, color: 'text.secondary', textAlign: 'center' }}>No orders</Text>
                  )}
                </Box>
              </Box>

              {/* Right detail */}
              <Box sx={{ flex: 1, minHeight: 0, p: 4, bg: 'white', overflowY: 'auto' }}>
                {selectedOrder ? (
                  <>
                    <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
                      <Flex sx={{ alignItems: 'center', gap: 3 }}>
                        <Heading as="h2" sx={{ fontSize: 3 }}>Order #{selectedOrder.id}</Heading>
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
                      <Flex sx={{ gap: 2 }}>
                        <Button
                          variant="secondary"
                          onClick={() => {
                            if (confirm('Delete this order?')) deleteMutation.mutate(selectedOrder.id)
                          }}
                        >
                          Delete
                        </Button>
                      </Flex>
                    </Flex>

                    {/* Order Info */}
                    <Card sx={{ p: 3, mb: 3 }}>
                      <Flex sx={{ gap: 4, flexWrap: 'wrap' }}>
                        <Box>
                          <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Customer</Text>
                          <Text sx={{ fontWeight: 'medium' }}>{selectedOrder.customer_name}</Text>
                        </Box>
                        <Box>
                          <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Total</Text>
                          <Text sx={{ fontWeight: 'medium' }}>{formatPrice(selectedOrder.total_price)}</Text>
                        </Box>
                        <Box>
                          <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Created</Text>
                          <Text>{formatDate(selectedOrder.created_at)}</Text>
                        </Box>
                        <Box>
                          <Text sx={{ color: 'text.secondary', fontSize: 0 }}>Status</Text>
                          <Select
                            value={selectedOrder.status}
                            onChange={(e) => updateStatusMutation.mutate({ id: selectedOrder.id, status: e.target.value })}
                            sx={{ py: 1, px: 2, fontSize: 0, mt: 1 }}
                          >
                            <option value="created">Created</option>
                            <option value="pending">Pending</option>
                            <option value="processing">Processing</option>
                            <option value="completed">Completed</option>
                            <option value="cancelled">Cancelled</option>
                          </Select>
                        </Box>
                      </Flex>
                    </Card>

                    {/* Order Items */}
                    <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
                      <Heading as="h3" sx={{ fontSize: 2 }}>Items</Heading>
                      <Button variant="secondary" onClick={openAddItemForm} sx={{ fontSize: 0, py: 1, px: 2 }}>
                        + Add Item
                      </Button>
                    </Flex>
                    <Card sx={{ p: 0, overflow: 'hidden' }}>
                      {selectedOrder.order_items && selectedOrder.order_items.length > 0 ? (
                        <Box as="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                          <Box as="thead" sx={{ bg: 'background' }}>
                            <Box as="tr">
                              <Box as="th" sx={{ p: 2, textAlign: 'left', fontSize: 0, fontWeight: 'medium', color: 'text.secondary' }}>Product</Box>
                              <Box as="th" sx={{ p: 2, textAlign: 'right', fontSize: 0, fontWeight: 'medium', color: 'text.secondary' }}>Price</Box>
                              <Box as="th" sx={{ p: 2, textAlign: 'right', fontSize: 0, fontWeight: 'medium', color: 'text.secondary' }}>Qty</Box>
                              <Box as="th" sx={{ p: 2, textAlign: 'right', fontSize: 0, fontWeight: 'medium', color: 'text.secondary' }}>Subtotal</Box>
                              <Box as="th" sx={{ p: 2, width: '50px' }}></Box>
                            </Box>
                          </Box>
                          <Box as="tbody">
                            {selectedOrder.order_items.map((item) => (
                              <Box as="tr" key={item.id} sx={{ borderTop: '1px solid', borderColor: 'border' }}>
                                <Box as="td" sx={{ p: 2, fontSize: 1 }}>{item.product_name}</Box>
                                <Box as="td" sx={{ p: 2, textAlign: 'right', fontSize: 1 }}>{formatPrice(item.price)}</Box>
                                <Box as="td" sx={{ p: 2, textAlign: 'right', fontSize: 1 }}>
                                  <Input
                                    type="number"
                                    min="1"
                                    defaultValue={item.qty}
                                    sx={{ width: '60px', textAlign: 'right', py: 1, px: 2, fontSize: 1 }}
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
                                <Box as="td" sx={{ p: 2, textAlign: 'right', fontSize: 1, fontWeight: 'medium' }}>{formatPrice(item.price * item.qty)}</Box>
                                <Box as="td" sx={{ p: 2, textAlign: 'center' }}>
                                  <Button
                                    variant="secondary"
                                    sx={{ fontSize: 0, py: 1, px: 2, bg: 'transparent', color: 'error', '&:hover': { bg: 'errorLight' } }}
                                    onClick={() => {
                                      if (confirm('Remove this item?')) {
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
                          <Box as="tfoot" sx={{ bg: 'background' }}>
                            <Box as="tr" sx={{ borderTop: '1px solid', borderColor: 'border' }}>
                              <td colSpan={3} style={{ padding: '8px', textAlign: 'right', fontWeight: 'bold', fontSize: '14px' }}>Total</td>
                              <Box as="td" sx={{ p: 2, textAlign: 'right', fontWeight: 'bold', fontSize: 1 }}>{formatPrice(selectedOrder.total_price)}</Box>
                              <Box as="td"></Box>
                            </Box>
                          </Box>
                        </Box>
                      ) : (
                        <Box sx={{ p: 4, textAlign: 'center', color: 'text.secondary' }}>
                          <Text>No items in this order</Text>
                          <Button variant="secondary" onClick={openAddItemForm} sx={{ mt: 2 }}>
                            Add First Item
                          </Button>
                        </Box>
                      )}
                    </Card>
                  </>
                ) : (
                  <Flex sx={{ height: '100%', alignItems: 'center', justifyContent: 'center', color: 'text.secondary' }}>
                    <Text>Select an order to view details</Text>
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
              <Heading as="h3" sx={{ mb: 3 }}>New Order</Heading>
              <Box as="form" onSubmit={submitCreateForm}>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="customer">Customer</Label>
                  <Select
                    id="customer"
                    value={createForm.customer_id || ''}
                    onChange={(e) => setCreateForm({ ...createForm, customer_id: Number(e.target.value) || null })}
                    required
                  >
                    <option value="">Select a customer</option>
                    {(customersRes || []).map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </Select>
                </Box>
                <Flex sx={{ gap: 2, justifyContent: 'flex-end' }}>
                  <Button type="button" variant="secondary" onClick={closeCreateForm}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createMutation.isLoading || !createForm.customer_id}>
                    Create Order
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
              <Heading as="h3" sx={{ mb: 3 }}>Add Item</Heading>
              <Box as="form" onSubmit={submitAddItemForm}>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="product">Product</Label>
                  <Select
                    id="product"
                    value={addItemForm.product_id || ''}
                    onChange={(e) => setAddItemForm({ ...addItemForm, product_id: Number(e.target.value) || null })}
                    required
                  >
                    <option value="">Select a product</option>
                    {(productsRes || []).map((p) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </Select>
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Label htmlFor="qty">Quantity</Label>
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
                    Cancel
                  </Button>
                  <Button type="submit" disabled={addItemMutation.isLoading || !addItemForm.product_id}>
                    Add Item
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
