"use client"

import React, { useState } from 'react'
import Link from 'next/link'
import Image from 'next/image'
import { Box, Typography, Button, Paper, Skeleton } from '@mui/material'
import { useTranslations, useLocale } from 'next-intl'
import { useQuery } from 'react-query'
import { useAuth } from '@/hooks/useAuth'
import Header from '@/components/Layout/Header'
import { Package, Tag, Users, BarChart2, Check, LayoutDashboard, ClipboardList, type LucideIcon } from 'lucide-react'
import { api } from '@/utils/api'
import type { Plan } from '@/types'

const BLOB_PATH =
  'M76.5 42.2c-12.3 8.5-28.2 12.8-38.5 24.2-10.3 11.4-15 29.8-12.8 46.2 2.2 16.4 12.6 30.8 25.5 42.2 12.9 11.4 28.3 19.8 42.2 25.2 13.9 5.4 26.3 7.8 38.5 4.2 12.2-3.6 24.2-12.6 35.5-22 11.3-9.4 21.9-19.4 30.5-30.4 8.6-11 15.2-23 18.5-35.5 3.3-12.5 3.3-25.5-1.5-37.5-4.8-12-14.4-23-26.4-31-12-8-26.4-13-40.4-14.5-14-1.5-27.6 1-39.6 6.5z'

const BlobShape = ({
  variant,
  sx = {},
}: {
  variant: 'primary' | 'accent' | 'success' | 'light'
  sx?: Record<string, unknown>
}) => {
  const colors = {
    primary: 'rgba(37, 99, 235, 0.1)',
    accent: 'rgba(245, 158, 11, 0.08)',
    success: 'rgba(16, 185, 129, 0.08)',
    light: 'rgba(255, 255, 255, 0.15)',
  }
  return (
    <Box
      component="span"
      sx={{
        position: 'absolute',
        pointerEvents: 'none',
        zIndex: 0,
        ...sx,
      }}
    >
      <svg viewBox="0 0 400 400" xmlns="http://www.w3.org/2000/svg" style={{ width: '100%', height: '100%', fill: colors[variant] }}>
        <path d={BLOB_PATH} />
      </svg>
    </Box>
  )
}

const CurveLines = () => (
  <Box
    sx={{
      position: 'absolute',
      inset: 0,
      width: '100%',
      height: '100%',
      pointerEvents: 'none',
      zIndex: 0,
      opacity: 0.4,
    }}
  >
    <svg xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" preserveAspectRatio="none" viewBox="0 0 1200 800">
      <path
        d="M0 200 Q300 100 600 200 T1200 200"
        fill="none"
        stroke="rgba(37, 99, 235, 0.15)"
        strokeWidth="1"
      />
      <path
        d="M0 400 Q400 300 800 400 T1200 400"
        fill="none"
        stroke="rgba(37, 99, 235, 0.12)"
        strokeWidth="1"
      />
      <path
        d="M0 350 Q200 250 600 350 T1200 350"
        fill="none"
        stroke="rgba(245, 158, 11, 0.25)"
        strokeWidth="1.5"
      />
      {/* Curved diagonal ~35° - cubic bezier from bottom-left toward top-right */}
      <path
        d="M0 500 C200 400 500 100 800 25"
        fill="none"
        stroke="rgba(16, 185, 129, 0.25)"
        strokeWidth="1.5"
      />
    </svg>
  </Box>
)

const FEATURES: { titleKey: string; descKey: string; icon: LucideIcon }[] = [
  { titleKey: 'landing.featureOrderManagement', descKey: 'landing.featureOrderManagementDesc', icon: Package },
  { titleKey: 'landing.featureProductCatalog', descKey: 'landing.featureProductCatalogDesc', icon: Tag },
  { titleKey: 'landing.featureCustomerDatabase', descKey: 'landing.featureCustomerDatabaseDesc', icon: Users },
  { titleKey: 'landing.featureDashboardReports', descKey: 'landing.featureDashboardReportsDesc', icon: BarChart2 },
]

