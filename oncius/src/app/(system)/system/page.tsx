"use client"

import { useMemo } from 'react'
import { useQuery } from 'react-query'
import { Box, Paper, Typography } from '@mui/material'
import { api } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { Store, CreditCard, TrendingUp, type LucideIcon } from 'lucide-react'

const subStatusColors: Record<string, { bg: string; color: string }> = {
  trialing:  { bg: '#E3F2FD', color: '#1565C0' },
  active:    { bg: '#E8F5E9', color: '#2E7D32' },
  expired:   { bg: '#FFEBEE', color: '#C62828' },
  cancelled: { bg: '#F5F5F5', color: '#616161' },
  past_due:  { bg: '#FFF3E0', color: '#E65100' },
}

const StatCard = ({ label, value, icon: Icon, sub }: { label: string; value: string | number; icon: LucideIcon; sub?: string }) => (
  <Paper sx={{ p: '24px', borderRadius: '12px', border: '1px solid', borderColor: 'grey.200', bgcolor: 'background.paper', flex: 1, minWidth: { xs: '100%', sm: 140 } }}>
    <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
      <Box>
        <Box sx={{ color: 'text.secondary', fontSize: '13px', fontWeight: 600, mb: '4px' }}>{label}</Box>
        <Box sx={{ fontSize: '22px', fontWeight: 700 }}>{value}</Box>
        {sub && <Box sx={{ fontSize: '12px', color: 'text.secondary', mt: '4px' }}>{sub}</Box>}
      </Box>
      <Box sx={{ opacity: 0.5 }}><Icon size={28} /></Box>
    </Box>
  </Paper>
)

export default function SystemDashboardPage() {
  const { data: statsRes, isLoading } = useQuery('system-stats', () => api.getSystemStats())
  const { data: shopsRes } = useQuery('system-shops', () => api.getSystemShops())

  const stats = statsRes?.data
  const shops: any[] = useMemo(() => shopsRes?.data ?? [], [shopsRes])

  const recentShops = useMemo(() => shops.slice(0, 8), [shops])

  function formatPrice(v: number) {
    return `Rp ${v.toLocaleString('id-ID')}`
  }

  function formatDate(d: string) {
    return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }

  if (isLoading) return <PageLoadingSkeleton />

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
      <Box sx={{ mb: '8px', display: 'flex', alignItems: 'center', gap: '10px' }}>
        <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700 }}>
          System Dashboard
        </Typography>
        <Box sx={{ px: '8px', py: '2px', bgcolor: 'warning.main', color: 'white', borderRadius: '4px', fontSize: '11px', fontWeight: 700 }}>
          SYSTEM
        </Box>
      </Box>

      {/* Stat cards */}
      <Box sx={{ display: 'flex', gap: '16px', mb: '32px', flexWrap: 'wrap' }}>
        <StatCard label="Total Shops" value={stats?.total_shops ?? 0} icon={Store} />
        <StatCard
          label="MRR"
          value={stats ? formatPrice(stats.mrr_idr) : '—'}
          icon={TrendingUp}
          sub={`${stats?.subs_active ?? 0} active`}
        />
        <StatCard label="Subscriptions" value={stats?.subs_trialing ?? 0} icon={CreditCard} sub="on trial" />
      </Box>

      {/* Sub status breakdown */}
      <Box sx={{ display: 'flex', gap: '12px', mb: '32px', flexWrap: 'wrap' }}>
        {[
          { label: 'Active', count: stats?.subs_active ?? 0, status: 'active' },
          { label: 'Trialing', count: stats?.subs_trialing ?? 0, status: 'trialing' },
          { label: 'Expired', count: stats?.subs_expired ?? 0, status: 'expired' },
          { label: 'Cancelled', count: stats?.subs_cancelled ?? 0, status: 'cancelled' },
        ].map(({ label, count, status }) => {
          const colors = subStatusColors[status] ?? { bg: '#F5F5F5', color: '#616161' }
          return (
            <Box key={status} sx={{ px: '16px', py: '10px', borderRadius: '8px', bgcolor: colors.bg, color: colors.color, fontSize: '14px', fontWeight: 600 }}>
              {label}: {count}
            </Box>
          )
        })}
      </Box>

      {/* Recent shops */}
      <Paper sx={{ borderRadius: '12px', border: '1px solid', borderColor: 'grey.200', bgcolor: 'background.paper', overflow: 'hidden' }}>
        <Box sx={{ p: '20px', borderBottom: '1px solid', borderColor: 'grey.200', bgcolor: 'action.hover' }}>
          <Typography component="h2" sx={{ fontSize: '15px', fontWeight: 600 }}>Recent Shops</Typography>
        </Box>
        <Box sx={{ overflowX: 'auto' }}>
          <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 600 }}>
            <Box component="thead">
              <Box component="tr" sx={{ bgcolor: 'action.hover' }}>
                {['Shop', 'Owner', 'Plan', 'Status', 'Joined'].map(col => (
                  <Box key={col} component="th" sx={{ p: '12px 16px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {col}
                  </Box>
                ))}
              </Box>
            </Box>
            <Box component="tbody">
              {recentShops.map((shop: any) => {
                const colors = subStatusColors[shop.sub_status] ?? { bg: '#F5F5F5', color: '#616161' }
                return (
                  <Box component="tr" key={shop.shop_id} sx={{ borderTop: '1px solid', borderColor: 'grey.200', '&:hover': { bgcolor: 'action.hover' } }}>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '14px', fontWeight: 600 }}>{shop.shop_name}</Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '14px' }}>
                      <Box>{shop.owner_name}</Box>
                      <Box sx={{ fontSize: '12px', color: 'text.secondary' }}>{shop.owner_email}</Box>
                    </Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '14px' }}>{shop.plan_name || '—'}</Box>
                    <Box component="td" sx={{ p: '10px 16px' }}>
                      <Box sx={{ display: 'inline-block', px: '8px', py: '2px', borderRadius: '4px', fontSize: '12px', fontWeight: 500, bgcolor: colors.bg, color: colors.color, textTransform: 'capitalize' }}>
                        {shop.sub_status || '—'}
                      </Box>
                    </Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '13px', color: 'text.secondary' }}>{formatDate(shop.joined_at)}</Box>
                  </Box>
                )
              })}
              {recentShops.length === 0 && (
                <Box component="tr">
                  <Box component="td" colSpan={5} sx={{ p: '32px', textAlign: 'center', color: 'text.secondary', fontSize: '14px' }}>No shops yet</Box>
                </Box>
              )}
            </Box>
          </Box>
        </Box>
      </Paper>
    </Box>
  )
}
