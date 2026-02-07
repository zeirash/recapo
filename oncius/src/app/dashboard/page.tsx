"use client"

import { useMemo } from 'react'
import { useQuery } from 'react-query'
import Link from 'next/link'
import { Box, Button, Card, Flex, Heading, Text } from 'theme-ui'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'
import { api } from '@/utils/api'

type Order = {
  id: number
  customer_name: string
  total_price: number
  status: string
  created_at: string
}

const statusColors: Record<string, { bg: string; color: string }> = {
  created: { bg: '#E3F2FD', color: '#1565C0' },
  pending: { bg: '#FFF3E0', color: '#E65100' },
  processing: { bg: '#F3E5F5', color: '#7B1FA2' },
  completed: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FFEBEE', color: '#C62828' },
}

const StatCard = ({
  label,
  value,
  icon,
}: {
  label: string
  value: number | string
  icon: string
}) => (
  <Card
    sx={{
      p: 4,
      borderRadius: 'large',
      boxShadow: 'small',
      border: '1px solid',
      borderColor: 'border',
      bg: 'white',
      flex: 1,
      minWidth: ['100%', 140],
    }}
  >
    <Flex sx={{ alignItems: 'flex-start', justifyContent: 'space-between' }}>
      <Box>
        <Text sx={{ color: 'text.secondary', fontSize: 1, fontWeight: 600, mb: 1, display: 'block' }}>
          {label}
        </Text>
        <Text sx={{ fontSize: 4, fontWeight: 700 }}>{value}</Text>
      </Box>
      <Box sx={{ fontSize: 4, opacity: 0.6 }}>{icon}</Box>
    </Flex>
  </Card>
)

const getThisMonthRange = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = now.getMonth()
  const dateFrom = `${year}-${String(month + 1).padStart(2, '0')}-01`
  const lastDay = new Date(year, month + 1, 0).getDate()
  const dateTo = `${year}-${String(month + 1).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`
  return { dateFrom, dateTo }
}

