"use client"

import { useState } from 'react'
import { useQuery } from 'react-query'
import { Box, Paper, Typography } from '@mui/material'
import { api } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'

const subStatusColors: Record<string, { bg: string; color: string }> = {
  trialing:  { bg: '#E3F2FD', color: '#1565C0' },
  active:    { bg: '#E8F5E9', color: '#2E7D32' },
  expired:   { bg: '#FFEBEE', color: '#C62828' },
  cancelled: { bg: '#F5F5F5', color: '#616161' },
  past_due:  { bg: '#FFF3E0', color: '#E65100' },
}

export default function SystemShopsPage() {
  const { data: shopsRes, isLoading } = useQuery('system-shops', () => api.getSystemShops())
  const [search, setSearch] = useState('')

  const shops: any[] = shopsRes?.data ?? []

  const filtered = shops.filter(s =>
    search === '' ||
    s.shop_name.toLowerCase().includes(search.toLowerCase()) ||
    s.owner_email.toLowerCase().includes(search.toLowerCase()) ||
    s.owner_name.toLowerCase().includes(search.toLowerCase())
  )

  function formatDate(d: string) {
    if (!d) return '—'
    return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }

  if (isLoading) return <PageLoadingSkeleton />

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
      <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '24px' }}>
        Shops ({shops.length})
      </Typography>

      <Box
        component="input"
        placeholder="Search by shop name, owner name or email..."
        value={search}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value)}
        sx={{
          mb: '24px', width: '100%', maxWidth: 400, px: '12px', py: '8px',
          border: '1px solid', borderColor: 'grey.300', borderRadius: '8px',
          fontSize: '14px', outline: 'none', bgcolor: 'background.paper', color: 'text.primary',
          '&:focus': { borderColor: 'primary.main' },
        }}
      />

      <Paper sx={{ borderRadius: '12px', border: '1px solid', borderColor: 'grey.200', bgcolor: 'background.paper', overflow: 'hidden' }}>
        <Box sx={{ overflowX: 'auto' }}>
          <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 700 }}>
            <Box component="thead">
              <Box component="tr" sx={{ bgcolor: 'action.hover' }}>
                {['#', 'Shop', 'Owner', 'Plan', 'Status', 'Period End', 'Joined'].map(col => (
                  <Box key={col} component="th" sx={{ p: '12px 16px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {col}
                  </Box>
                ))}
              </Box>
            </Box>
            <Box component="tbody">
              {filtered.map((shop: any) => {
                const colors = subStatusColors[shop.sub_status] ?? { bg: '#F5F5F5', color: '#616161' }
                return (
                  <Box component="tr" key={shop.shop_id} sx={{ borderTop: '1px solid', borderColor: 'grey.200', '&:hover': { bgcolor: 'action.hover' } }}>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '13px', color: 'text.secondary' }}>{shop.shop_id}</Box>
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
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '13px', color: 'text.secondary' }}>
                      {shop.trial_ends_at ? `Trial: ${formatDate(shop.trial_ends_at)}` : formatDate(shop.period_end)}
                    </Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '13px', color: 'text.secondary' }}>{formatDate(shop.joined_at)}</Box>
                  </Box>
                )
              })}
              {filtered.length === 0 && (
                <Box component="tr">
                  <Box component="td" colSpan={7} sx={{ p: '32px', textAlign: 'center', color: 'text.secondary', fontSize: '14px' }}>
                    {search ? 'No results found' : 'No shops yet'}
                  </Box>
                </Box>
              )}
            </Box>
          </Box>
        </Box>
      </Paper>
    </Box>
  )
}
