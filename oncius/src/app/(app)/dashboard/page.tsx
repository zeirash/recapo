"use client"

import { useMemo, useState } from 'react'
import { useQuery } from 'react-query'
import Link from 'next/link'
import { Box, Button, Paper, Tooltip, Typography } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { api } from '@/utils/api'
import { Package, DollarSign, Users, ClipboardList, AlertTriangle, CreditCard, Info, type LucideIcon } from 'lucide-react'

type Order = {
  id: number
  customer_name: string
  total_price: number
  status: string
  payment_status: string
  created_at: string
}

const lightStatusColors: Record<string, { bg: string; color: string }> = {
  created: { bg: '#E3F2FD', color: '#1565C0' },
  in_progress: { bg: '#FFF3E0', color: '#E65100' },
  in_delivery: { bg: '#F3E5F5', color: '#7B1FA2' },
  done: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FFEBEE', color: '#C62828' },
}

const darkStatusColors: Record<string, { bg: string; color: string }> = {
  created: { bg: '#1e3a5f', color: '#90caf9' },
  in_progress: { bg: '#3e2000', color: '#ffb74d' },
  in_delivery: { bg: '#2d1b4e', color: '#ce93d8' },
  done: { bg: '#1b3a2d', color: '#81c784' },
  cancelled: { bg: '#3e1a1a', color: '#ef9a9a' },
}

const StatCard = ({
  label,
  value,
  icon: Icon,
  delta,
  sublabel,
  tooltip,
}: {
  label: string
  value: number | string
  icon: LucideIcon
  delta?: { display: string; positive: boolean }
  sublabel?: string
  tooltip?: string
}) => (
  <Paper
    sx={{
      p: '24px',
      borderRadius: '12px',
      boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
      border: '1px solid',
      borderColor: 'grey.200',
      bgcolor: 'background.paper',
      flex: 1,
      minWidth: { xs: '100%', sm: 140 },
    }}
  >
    <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
      <Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px', mb: '4px' }}>
          <Box sx={{ color: 'text.secondary', fontSize: '14px', fontWeight: 600 }}>
            {label}
          </Box>
          {tooltip && (
            <Tooltip title={tooltip} placement="top" arrow>
              <Box sx={{ color: 'text.disabled', display: 'flex', cursor: 'help' }}>
                <Info size={13} />
              </Box>
            </Tooltip>
          )}
        </Box>
        <Box sx={{ fontSize: '20px', fontWeight: 700 }}>{value}</Box>
        {sublabel && (
          <Box sx={{ fontSize: '12px', color: 'text.secondary', mt: '2px' }}>{sublabel}</Box>
        )}
        {delta && (
          <Box sx={{ fontSize: '12px', color: delta.positive ? 'success.main' : 'error.main', mt: '4px' }}>
            {delta.positive ? '↑' : '↓'} {delta.display}
          </Box>
        )}
      </Box>
      <Box sx={{ opacity: 0.6 }}><Icon size={28} /></Box>
    </Box>
  </Paper>
)

type Preset = 'this_month' | 'last_7' | 'last_30' | 'all_time'