type ScreenshotItem = {
  tabKey: string
  descKey: string
  src: string
  icon: LucideIcon
}

const SCREENSHOTS: ScreenshotItem[] = [
  { tabKey: 'landing.screenshotDashboard', descKey: 'landing.screenshotDashboardDesc', src: '/screenshots/dashboard.png', icon: LayoutDashboard },
  { tabKey: 'landing.screenshotOrders',    descKey: 'landing.screenshotOrdersDesc',    src: '/screenshots/orders.png',    icon: ClipboardList },
  { tabKey: 'landing.screenshotProducts',  descKey: 'landing.screenshotProductsDesc',  src: '/screenshots/products.png',  icon: Tag },
  { tabKey: 'landing.screenshotCustomers', descKey: 'landing.screenshotCustomersDesc', src: '/screenshots/customers.png', icon: Users },
]

const ScreenshotsSection = () => {
  const [active, setActive] = useState(0)
  const t = useTranslations()
  const current = SCREENSHOTS[active]

  return (
    <Box sx={{ py: '64px' }}>
      <Box sx={{ maxWidth: 1200, mx: 'auto', px: { xs: '16px', sm: '24px' } }}>
        <Typography
          component="h3"
          sx={{ fontSize: { xs: '20px', sm: '24px' }, fontWeight: 700, textAlign: 'center', mb: '8px', color: 'grey.800' }}
        >
          {t('landing.screenshotsTitle')}
        </Typography>
        <Box sx={{ fontSize: { xs: '14px', sm: '16px' }, color: 'grey.500', textAlign: 'center', mb: '40px' }}>
          {t('landing.screenshotsSubtitle')}
        </Box>

        {/* Tab buttons */}
        <Box sx={{ display: 'flex', justifyContent: 'center', gap: '8px', mb: '32px', flexWrap: 'wrap' }}>
          {SCREENSHOTS.map((s, i) => {
            const Icon = s.icon
            const isActive = i === active
            return (
              <Box
                key={s.tabKey}
                component="button"
                onClick={() => setActive(i)}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  px: '20px',
                  py: '10px',
                  zIndex: 1,
                  borderRadius: '8px',
                  border: '2px solid',
                  borderColor: isActive ? 'primary.main' : 'grey.200',
                  bgcolor: isActive ? 'primary.main' : 'white',
                  color: isActive ? 'white' : 'grey.600',
                  cursor: 'pointer',
                  fontWeight: 600,
                  fontSize: '14px',
                  transition: 'all 0.15s',
                  '&:hover': { borderColor: 'primary.main', color: isActive ? 'white' : 'primary.main' },
                }}
              >
                <Icon size={16} />
                {t(s.tabKey)}
              </Box>
            )
          })}
        </Box>

        {/* Screenshot + description */}
        <Box
          sx={{
            display: 'flex',
            flexDirection: { xs: 'column', md: 'row' },
            gap: '40px',
            alignItems: { xs: 'stretch', md: 'center' },
          }}
        >
          {/* Browser-frame screenshot */}
          <Box
            sx={{
              flex: 1,
              borderRadius: '12px',
              overflow: 'hidden',
              boxShadow: '0 20px 60px rgba(0,0,0,0.12)',
              border: '1px solid',
              borderColor: 'grey.200',
              bgcolor: 'grey.100',
              zIndex: 1,
            }}
          >
            {/* Fake browser chrome */}
            <Box
              sx={{
                bgcolor: 'grey.100',
                borderBottom: '1px solid',
                borderColor: 'grey.200',
                px: '16px',
                py: '10px',
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
              }}
            >
              {['#FF5F57', '#FFBD2E', '#28C840'].map((c) => (
                <Box key={c} sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: c }} />
              ))}
              <Box
                sx={{
                  ml: '12px',
                  flex: 1,
                  bgcolor: 'white',
                  borderRadius: '4px',
                  height: '24px',
                  maxWidth: 300,
                  border: '1px solid',
                  borderColor: 'grey.200',
                }}
              />
            </Box>
            <Image
              key={current.src}
              src={current.src}
              alt={t(current.tabKey)}
              width={1440}
              height={900}
              style={{ width: '100%', height: 'auto', display: 'block' }}
            />
          </Box>

          {/* Description */}
          <Box sx={{ width: { xs: '100%', md: 320 }, flexShrink: 0, zIndex: 1 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', mb: '16px' }}>
              <Box
                sx={{
                  p: '10px',
                  borderRadius: '10px',
                  bgcolor: 'primary.main',
                  color: 'white',
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <current.icon size={24} />
              </Box>
              <Typography component="h4" sx={{ fontSize: '20px', fontWeight: 700, color: 'grey.800' }}>
                {t(current.tabKey)}
              </Typography>
            </Box>
            <Box sx={{ fontSize: '16px', color: 'grey.600', lineHeight: 1.7 }}>
              {t(current.descKey)}
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  )
}

const formatPriceIDR = (price: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(price)

export default function HomePage() {
  const t = useTranslations()
  const locale = useLocale()
  const { isAuthenticated } = useAuth()

  const { data: plansRes, isLoading: plansLoading } = useQuery('plans', () => api.getPlans(), {
    staleTime: 5 * 60 * 1000,
  })
  const plans: Plan[] = plansRes?.data ?? []

  return (
    <Box
      sx={{
        minHeight: '100vh',
        position: 'relative',
        overflow: 'hidden',
        background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)',
      }}
    >
      {/* Curve lines */}
      <CurveLines />

      {/* Decorative blobs */}
      <BlobShape
        variant="primary"
        sx={{
          width: 400,
          height: 400,
          top: -120,
          right: -100,
        }}
      />
      <BlobShape
        variant="accent"
        sx={{
          width: 320,
          height: 320,
          bottom: '30%',
          left: -80,
        }}
      />
      <BlobShape
        variant="success"
        sx={{
          width: 280,
          height: 280,
          top: '55%',
          right: -60,
        }}
      />

      <Header />

      {/* Hero */}
      <Box
        sx={{
          position: 'relative',
          maxWidth: 1200,
          mx: 'auto',
          px: { xs: '16px', sm: '24px' },
          py: { xs: '48px', sm: '96px' },
          textAlign: 'center',
        }}
      >
        <Typography
          component="h2"
          sx={{
            fontSize: { xs: '24px', sm: '30px', md: '36px' },
            fontWeight: 700,
            lineHeight: 1.2,
            mb: '16px',
            color: 'grey.800',
          }}
        >
          {t('landing.heroTitle')}{' '}
          <Box component="span" sx={{ color: 'primary.main' }}>
            {t('landing.heroHighlight')}
          </Box>
        </Typography>
        <Box
          sx={{
            display: 'block',
            fontSize: { xs: '16px', sm: '18px' },
            color: 'grey.500',
            maxWidth: 840,
            mx: 'auto',
            mb: '16px',
            lineHeight: 1.6,
          }}
        >
          {t('landing.heroDescription')}
        </Box>
        <Box sx={{ display: 'flex', gap: '16px', justifyContent: 'center', flexWrap: 'wrap' }}>
          {isAuthenticated ? (
            <Link href="/dashboard">
              <Button variant="contained" disableElevation sx={{ px: '32px', py: '16px', fontSize: '16px' }}>
                {t('landing.goToDashboard')}
              </Button>
            </Link>
          ) : (
            <>
              <Link href="/register">
                <Button variant="contained" disableElevation sx={{ px: '32px', py: '16px', fontSize: '16px' }}>
                  {t('landing.startForFree')}
                </Button>
              </Link>
              <Link href="#pricing">
                <Button variant="outlined" sx={{ px: '32px', py: '16px', fontSize: '16px', bgcolor: '#f8fafc', '&:hover': { bgcolor: '#f1f5f9' } }}>
                  {t('landing.viewPricing')}
                </Button>
              </Link>
            </>
          )}
        </Box>
      </Box>

      {/* About / Features */}
      <Box
        sx={{
          bgcolor: 'grey.100',
          py: '48px',
        }}
      >
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: { xs: '16px', sm: '24px' } }}>
          <Typography
            component="h3"
            sx={{
              fontSize: { xs: '20px', sm: '24px' },
              fontWeight: 700,
              textAlign: 'center',
              mb: '8px',
              color: 'grey.800',
            }}
          >
            {t('landing.featuresTitle')}
          </Typography>
          <Box
            sx={{
              display: 'block',
              fontSize: { xs: '14px', sm: '16px' },
              color: 'grey.500',
              textAlign: 'center',
              maxWidth: 600,
              mx: 'auto',
              mb: '48px',
            }}
          >
            {t('landing.featuresSubtitle')}
          </Box>
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(4, 1fr)' },
              gap: '24px',
            }}
          >
            {FEATURES.map((f) => (
              <Paper
                key={f.titleKey}
                sx={{
                  p: '24px',
                  borderRadius: '12px',
                  border: '1px solid',
                  borderColor: 'grey.200',
                  bgcolor: 'white',
                  textAlign: 'center',
                  zIndex: 0,
                }}
              >
                <Box sx={{ mb: '8px' }}><f.icon size={36} /></Box>
                <Typography component="h4" sx={{ fontSize: '16px', mb: '8px', color: 'grey.800' }}>
                  {t(f.titleKey)}
                </Typography>
                <Box sx={{ fontSize: '14px', color: 'grey.500', lineHeight: 1.6 }}>
                  {t(f.descKey)}
                </Box>
              </Paper>
            ))}
          </Box>
        </Box>
      </Box>

      {/* Screenshots showcase */}
      <ScreenshotsSection />

      {/* Pricing */}
      <Box
        id="pricing"
        sx={{
          py: '48px',
        }}
      >
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: { xs: '16px', sm: '24px' } }}>
          <Typography
            component="h3"
            sx={{
              fontSize: { xs: '20px', sm: '24px' },
              fontWeight: 700,
              textAlign: 'center',
              mb: '16px',
              color: 'grey.800',
            }}
          >
            {t('landing.pricingTitle')}
          </Typography>
          <Box sx={{ display: 'flex', justifyContent: 'center', flexWrap: 'wrap', gap: '24px' }}>
            {plansLoading
              ? Array.from({ length: 2 }).map((_, i) => (
                  <Skeleton key={i} variant="rectangular" width={320} height={380} sx={{ borderRadius: '12px' }} />
                ))
              : plans.map((plan) => (
                  <Paper
                    key={plan.id}
                    sx={{
                      p: '32px',
                      borderRadius: '12px',
                      border: '2px solid',
                      borderColor: 'grey.200',
                      bgcolor: 'white',
                      width: 320,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <Typography component="h4" sx={{ fontSize: '18px', mb: '4px', color: 'grey.800' }}>
                      {plan.display_name}
                    </Typography>
                    <Box sx={{ display: 'flex', alignItems: 'baseline', mb: '8px' }}>
                      <Box sx={{ fontSize: '24px', fontWeight: 700, color: 'grey.800' }}>
                        {formatPriceIDR(plan.price_idr)}
                      </Box>
                      <Box sx={{ fontSize: '14px', color: 'grey.500', ml: '4px' }}>
                        {t('landing.pricingPeriod')}
                      </Box>
                    </Box>
                    <Box sx={{ fontSize: '14px', color: 'grey.500', mb: '24px' }}>
                      {locale === 'id' ? plan.description_id : plan.description_en}
                    </Box>
                    <Box component="ul" sx={{ listStyle: 'none', p: 0, m: 0, mb: '24px', flex: 1 }}>
                      {[
                        t('landing.pricingFeature1'),
                        t('landing.pricingFeature2'),
                        t('landing.pricingFeature3'),
                        t('landing.pricingFeature4'),
                      ].map((feature) => (
                        <Box
                          key={feature}
                          component="li"
                          sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            py: '8px',
                            borderBottom: '1px solid',
                            borderColor: 'grey.200',
                            fontSize: '14px',
                            color: 'grey.800',
                          }}
                        >
                          <Check size={16} color="green" />
                          {feature}
                        </Box>
                      ))}
                      <Box
                        component="li"
                        sx={{ display: 'flex', alignItems: 'center', gap: '8px', py: '8px', fontSize: '14px', color: 'grey.800' }}
                      >
                        <Users size={16} />
                        {t('landing.pricingMaxUsers', { count: plan.max_users })}
                      </Box>
                    </Box>
                    <Box sx={{ mt: 'auto' }}>
                      <Link href="/register" style={{ display: 'block' }}>
                        <Button variant="contained" disableElevation sx={{ width: '100%', py: '16px' }}>
                          {t('landing.pricingCta')}
                        </Button>
                      </Link>
                    </Box>
                  </Paper>
                ))}
          </Box>
        </Box>
      </Box>

      {/* CTA */}
      <Box sx={{ maxWidth: 1000, mx: 'auto', px: { xs: '16px', sm: '24px' }, py: '48px' }}>
        <Paper
          sx={{
            position: 'relative',
            overflow: 'hidden',
            p: '48px',
            borderRadius: '12px',
            bgcolor: 'primary.main',
            border: 'none',
            boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)',
            textAlign: 'center',
          }}
        >
          <BlobShape
            variant="light"
            sx={{
              width: 180,
              height: 180,
              top: -40,
              right: -30,
            }}
          />
          <BlobShape
            variant="light"
            sx={{
              width: 240,
              height: 240,
              top: 120,
              right: 100,
            }}
          />
          <BlobShape
            variant="light"
            sx={{
              width: 520,
              height: 480,
              bottom: -240,
              left: -120,
            }}
          />
          <Box sx={{ position: 'relative', zIndex: 1 }}>
            <Typography
              component="h3"
              sx={{
                fontSize: { xs: '20px', sm: '24px' },
                fontWeight: 700,
                color: 'white',
                mb: '8px',
              }}
            >
              {t('landing.ctaTitle')}
            </Typography>
            <Box sx={{ fontSize: { xs: '14px', sm: '16px' }, color: 'rgba(255,255,255,0.9)', mb: '24px', maxWidth: 560, mx: 'auto', display: 'block' }}>
              {isAuthenticated ? t('landing.ctaSignedIn') : t('landing.ctaJoin')}
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'center' }}>
              <Link href={isAuthenticated ? '/dashboard' : '/register'}>
                <Button
                  sx={{
                    bgcolor: 'white',
                    color: 'primary.main',
                    px: '32px',
                    py: '16px',
                    fontSize: '16px',
                    fontWeight: 600,
                    border: 'none',
                    borderRadius: '8px',
                    cursor: 'pointer',
                    '&:hover': { bgcolor: 'grey.100', color: 'primary.main' },
                  }}
                >
                  {isAuthenticated ? t('landing.goToDashboard') : t('landing.ctaButton')}
                </Button>
              </Link>
            </Box>
          </Box>
        </Paper>
      </Box>

      {/* Footer */}
      <Box
        component="footer"
        sx={{
          py: '24px',
          borderTop: '1px solid',
          borderColor: 'grey.200',
          bgcolor: 'grey.100',
        }}
      >
        <Box
          sx={{
            display: 'flex',
            maxWidth: 1200,
            mx: 'auto',
            px: { xs: '16px', sm: '24px' },
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '16px',
          }}
        >
          <Box sx={{ fontSize: '14px', color: 'grey.500' }}>{t('landing.footer')}</Box>
        </Box>
      </Box>
    </Box>
  )
}
