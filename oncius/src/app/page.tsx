"use client"

import Link from 'next/link'
import { Box, Heading, Text, Button, Flex, Card } from 'theme-ui'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import LandingHeader from '@/components/LandingHeader'

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
      sx={{
        position: 'absolute',
        pointerEvents: 'none',
        zIndex: 0,
        ...sx,
      }}
      as="span"
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

const FEATURES = [
  { titleKey: 'landing.featureOrderManagement', descKey: 'landing.featureOrderManagementDesc', icon: '📦' },
  { titleKey: 'landing.featureProductCatalog', descKey: 'landing.featureProductCatalogDesc', icon: '🏷️' },
  { titleKey: 'landing.featureCustomerDatabase', descKey: 'landing.featureCustomerDatabaseDesc', icon: '👥' },
  { titleKey: 'landing.featureDashboardReports', descKey: 'landing.featureDashboardReportsDesc', icon: '📊' },
]

export default function HomePage() {
  const t = useTranslations()
  const { isAuthenticated, isLoadingUser } = useAuth()

  if (isLoadingUser) {
    return (
      <Box sx={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Text>{t('common.loading')}</Text>
      </Box>
    )
  }

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

      <LandingHeader />

      {/* Hero */}
      <Box
        sx={{
          position: 'relative',
          maxWidth: 1200,
          mx: 'auto',
          px: [3, 4],
          py: [6, 8],
          textAlign: 'center',
        }}
      >
        <Heading
          as="h2"
          sx={{
            fontSize: [5, 6, 7],
            fontWeight: 700,
            lineHeight: 1.2,
            mb: 3,
            color: 'text',
          }}
        >
          {t('landing.heroTitle')}{' '}
          <Box as="span" sx={{ color: 'primary' }}>
            {t('landing.heroHighlight')}
          </Box>
        </Heading>
        <Text
          sx={{
            display: 'block',
            fontSize: [2, 3],
            color: 'text.secondary',
            maxWidth: 840,
            mx: 'auto',
            mb: 3,
            lineHeight: 1.6,
          }}
        >
          {t('landing.heroDescription')}
        </Text>
        <Flex sx={{ gap: 3, justifyContent: 'center', flexWrap: 'wrap' }}>
          {isAuthenticated ? (
            <Link href="/dashboard">
              <Button variant="primary" sx={{ px: 5, py: 3, fontSize: 2 }}>
                {t('landing.goToDashboard')}
              </Button>
            </Link>
          ) : (
            <>
              <Link href="/register">
                <Button variant="primary" sx={{ px: 5, py: 3, fontSize: 2 }}>
                  {t('landing.startForFree')}
                </Button>
              </Link>
              <Link href="#pricing">
                <Button variant="secondary" sx={{ px: 5, py: 3, fontSize: 2 }}>
                  {t('landing.viewPricing')}
                </Button>
              </Link>
            </>
          )}
        </Flex>
      </Box>

      {/* About / Features */}
      <Box
        sx={{
          bg: 'backgroundLight',
          py: 6,
        }}
      >
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: [3, 4] }}>
          <Heading
            as="h3"
            sx={{
              fontSize: [4, 5],
              fontWeight: 700,
              textAlign: 'center',
              mb: 2,
              color: 'text',
            }}
          >
            {t('landing.featuresTitle')}
          </Heading>
          <Text
            sx={{
              display: 'block',
              fontSize: [1, 2],
              color: 'text.secondary',
              textAlign: 'center',
              maxWidth: 600,
              mx: 'auto',
              mb: 6,
            }}
          >
            {t('landing.featuresSubtitle')}
          </Text>
          <Flex
            sx={{
              display: 'grid',
              gridTemplateColumns: ['1fr', '1fr 1fr', 'repeat(4, 1fr)'],
              gap: 4,
            }}
          >
            {FEATURES.map((f) => (
              <Card
                key={f.titleKey}
                sx={{
                  p: 4,
                  borderRadius: 'large',
                  border: '1px solid',
                  borderColor: 'border',
                  bg: 'white',
                  textAlign: 'center',
                  zIndex: 0,
                }}
              >
                <Text sx={{ fontSize: 5, mb: 2 }}>{f.icon}</Text>
                <Heading as="h4" sx={{ fontSize: 2, mb: 2, color: 'text' }}>
                  {t(f.titleKey)}
                </Heading>
                <Text sx={{ fontSize: 1, color: 'text.secondary', lineHeight: 1.6 }}>
                  {t(f.descKey)}
                </Text>
              </Card>
            ))}
          </Flex>
        </Box>
      </Box>

      {/* Pricing */}
      <Box
        id="pricing"
        sx={{
          py: 6,
        }}
      >
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: [3, 4] }}>
          <Heading
            as="h3"
            sx={{
              fontSize: [4, 5],
              fontWeight: 700,
              textAlign: 'center',
              mb: 3,
              color: 'text',
            }}
          >
            {t('landing.pricingTitle')}
          </Heading>
          <Flex sx={{ justifyContent: 'center' }}>
            <Card
              sx={{
                p: 5,
                borderRadius: 'large',
                border: '2px solid',
                borderColor: 'border',
                bg: 'white',
                maxWidth: 360,
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <Heading as="h4" sx={{ fontSize: 3, mb: 1, color: 'text' }}>
                {t('landing.pricingName')}
              </Heading>
              <Flex sx={{ alignItems: 'baseline', mb: 2 }}>
                <Text sx={{ fontSize: 5, fontWeight: 700, color: 'text' }}>{t('landing.pricingPrice')}</Text>
                <Text sx={{ fontSize: 1, color: 'text.secondary', ml: 1 }}>{t('landing.pricingPeriod')}</Text>
              </Flex>
              <Text sx={{ fontSize: 1, color: 'text.secondary', mb: 4 }}>
                {t('landing.pricingDescription')}
              </Text>
              <Box as="ul" sx={{ listStyle: 'none', p: 0, m: 0, mb: 4, flex: 1 }}>
                {[
                  t('landing.pricingFeature1'),
                  t('landing.pricingFeature2'),
                  t('landing.pricingFeature3'),
                  t('landing.pricingFeature4'),
                ].map((feature) => (
                  <Flex
                    key={feature}
                    as="li"
                    sx={{
                      alignItems: 'center',
                      gap: 2,
                      py: 2,
                      borderBottom: '1px solid',
                      borderColor: 'border',
                      fontSize: 1,
                      color: 'text',
                    }}
                  >
                    <Text sx={{ color: 'success' }}>✓</Text>
                    {feature}
                  </Flex>
                ))}
              </Box>
              <Box sx={{ mt: 'auto' }}>
                <Link href="/register" style={{ display: 'block' }}>
                  <Button variant="primary" sx={{ width: '100%', py: 3 }}>
                    {t('landing.pricingCta')}
                  </Button>
                </Link>
              </Box>
            </Card>
          </Flex>
        </Box>
      </Box>

      {/* CTA */}
      <Box sx={{ maxWidth: 1000, mx: 'auto', px: [3, 4], py: 6 }}>
        <Card
          sx={{
            position: 'relative',
            overflow: 'hidden',
            p: 6,
            borderRadius: 'large',
            bg: 'primary',
            border: 'none',
            boxShadow: 'medium',
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
          <Heading
            as="h3"
            sx={{
              fontSize: [4, 5],
              fontWeight: 700,
              color: 'white',
              mb: 2,
            }}
          >
            {t('landing.ctaTitle')}
          </Heading>
          <Text sx={{ fontSize: [1, 2], color: 'rgba(255,255,255,0.9)', mb: 4, maxWidth: 560, mx: 'auto', display: 'block' }}>
            {isAuthenticated ? t('landing.ctaSignedIn') : t('landing.ctaJoin')}
          </Text>
          <Flex sx={{ justifyContent: 'center' }}>
            <Link href={isAuthenticated ? '/dashboard' : '/register'}>
              <Button
                sx={{
                  bg: 'white',
                  color: 'primary',
                  px: 5,
                  py: 3,
                  fontSize: 2,
                  fontWeight: 600,
                  border: 'none',
                  borderRadius: 'medium',
                  cursor: 'pointer',
                  '&:hover': { bg: 'backgroundLight', color: 'primary' },
                }}
              >
                {isAuthenticated ? t('landing.goToDashboard') : t('landing.ctaButton')}
              </Button>
            </Link>
          </Flex>
          </Box>
        </Card>
      </Box>

      {/* Footer */}
      <Box
        as="footer"
        sx={{
          py: 4,
          borderTop: '1px solid',
          borderColor: 'border',
          bg: 'backgroundLight',
        }}
      >
        <Flex
          sx={{
            maxWidth: 1200,
            mx: 'auto',
            px: [3, 4],
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: 3,
          }}
        >
          <Text sx={{ fontSize: 1, color: 'text.secondary' }}>{t('landing.footer')}</Text>
        </Flex>
      </Box>
    </Box>
  )
}
