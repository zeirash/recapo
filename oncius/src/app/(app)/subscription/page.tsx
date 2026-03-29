"use client"

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from 'react-query'
import { Box, Typography, Button, Paper, Chip, CircularProgress, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material'
import { useTranslations, useLocale } from 'next-intl'
import { Check, Users, Crown, Info } from 'lucide-react'
import PageLoadingSkeleton from '@/components/ui/PageLoadingSkeleton'
import { api } from '@/utils/api'
import { useAuth } from '@/hooks/useAuth'
import type { Plan, Subscription } from '@/types'

const formatPriceIDR = (price: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(price)

const formatDate = (dateStr: string) =>
  new Date(dateStr).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })

export default function SubscriptionPage() {
  const t = useTranslations()
  const locale = useLocale()
  const router = useRouter()
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [checkoutError, setCheckoutError] = useState<string | null>(null)
  const [showCancelDialog, setShowCancelDialog] = useState(false)

  const { data: plansRes, isLoading: plansLoading } = useQuery('plans', () => api.getPlans(), {
    staleTime: 5 * 60 * 1000,
    enabled: user?.role !== 'system',
  })
  const plans: Plan[] = plansRes?.data ?? []

  const { data: subRes, isLoading: subLoading } = useQuery('subscription', () => api.getSubscription(), {
    enabled: user?.role !== 'system',
  })
  const subscription: Subscription | null = subRes?.data ?? null

  const cancelMutation = useMutation(() => api.cancelSubscription(), {
    onSuccess: () => {
      queryClient.invalidateQueries('subscription')
      setShowCancelDialog(false)
    },
  })

  const checkoutMutation = useMutation(
    (planId: number) => api.checkout(planId),
    {
      onSuccess: (res) => {
        if (res.data?.redirect_url) {
          window.location.href = res.data.redirect_url
        }
      },
      onError: (err: any) => {
        setCheckoutError(err.message || 'Checkout failed')
      },
    }
  )

  if (user?.role === 'system') {
    router.replace('/dashboard')
    return null
  }

  const isLoading = plansLoading || subLoading

  return (
    <>
      <Box sx={{ maxWidth: 900, mx: 'auto', px: { xs: '16px', sm: '24px' }, py: '32px' }}>
        <Typography component="h1" sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, mb: '8px' }}>
          {t('subscription.title')}
        </Typography>
        <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '24px' }}>
          {t('subscription.subtitle')}
        </Box>

        {/* Manual renewal notice */}
        <Box sx={{ display: 'flex', gap: '12px', p: '16px', bgcolor: 'info.50', border: '1px solid', borderColor: 'info.200', borderRadius: '10px', mb: '32px' }}>
          <Info size={18} color="var(--mui-palette-info-main)" style={{ flexShrink: 0, marginTop: '1px' }} />
          <Box>
            <Typography sx={{ fontSize: '14px', fontWeight: 600, color: 'info.dark', mb: '4px' }}>
              {t('subscription.manualRenewalTitle')}
            </Typography>
            <Typography sx={{ fontSize: '13px', color: 'info.dark', lineHeight: 1.5 }}>
              {t('subscription.manualRenewalBody')}
            </Typography>
          </Box>
        </Box>

        {/* Current subscription */}
        {!subLoading && subscription && (
          <Paper sx={{ p: '24px', borderRadius: '12px', border: '1px solid', borderColor: 'grey.200', mb: '32px' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px', mb: '12px' }}>
              <Crown size={18} />
              <Typography sx={{ fontWeight: 600, fontSize: '16px' }}>
                {t('subscription.currentPlan')}
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
              <Typography sx={{ fontSize: '18px', fontWeight: 700 }}>
                {subscription.plan.display_name}
              </Typography>
              <Chip
                label={subscription.status}
                size="small"
                sx={{
                  bgcolor: subscription.status === 'active' ? 'success.light' : 'warning.light',
                  color: subscription.status === 'active' ? 'success.dark' : 'warning.dark',
                  fontWeight: 600,
                  textTransform: 'capitalize',
                }}
              />
            </Box>
            {subscription.status === 'trialing' && subscription.trial_ends_at && (
              <Box sx={{ fontSize: '13px', color: 'text.secondary', mt: '8px' }}>
                {t('subscription.trialEnds', { date: formatDate(subscription.trial_ends_at) })}
              </Box>
            )}
            {subscription.status === 'active' && (
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mt: '8px', flexWrap: 'wrap', gap: '8px' }}>
                <Box sx={{ fontSize: '13px', color: 'text.secondary' }}>
                  {t('subscription.renewsOn', { date: formatDate(subscription.current_period_end) })}
                </Box>
                <Button
                  variant="outlined"
                  color="error"
                  size="small"
                  onClick={() => setShowCancelDialog(true)}
                >
                  {t('subscription.cancelButton')}
                </Button>
              </Box>
            )}
          </Paper>
        )}

        {/* Plans */}
        <Typography sx={{ fontWeight: 600, fontSize: '16px', mb: '16px' }}>
          {t('subscription.availablePlans')}
        </Typography>

        {checkoutError && (
          <Box sx={{ mb: '16px', p: '12px', bgcolor: 'error.light', borderRadius: '8px', fontSize: '14px', color: 'error.dark' }}>
            {checkoutError}
          </Box>
        )}

        {isLoading ? (
          <PageLoadingSkeleton />
        ) : (
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: '24px' }}>
            {plans.map((plan) => {
              const isOnThisPlan = subscription?.plan?.id === plan.id
              const isCurrent = isOnThisPlan && subscription?.status === 'active'
              return (
                <Paper
                  key={plan.id}
                  sx={{
                    p: '32px',
                    borderRadius: '12px',
                    border: '2px solid',
                    borderColor: isCurrent ? 'primary.main' : 'grey.200',
                    bgcolor: 'background.paper',
                    width: { xs: '100%', sm: 300 },
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  {isOnThisPlan && (
                    <Chip label={t('subscription.currentBadge')} size="small" color="primary" sx={{ alignSelf: 'flex-start', mb: '12px' }} />
                  )}
                  <Typography sx={{ fontSize: '18px', fontWeight: 700, mb: '4px' }}>
                    {plan.display_name}
                  </Typography>
                  <Box sx={{ display: 'flex', alignItems: 'baseline', mb: '8px' }}>
                    <Box sx={{ fontSize: '24px', fontWeight: 700 }}>
                      {formatPriceIDR(plan.price_idr)}
                    </Box>
                    <Box sx={{ fontSize: '13px', color: 'text.secondary', ml: '4px' }}>
                      {t('landing.pricingPeriod')}
                    </Box>
                  </Box>
                  <Box sx={{ fontSize: '14px', color: 'text.secondary', mb: '20px' }}>
                    {locale === 'id' ? plan.description_id : plan.description_en}
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px', mb: '20px', fontSize: '14px', color: 'text.secondary' }}>
                    <Users size={15} />
                    {t('landing.pricingMaxUsers', { count: plan.max_users })}
                  </Box>
                  <Box sx={{ mt: 'auto' }}>
                    <Button
                      variant={isCurrent ? 'outlined' : 'contained'}
                      disableElevation
                      fullWidth
                      disabled={isCurrent || checkoutMutation.isLoading}
                      onClick={() => {
                        setCheckoutError(null)
                        checkoutMutation.mutate(plan.id)
                      }}
                      sx={{ py: '12px' }}
                    >
                      {checkoutMutation.isLoading && checkoutMutation.variables === plan.id ? (
                        <CircularProgress size={18} color="inherit" />
                      ) : isCurrent ? (
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <Check size={16} />
                          {t('subscription.currentBadge')}
                        </Box>
                      ) : (
                        t('subscription.choosePlan')
                      )}
                    </Button>
                  </Box>
                </Paper>
              )
            })}
          </Box>
        )}
      </Box>

      <Dialog open={showCancelDialog} onClose={() => setShowCancelDialog(false)} maxWidth="xs" fullWidth>
        <DialogTitle>{t('subscription.cancelConfirmTitle')}</DialogTitle>
        <DialogContent>
          <Box sx={{ fontSize: '14px', color: 'text.secondary' }}>
            {t('subscription.cancelConfirmBody')}
          </Box>
          {cancelMutation.isError && (
            <Box sx={{ mt: '12px', fontSize: '13px', color: 'error.main' }}>
              {(cancelMutation.error as Error)?.message || t('subscription.cancelError')}
            </Box>
          )}
        </DialogContent>
        <DialogActions sx={{ px: '24px', pb: '16px' }}>
          <Button variant="outlined" onClick={() => setShowCancelDialog(false)} disabled={cancelMutation.isLoading}>
            {t('subscription.keepPlan')}
          </Button>
          <Button
            variant="contained"
            color="error"
            disableElevation
            onClick={() => cancelMutation.mutate()}
            disabled={cancelMutation.isLoading}
          >
            {cancelMutation.isLoading ? <CircularProgress size={18} color="inherit" /> : t('subscription.confirmCancel')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}