const toDateStr = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`

const getDateRange = (preset: Preset): { dateFrom: string | undefined; dateTo: string | undefined } => {
  const now = new Date()
  if (preset === 'all_time') return { dateFrom: undefined, dateTo: undefined }
  if (preset === 'last_7') {
    const from = new Date(now); from.setDate(now.getDate() - 6)
    return { dateFrom: toDateStr(from), dateTo: toDateStr(now) }
  }
  if (preset === 'last_30') {
    const from = new Date(now); from.setDate(now.getDate() - 29)
    return { dateFrom: toDateStr(from), dateTo: toDateStr(now) }
  }
  // this_month
  const year = now.getFullYear()
  const month = now.getMonth()
  const lastDay = new Date(year, month + 1, 0).getDate()
  return {
    dateFrom: `${year}-${String(month + 1).padStart(2, '0')}-01`,
    dateTo: `${year}-${String(month + 1).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`,
  }
}

const getPrevDateRange = (preset: Preset): { dateFrom: string | undefined; dateTo: string | undefined } => {
  if (preset === 'all_time') return { dateFrom: undefined, dateTo: undefined }
  const now = new Date()
  if (preset === 'last_7') {
    const to = new Date(now); to.setDate(now.getDate() - 7)
    const from = new Date(now); from.setDate(now.getDate() - 13)
    return { dateFrom: toDateStr(from), dateTo: toDateStr(to) }
  }
  if (preset === 'last_30') {
    const to = new Date(now); to.setDate(now.getDate() - 30)
    const from = new Date(now); from.setDate(now.getDate() - 59)
    return { dateFrom: toDateStr(from), dateTo: toDateStr(to) }
  }
  // this_month → previous calendar month
  const year = now.getFullYear()
  const month = now.getMonth()
  const prevMonth = month === 0 ? 11 : month - 1
  const prevYear = month === 0 ? year - 1 : year
  const lastDay = new Date(prevYear, prevMonth + 1, 0).getDate()
  return {
    dateFrom: `${prevYear}-${String(prevMonth + 1).padStart(2, '0')}-01`,
    dateTo: `${prevYear}-${String(prevMonth + 1).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`,
  }
}

const DashboardPage = () => {
  const t = useTranslations()
  const { isAuthenticated } = useAuth()
  const theme = useTheme()
  const statusColors = theme.palette.mode === 'dark' ? darkStatusColors : lightStatusColors
  const [preset, setPreset] = useState<Preset>('this_month')
  const { dateFrom, dateTo } = getDateRange(preset)
  const { dateFrom: prevFrom, dateTo: prevTo } = getPrevDateRange(preset)

  const { data: ordersRes, isLoading: ordersLoading } = useQuery(
    ['orders', 'dashboard', preset],
    async () => {
      const res = await api.getOrders({ date_from: dateFrom, date_to: dateTo })
      if (!res.success) throw new Error(res.message)
      return res.data as Order[]
    }
  )

  const { data: prevOrdersRes } = useQuery(
    ['orders', 'dashboard', 'prev', preset],
    async () => {
      const res = await api.getOrders({ date_from: prevFrom, date_to: prevTo })
      if (!res.success) throw new Error(res.message)
      return res.data as Order[]
    },
    { enabled: preset !== 'all_time' }
  )

  const { data: orderStatsRes } = useQuery(
    ['orders', 'stats', preset],
    async () => {
      const res = await api.getOrderStats({ date_from: dateFrom, date_to: dateTo })
      if (!res.success) throw new Error(res.message)
      return res.data as { total_revenue: number; net_sales: number }
    }
  )

  const { data: prevOrderStatsRes } = useQuery(
    ['orders', 'stats', 'prev', preset],
    async () => {
      const res = await api.getOrderStats({ date_from: prevFrom, date_to: prevTo })
      if (!res.success) throw new Error(res.message)
      return res.data as { total_revenue: number; net_sales: number }
    },
    { enabled: preset !== 'all_time' }
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
    return {
      totalOrders: orders.length,
      revenue: orders.reduce((sum, o) => sum + (o.total_price || 0), 0),
      saldo: orderStatsRes?.total_revenue ?? 0,
      netSales: orderStatsRes?.net_sales ?? 0,
      customers: (customersRes || []).length,
      products: (productsRes || []).length,
    }
  }, [ordersRes, orderStatsRes, customersRes, productsRes])

  const netSalesSublabel = t('dashboard.netSalesSublabel', { amount: formatPrice(stats.netSales) })

  const recentOrders = useMemo(() => {
    const orders = ordersRes || []
    return [...orders]
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 5)
  }, [ordersRes])

  const statusBreakdown = useMemo(() => {
    const orders = ordersRes || []
    const statuses = ['created', 'in_progress', 'in_delivery', 'done', 'cancelled']
    return statuses.map((status) => ({
      status,
      count: orders.filter((o) => o.status === status).length,
    }))
  }, [ordersRes])

  const deltas = useMemo(() => {
    if (preset === 'all_time' || !prevOrdersRes || !prevOrderStatsRes) return null
    const prevRevenue = prevOrdersRes.reduce((sum, o) => sum + (o.total_price || 0), 0)
    const ordersDiff = stats.totalOrders - prevOrdersRes.length
    const revenueDiff = stats.revenue - prevRevenue
    const saldoDiff = stats.saldo - prevOrderStatsRes.total_revenue
    return {
      orders: { display: `${ordersDiff > 0 ? '+' : ''}${ordersDiff}`, positive: ordersDiff >= 0 },
      revenue: { display: `${revenueDiff > 0 ? '+' : ''}${revenueDiff.toLocaleString()}`, positive: revenueDiff >= 0 },
      saldo: { display: `${saldoDiff > 0 ? '+' : ''}${saldoDiff.toLocaleString()}`, positive: saldoDiff >= 0 },
    }
  }, [preset, prevOrdersRes, prevOrderStatsRes, stats])

  const topCustomers = useMemo(() => {
    const orders = ordersRes || []
    const map: Record<string, { name: string; orders: number; revenue: number }> = {}
    for (const o of orders) {
      if (!o.customer_name) continue
      if (!map[o.customer_name]) map[o.customer_name] = { name: o.customer_name, orders: 0, revenue: 0 }
      map[o.customer_name].orders += 1
      map[o.customer_name].revenue += o.total_price || 0
    }
    return Object.values(map)
      .sort((a, b) => b.revenue - a.revenue)
      .slice(0, 3)
  }, [ordersRes])

  const staleOrders = useMemo(() => {
    const orders = ordersRes || []
    const threeDaysAgo = Date.now() - 3 * 24 * 60 * 60 * 1000
    return orders.filter(
      (o) => o.status === 'created' && new Date(o.created_at).getTime() < threeDaysAgo
    )
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
    return statusColors[status] || (theme.palette.mode === 'dark' ? { bg: '#2a2a2a', color: '#bdbdbd' } : { bg: '#F5F5F5', color: '#616161' })
  }

  if (!isAuthenticated) {
    return (
      <Box sx={{ p: '24px', textAlign: 'center' }}>
        <Box>{t('dashboard.loginRequired')}</Box>
      </Box>
    )
  }

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: '12px', mb: '24px' }}>
          <Box>
            <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '2px' }}>
              {t('nav.dashboard')}
            </Typography>
            <Typography sx={{ fontSize: '13px', color: 'text.secondary' }}>
              {t(`dashboard.preset_${preset}`)}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
            {(['this_month', 'last_7', 'last_30', 'all_time'] as Preset[]).map((p) => (
              <Button
                key={p}
                size="small"
                variant={preset === p ? 'contained' : 'outlined'}
                disableElevation
                onClick={() => setPreset(p)}
                sx={{ fontSize: '12px', py: '4px', px: '10px', minWidth: 0, whiteSpace: 'nowrap' }}
              >
                {t(`dashboard.preset_${p}`)}
              </Button>
            ))}
          </Box>
        </Box>

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
              <StatCard label={t('dashboard.totalOrders')} value={stats.totalOrders} icon={ClipboardList} delta={deltas?.orders} />
              <StatCard label={t('dashboard.revenue')} value={formatPrice(stats.revenue)} icon={DollarSign} delta={deltas?.revenue} tooltip={t('dashboard.revenueTooltip')} sublabel={netSalesSublabel} />
              <StatCard label={t('dashboard.customers')} value={stats.customers} icon={Users} />
              <StatCard label={t('dashboard.products')} value={stats.products} icon={Package} />
              <StatCard label={t('dashboard.saldo')} value={formatPrice(stats.saldo)} icon={CreditCard} delta={deltas?.saldo} tooltip={t('dashboard.saldoTooltip')} />
            </Box>

            {/* Status breakdown */}
            <Box sx={{ mb: '24px' }}>
              <Box sx={{ display: 'flex', alignItems: 'baseline', gap: '8px', mb: '10px' }}>
                <Typography sx={{ fontSize: '13px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  {t('dashboard.ordersByStatus')}
                </Typography>
                <Typography sx={{ fontSize: '12px', color: 'text.disabled' }}>
                  {t(`dashboard.preset_${preset}`)}
                </Typography>
              </Box>
              <Box sx={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                {statusBreakdown.map(({ status, count }) => {
                  const style = getStatusStyle(status)
                  return (
                    <Box
                      key={status}
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '6px',
                        px: '12px',
                        py: '6px',
                        borderRadius: '20px',
                        bgcolor: style.bg,
                        color: style.color,
                        fontSize: '13px',
                        fontWeight: 500,
                      }}
                    >
                      <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: style.color, flexShrink: 0 }} />
                      {t(`orderStatus.${status}`)}
                      <Box component="span" sx={{ fontWeight: 700 }}>{count}</Box>
                    </Box>
                  )
                })}
              </Box>
            </Box>

            {/* Quick links */}
            <Box sx={{ display: 'flex', gap: '8px', mb: '32px', justifyContent: 'center' }}>
              <Link href="/orders" passHref legacyBehavior>
                <Button component="a" variant="outlined" sx={{ flex: 1, fontSize: { xs: '11px', sm: '14px' }, whiteSpace: 'nowrap' }}>{t('dashboard.newOrder')}</Button>
              </Link>
              <Link href="/customers" passHref legacyBehavior>
                <Button component="a" variant="outlined" sx={{ flex: 1, fontSize: { xs: '11px', sm: '14px' }, whiteSpace: 'nowrap' }}>{t('dashboard.addCustomer')}</Button>
              </Link>
              <Link href="/products" passHref legacyBehavior>
                <Button component="a" variant="outlined" sx={{ flex: 1, fontSize: { xs: '11px', sm: '14px' }, whiteSpace: 'nowrap' }}>{t('dashboard.addProduct')}</Button>
              </Link>
            </Box>

            {/* Top customers */}
            {topCustomers.length > 0 && (
              <Box sx={{ mb: '32px' }}>
                <Typography sx={{ fontSize: '13px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em', mb: '10px' }}>
                  {t('dashboard.topCustomers')}
                </Typography>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {topCustomers.map((c, i) => (
                    <Box
                      key={c.name}
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        px: '16px',
                        py: '12px',
                        borderRadius: '8px',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        bgcolor: 'background.paper',
                      }}
                    >
                      <Box sx={{
                        width: 28, height: 28, borderRadius: '50%', bgcolor: 'action.selected',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: '13px', fontWeight: 700, color: 'text.secondary', flexShrink: 0,
                      }}>
                        {i + 1}
                      </Box>
                      <Box sx={{ flex: 1, minWidth: 0 }}>
                        <Box sx={{ fontWeight: 600, fontSize: '14px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {c.name}
                        </Box>
                        <Box sx={{ fontSize: '12px', color: 'text.secondary' }}>
                          {t('dashboard.topCustomerOrders', { count: c.orders })}
                        </Box>
                      </Box>
                      <Box sx={{ fontWeight: 700, fontSize: '14px', flexShrink: 0 }}>
                        {formatPrice(c.revenue)}
                      </Box>
                    </Box>
                  ))}
                </Box>
              </Box>
            )}

            {/* Stale orders alert */}
            {staleOrders.length > 0 && (
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '10px',
                  mb: '16px',
                  px: '16px',
                  py: '12px',
                  borderRadius: '8px',
                  bgcolor: theme.palette.mode === 'dark' ? '#3e2000' : '#FFF3E0',
                  color: theme.palette.mode === 'dark' ? '#ffb74d' : '#E65100',
                  border: '1px solid',
                  borderColor: theme.palette.mode === 'dark' ? '#5a3000' : '#FFE0B2',
                }}
              >
                <AlertTriangle size={18} style={{ flexShrink: 0 }} />
                <Typography sx={{ fontSize: '14px', fontWeight: 500, color: 'inherit' }}>
                  {t('dashboard.staleOrdersAlert', { count: staleOrders.length })}
                </Typography>
              </Box>
            )}

            {/* Recent orders */}
            <Paper
              sx={{
                borderRadius: '12px',
                boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
                border: '1px solid',
                borderColor: 'grey.200',
                bgcolor: 'background.paper',
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
                  bgcolor: 'action.hover',
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
                <Box sx={{ overflowX: 'auto' }}>
                <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 480 }}>
                  <Box component="thead">
                    <Box component="tr" sx={{ bgcolor: 'action.hover' }}>
                      <Box
                        component="th"
                        sx={{
                          p: '16px',
                          textAlign: 'left',
                          fontSize: '12px',
                          fontWeight: 600,
                          color: 'text.secondary',
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
                          color: 'text.secondary',
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
                          color: 'text.secondary',
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
                          color: 'text.secondary',
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
                          color: 'text.secondary',
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
                            '&:hover': { bgcolor: 'action.hover' },
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
                          <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px', color: 'text.secondary' }}>
                            {formatDate(order.created_at)}
                          </Box>
                        </Box>
                      )
                    })}
                  </Box>
                </Box>
                </Box>
              ) : (
                <Box sx={{ p: '32px', textAlign: 'center', color: 'text.secondary' }}>
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
  )
}

export default DashboardPage