const DashboardPage = () => {
  const t = useTranslations()
  const { isAuthenticated } = useAuth()
  const { dateFrom, dateTo } = getThisMonthRange()

  const { data: ordersRes, isLoading: ordersLoading } = useQuery(
    ['orders', 'dashboard', dateFrom, dateTo],
    async () => {
      const res = await api.getOrders({ date_from: dateFrom, date_to: dateTo })
      if (!res.success) throw new Error(res.message)
      return res.data as Order[]
    }
  )

  const { data: customersRes } = useQuery(
    ['customers'],
    async () => {
      const res = await api.getCustomers()
      if (!res.success) throw new Error(res.message)
      return res.data as { id: number }[]
    }
  )

  const { data: productsRes } = useQuery(
    ['products'],
    async () => {
      const res = await api.getProducts()
      if (!res.success) throw new Error(res.message)
      return res.data as { id: number }[]
    }
  )

  const stats = useMemo(() => {
    const orders = ordersRes || []
    const totalRevenue = orders.reduce((sum, o) => sum + (o.total_price || 0), 0)
    return {
      totalOrders: orders.length,
      revenue: totalRevenue,
      customers: (customersRes || []).length,
      products: (productsRes || []).length,
    }
  }, [ordersRes, customersRes, productsRes])

  const recentOrders = useMemo(() => {
    const orders = ordersRes || []
    return [...orders]
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 5)
  }, [ordersRes])

  function formatPrice(price: number) {
    return price.toLocaleString()
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  function getStatusStyle(status: string) {
    return statusColors[status] || { bg: '#F5F5F5', color: '#616161' }
  }

  if (!isAuthenticated) {
    return (
      <Layout>
        <Box sx={{ p: 4, textAlign: 'center' }}>
          <Text>{t('dashboard.loginRequired')}</Text>
        </Box>
      </Layout>
    )
  }

  return (
    <Layout>
      <Box sx={{ maxWidth: 1200, mx: 'auto', p: [4, 5] }}>
        <Heading as="h1" sx={{ mb: 4, fontSize: 3 }}>
          {t('nav.dashboard')}
        </Heading>

        {ordersLoading ? (
          <Text sx={{ color: 'text.secondary' }}>{t('common.loading')}</Text>
        ) : (
          <>
            {/* Stat cards */}
            <Flex
              sx={{
                gap: 3,
                mb: 5,
                flexWrap: 'wrap',
              }}
            >
              <StatCard label="Total Orders (this month)" value={stats.totalOrders} icon="📦" />
              <StatCard label="Revenue (this month)" value={formatPrice(stats.revenue)} icon="💰" />
              <StatCard label="Customers" value={stats.customers} icon="👥" />
              <StatCard label="Products" value={stats.products} icon="🛍️" />
            </Flex>

            {/* Quick links */}
            <Flex sx={{ gap: 2, mb: 5, flexWrap: 'wrap' }}>
              <Link href="/orders" passHref legacyBehavior>
                <Button as="a" variant="secondary">
                  {t('dashboard.newOrder')}
                </Button>
              </Link>
              <Link href="/customers" passHref legacyBehavior>
                <Button as="a" variant="secondary">
                  {t('dashboard.addCustomer')}
                </Button>
              </Link>
              <Link href="/products" passHref legacyBehavior>
                <Button as="a" variant="secondary">
                  {t('dashboard.addProduct')}
                </Button>
              </Link>
            </Flex>

            {/* Recent orders */}
            <Card
              sx={{
                borderRadius: 'large',
                boxShadow: 'small',
                border: '1px solid',
                borderColor: 'border',
                bg: 'white',
                overflow: 'hidden',
              }}
            >
              <Flex
                sx={{
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  p: 4,
                  borderBottom: '1px solid',
                  borderColor: 'border',
                  bg: 'background.secondary',
                }}
              >
                <Heading as="h2" sx={{ fontSize: 2, fontWeight: 600 }}>
                  {t('dashboard.recentOrders')}
                </Heading>
                <Link href="/orders" passHref legacyBehavior>
                  <Button as="a" variant="secondary" sx={{ fontSize: 0, py: 1, px: 2 }}>
                    {t('dashboard.viewAll')}
                  </Button>
                </Link>
              </Flex>
              {recentOrders.length > 0 ? (
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
                        {t('dashboard.order')}
                      </Box>
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
                        {t('common.customer')}
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
                        {t('common.total')}
                      </Box>
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
                        {t('common.status')}
                      </Box>
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
                        {t('dashboard.date')}
                      </Box>
                    </Box>
                  </Box>
                  <Box as="tbody">
                    {recentOrders.map((order) => {
                      const statusStyle = getStatusStyle(order.status)
                      return (
                        <Box
                          as="tr"
                          key={order.id}
                          sx={{
                            borderTop: '1px solid',
                            borderColor: 'border',
                            '&:hover': { bg: 'background.secondary' },
                          }}
                        >
                          <Box as="td" sx={{ py: 2, px: 3, fontSize: 1 }}>
                            <Link
                              href="/orders"
                              style={{ color: 'inherit', fontWeight: 600, textDecoration: 'none' }}
                            >
                              #{order.id}
                            </Link>
                          </Box>
                          <Box as="td" sx={{ py: 2, px: 3, fontSize: 1 }}>
                            {order.customer_name}
                          </Box>
                          <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontSize: 1 }}>
                            {formatPrice(order.total_price)}
                          </Box>
                          <Box as="td" sx={{ py: 2, px: 3 }}>
                            <Box
                              sx={{
                                display: 'inline-block',
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
                              {t(`orderStatus.${order.status}`)}
                            </Box>
                          </Box>
                          <Box as="td" sx={{ py: 2, px: 3, fontSize: 1, color: 'text.secondary' }}>
                            {formatDate(order.created_at)}
                          </Box>
                        </Box>
                      )
                    })}
                  </Box>
                </Box>
              ) : (
                <Box sx={{ p: 5, textAlign: 'center', color: 'text.secondary' }}>
                  <Text sx={{ fontSize: 2, display: 'block', mb: 2 }}>{t('dashboard.noOrdersYet')}</Text>
                  <Link href="/orders" passHref legacyBehavior>
                    <Button as="a">{t('dashboard.createFirstOrder')}</Button>
                  </Link>
                </Box>
              )}
            </Card>
          </>
        )}
      </Box>
    </Layout>
  )
}

export default DashboardPage
