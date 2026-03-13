"use client"

import { useMemo } from 'react'
import { useQuery } from 'react-query'
import Link from 'next/link'
import { Box, Button, Paper, Typography } from '@mui/material'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { api } from '@/utils/api'
import { Package, DollarSign, Users, ClipboardList, type LucideIcon } from 'lucide-react'

type Order = {
  id: number
  customer_name: string
  total_price: number
  status: string
  created_at: string
}

const statusColors: Record<string, { bg: string; color: string }> = {
  created: { bg: '#E3F2FD', color: '#1565C0' },
  in_progress: { bg: '#FFF3E0', color: '#E65100' },
  in_delivery: { bg: '#F3E5F5', color: '#7B1FA2' },
  done: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FFEBEE', color: '#C62828' },
}

const StatCard = ({
  label,
  value,
  icon: Icon,
}: {
  label: string
  value: number | string
  icon: LucideIcon
}) => (
  <Paper
    sx={{
      p: '24px',
      borderRadius: '12px',
      boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
      border: '1px solid',
      borderColor: 'grey.200',
      bgcolor: 'white',
      flex: 1,
      minWidth: { xs: '100%', sm: 140 },
    }}
  >
    <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
      <Box>
        <Box sx={{ color: 'grey.500', fontSize: '14px', fontWeight: 600, mb: '4px', display: 'block' }}>
          {label}
        </Box>
        <Box sx={{ fontSize: '20px', fontWeight: 700 }}>{value}</Box>
      </Box>
      <Box sx={{ opacity: 0.6 }}><Icon size={28} /></Box>
    </Box>
  </Paper>
)

const getThisMonthRange = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = now.getMonth()
  const dateFrom = `${year}-${String(month + 1).padStart(2, '0')}-01`
  const lastDay = new Date(year, month + 1, 0).getDate()
  const dateTo = `${year}-${String(month + 1).padStart(2, '0')}-${String(lastDay).padStart(2, '00')}`
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
        <Box sx={{ p: '24px', textAlign: 'center' }}>
          <Box>{t('dashboard.loginRequired')}</Box>
        </Box>
      </Layout>
    )
  }

  return (
    <Layout>
      <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
        <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '8px', color: 'grey.800' }}>
          {t('nav.dashboard')}
        </Typography>

        {ordersLoading ? (
          <PageLoadingSkeleton />
        ) : (
          <>
            {/* Stat cards */}
            <Box
              sx={{
                display: 'flex',
                gap: '16px',
                mb: '32px',
                flexWrap: 'wrap',
              }}
            >
              <StatCard label={t('dashboard.totalOrdersThisMonth')} value={stats.totalOrders} icon={ClipboardList} />
              <StatCard label={t('dashboard.revenueThisMonth')} value={formatPrice(stats.revenue)} icon={DollarSign} />
              <StatCard label={t('dashboard.customers')} value={stats.customers} icon={Users} />
              <StatCard label={t('dashboard.products')} value={stats.products} icon={Package} />
            </Box>

            {/* Quick links */}
            <Box sx={{ display: 'flex', gap: '8px', mb: '32px', flexWrap: 'wrap' }}>
              <Link href="/orders" passHref legacyBehavior>
                <Button component="a" variant="outlined">
                  {t('dashboard.newOrder')}
                </Button>
              </Link>
              <Link href="/customers" passHref legacyBehavior>
                <Button component="a" variant="outlined">
                  {t('dashboard.addCustomer')}
                </Button>
              </Link>
              <Link href="/products" passHref legacyBehavior>
                <Button component="a" variant="outlined">
                  {t('dashboard.addProduct')}
                </Button>
              </Link>
            </Box>

            {/* Recent orders */}
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
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  p: '24px',
                  borderBottom: '1px solid',
                  borderColor: 'grey.200',
                  bgcolor: 'grey.50',
                }}
              >
                <Typography component="h2" sx={{ fontSize: '16px', fontWeight: 600 }}>
                  {t('dashboard.recentOrders')}
                </Typography>
                <Link href="/orders" passHref legacyBehavior>
                  <Button component="a" variant="outlined" sx={{ fontSize: '12px', py: '4px', px: '8px' }}>
                    {t('dashboard.viewAll')}
                  </Button>
                </Link>
              </Box>
              {recentOrders.length > 0 ? (
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
                        {t('dashboard.order')}
                      </Box>
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
                        {t('common.customer')}
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
                        {t('common.total')}
                      </Box>
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
                        {t('common.status')}
                      </Box>
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
                        {t('dashboard.date')}
                      </Box>
                    </Box>
                  </Box>
                  <Box component="tbody">
                    {recentOrders.map((order) => {
                      const statusStyle = getStatusStyle(order.status)
                      return (
                        <Box
                          component="tr"
                          key={order.id}
                          sx={{
                            borderTop: '1px solid',
                            borderColor: 'grey.200',
                            '&:hover': { bgcolor: 'grey.50' },
                          }}
                        >
                          <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px' }}>
                            <Link
                              href="/orders"
                              style={{ color: 'inherit', fontWeight: 600, textDecoration: 'none' }}
                            >
                              #{order.id}
                            </Link>
                          </Box>
                          <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px' }}>
                            {order.customer_name}
                          </Box>
                          <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontSize: '14px' }}>
                            {formatPrice(order.total_price)}
                          </Box>
                          <Box component="td" sx={{ py: '8px', px: '16px' }}>
                            <Box
                              sx={{
                                display: 'inline-block',
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
                              {t(`orderStatus.${order.status}`)}
                            </Box>
                          </Box>
                          <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px', color: 'grey.500' }}>
                            {formatDate(order.created_at)}
                          </Box>
                        </Box>
                      )
                    })}
                  </Box>
                </Box>
              ) : (
                <Box sx={{ p: '32px', textAlign: 'center', color: 'grey.500' }}>
                  <Box sx={{ fontSize: '16px', display: 'block', mb: '8px' }}>{t('dashboard.noOrdersYet')}</Box>
                  <Link href="/orders" passHref legacyBehavior>
                    <Button component="a" variant="contained" disableElevation>{t('dashboard.createFirstOrder')}</Button>
                  </Link>
                </Box>
              )}
            </Paper>
          </>
        )}
      </Box>
    </Layout>
  )
}

export default DashboardPage
