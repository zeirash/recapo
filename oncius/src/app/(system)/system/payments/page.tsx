"use client"

import { useState } from 'react'
import { useQuery } from 'react-query'
import { Box, Button, MenuItem, Paper, Select, Typography } from '@mui/material'
import DateRangeFilter from '@/components/ui/DateRangeFilter'
import { api } from '@/utils/api'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'

const FILTER_SESSION_KEY = 'system_payments_filter_state'

type SortBy = 'created_at,desc' | 'created_at,asc' | 'amount_idr,desc' | 'amount_idr,asc'

const paymentStatusColors: Record<string, { bg: string; color: string }> = {
  settlement: { bg: '#E8F5E9', color: '#2E7D32' },
  capture:    { bg: '#E8F5E9', color: '#2E7D32' },
  pending:    { bg: '#FFF3E0', color: '#E65100' },
  deny:       { bg: '#FFEBEE', color: '#C62828' },
  cancel:     { bg: '#FFEBEE', color: '#C62828' },
  expire:     { bg: '#FFEBEE', color: '#C62828' },
  failure:    { bg: '#FFEBEE', color: '#C62828' },
}

function getStoredFilterState() {
  if (typeof window === 'undefined') return null
  try {
    const s = sessionStorage.getItem(FILTER_SESSION_KEY)
    return s ? JSON.parse(s) : null
  } catch {
    return null
  }
}

function formatPrice(v: number) {
  return `Rp ${v.toLocaleString('id-ID')}`
}

function formatDate(d: string | null) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export default function SystemPaymentsPage() {
  const stored = getStoredFilterState()

  const [dateFrom, setDateFrom] = useState<string>(stored?.dateFrom ?? '')
  const [dateTo, setDateTo] = useState<string>(stored?.dateTo ?? '')
  const [statusFilter, setStatusFilter] = useState<string>(stored?.statusFilter ?? '')
  const [sortBy, setSortBy] = useState<SortBy>(stored?.sortBy ?? 'created_at,desc')

  function saveState(patch: object) {
    try {
      sessionStorage.setItem(FILTER_SESSION_KEY, JSON.stringify({
        dateFrom, dateTo, statusFilter, sortBy, ...patch,
      }))
    } catch {}
  }

  function update<T>(setter: (v: T) => void, key: string) {
    return (val: T) => { setter(val); saveState({ [key]: val }) }
  }

  const { data: paymentsRes, isLoading } = useQuery(
    ['system-payments', dateFrom, dateTo, statusFilter, sortBy],
    () => api.getSystemPayments({
      date_from: dateFrom || undefined,
      date_to: dateTo || undefined,
      status: statusFilter || undefined,
      sort: sortBy,
    })
  )

  const payments: any[] = paymentsRes?.data ?? []

  const totalRevenue = payments
    .filter(p => p.status === 'settlement' || p.status === 'capture')
    .reduce((sum, p) => sum + (p.amount_idr || 0), 0)

  function handleResetFilters() {
    setDateFrom('')
    setDateTo('')
    setStatusFilter('')
    setSortBy('created_at,desc')
    saveState({ dateFrom: '', dateTo: '', statusFilter: '', sortBy: 'created_at,desc' })
  }

  const isFiltered = dateFrom || dateTo || statusFilter || sortBy !== 'created_at,desc'

  if (isLoading) return <PageLoadingSkeleton />

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
      <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '8px' }}>
        Payments ({payments.length})
      </Typography>
      <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '24px' }}>
        Total settled: <Box component="span" sx={{ fontWeight: 700, color: 'success.main' }}>{formatPrice(totalRevenue)}</Box>
      </Box>

      {/* Filters */}
      <Box sx={{ display: 'flex', gap: '12px', mb: '16px', flexWrap: 'wrap', alignItems: 'flex-end' }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
          <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>Sort</Box>
          <Select
            size="small"
            value={sortBy}
            onChange={(e) => update<SortBy>(setSortBy, 'sortBy')(e.target.value as SortBy)}
            sx={{ height: 34, fontSize: '13px', borderRadius: '6px', minWidth: 160 }}
            MenuProps={{ anchorOrigin: { vertical: 'bottom', horizontal: 'left' }, transformOrigin: { vertical: 'top', horizontal: 'left' } }}
          >
            <MenuItem value="created_at,desc" sx={{ fontSize: '13px' }}>Newest first</MenuItem>
            <MenuItem value="created_at,asc"  sx={{ fontSize: '13px' }}>Oldest first</MenuItem>
            <MenuItem value="amount_idr,desc" sx={{ fontSize: '13px' }}>Amount: high to low</MenuItem>
            <MenuItem value="amount_idr,asc"  sx={{ fontSize: '13px' }}>Amount: low to high</MenuItem>
          </Select>
        </Box>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
          <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>Status</Box>
          <Select
            size="small"
            displayEmpty
            value={statusFilter}
            onChange={(e) => update<string>(setStatusFilter, 'statusFilter')(e.target.value)}
            sx={{ height: 34, fontSize: '13px', borderRadius: '6px', minWidth: 130 }}
            MenuProps={{ anchorOrigin: { vertical: 'bottom', horizontal: 'left' }, transformOrigin: { vertical: 'top', horizontal: 'left' } }}
          >
            <MenuItem value="" sx={{ fontSize: '13px' }}>All</MenuItem>
            <MenuItem value="settlement" sx={{ fontSize: '13px' }}>Settlement</MenuItem>
            <MenuItem value="capture" sx={{ fontSize: '13px' }}>Capture</MenuItem>
            <MenuItem value="pending" sx={{ fontSize: '13px' }}>Pending</MenuItem>
            <MenuItem value="deny" sx={{ fontSize: '13px' }}>Deny</MenuItem>
            <MenuItem value="cancel" sx={{ fontSize: '13px' }}>Cancel</MenuItem>
            <MenuItem value="expire" sx={{ fontSize: '13px' }}>Expire</MenuItem>
            <MenuItem value="failure" sx={{ fontSize: '13px' }}>Failure</MenuItem>
          </Select>
        </Box>
        <DateRangeFilter
          dateFrom={dateFrom}
          dateTo={dateTo}
          onDateFromChange={update<string>(setDateFrom, 'dateFrom')}
          onDateToChange={update<string>(setDateTo, 'dateTo')}
        />
        {isFiltered && (
          <Button
            size="small"
            variant="outlined"
            onClick={handleResetFilters}
            sx={{ height: 34, fontSize: '13px', borderRadius: '6px', textTransform: 'none', alignSelf: 'flex-end', flexShrink: 0 }}
          >
            Reset
          </Button>
        )}
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
                  <Box component="td" colSpan={7} sx={{ p: '32px', textAlign: 'center', color: 'text.secondary', fontSize: '14px' }}>No payments found</Box>
                </Box>
              )}
            </Box>
          </Box>
        </Box>
      </Paper>
    </Box>
  )
}
