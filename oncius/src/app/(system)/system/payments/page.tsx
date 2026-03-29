"use client"

import { useQuery } from 'react-query'
import { Box, Paper, Typography } from '@mui/material'
import { api } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'

const paymentStatusColors: Record<string, { bg: string; color: string }> = {
  settlement: { bg: '#E8F5E9', color: '#2E7D32' },
  capture:    { bg: '#E8F5E9', color: '#2E7D32' },
  pending:    { bg: '#FFF3E0', color: '#E65100' },
  deny:       { bg: '#FFEBEE', color: '#C62828' },
  cancel:     { bg: '#FFEBEE', color: '#C62828' },
  expire:     { bg: '#FFEBEE', color: '#C62828' },
  failure:    { bg: '#FFEBEE', color: '#C62828' },
}

export default function SystemPaymentsPage() {
  const { data: paymentsRes, isLoading } = useQuery('system-payments', () => api.getSystemPayments())

  const payments: any[] = paymentsRes?.data ?? []

  const totalRevenue = payments
    .filter(p => p.status === 'settlement' || p.status === 'capture')
    .reduce((sum, p) => sum + (p.amount_idr || 0), 0)

  function formatPrice(v: number) {
    return `Rp ${v.toLocaleString('id-ID')}`
  }

  function formatDate(d: string | null) {
    if (!d) return '—'
    return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' })
  }

  if (isLoading) return <PageLoadingSkeleton />

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
      <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '8px' }}>
        Payments ({payments.length})
      </Typography>
      <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '24px' }}>
        Total settled: <Box component="span" sx={{ fontWeight: 700, color: 'success.main' }}>{formatPrice(totalRevenue)}</Box>
      </Box>

      <Paper sx={{ borderRadius: '12px', border: '1px solid', borderColor: 'grey.200', bgcolor: 'background.paper', overflow: 'hidden' }}>
        <Box sx={{ overflowX: 'auto' }}>
          <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse', minWidth: 700 }}>
            <Box component="thead">
              <Box component="tr" sx={{ bgcolor: 'action.hover' }}>
                {['Shop', 'Plan', 'Amount', 'Status', 'Midtrans ID', 'Paid At', 'Created'].map(col => (
                  <Box key={col} component="th" sx={{ p: '12px 16px', textAlign: 'left', fontSize: '12px', fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {col}
                  </Box>
                ))}
              </Box>
            </Box>
            <Box component="tbody">
              {payments.map((pay: any, i: number) => {
                const colors = paymentStatusColors[pay.status] ?? { bg: '#F5F5F5', color: '#616161' }
                return (
                  <Box component="tr" key={i} sx={{ borderTop: '1px solid', borderColor: 'grey.200', '&:hover': { bgcolor: 'action.hover' } }}>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '14px', fontWeight: 600 }}>{pay.shop_name}</Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '14px' }}>{pay.plan_name}</Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '14px', fontWeight: 600 }}>{formatPrice(pay.amount_idr)}</Box>
                    <Box component="td" sx={{ p: '10px 16px' }}>
                      <Box sx={{ display: 'inline-block', px: '8px', py: '2px', borderRadius: '4px', fontSize: '12px', fontWeight: 500, bgcolor: colors.bg, color: colors.color, textTransform: 'capitalize' }}>
                        {pay.status}
                      </Box>
                    </Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '12px', color: 'text.secondary', fontFamily: 'monospace' }}>{pay.midtrans_order_id}</Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '13px', color: 'text.secondary' }}>{formatDate(pay.paid_at)}</Box>
                    <Box component="td" sx={{ p: '10px 16px', fontSize: '13px', color: 'text.secondary' }}>{formatDate(pay.created_at)}</Box>
                  </Box>
                )
              })}
              {payments.length === 0 && (
                <Box component="tr">
                  <Box component="td" colSpan={7} sx={{ p: '32px', textAlign: 'center', color: 'text.secondary', fontSize: '14px' }}>No payments yet</Box>
                </Box>
              )}
            </Box>
          </Box>
        </Box>
      </Paper>
    </Box>
  )
}
