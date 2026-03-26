"use client"

import { useState, useCallback, useEffect } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { Box, Button, Card, IconButton, Typography, OutlinedInput } from '@mui/material'
import { ThemeProvider } from '@mui/material/styles'
import { createAppTheme } from '@/theme'
import { Plus, Minus, ImageIcon, ChevronUp, ChevronDown, ShoppingCart } from 'lucide-react'
import { useTranslations, useLocale } from 'next-intl'
import { api, resolveImageURL } from '@/utils/api'
import RecapoLogoText from '@/components/ui/RecapoLogoText'
import { useChangeLocale } from '@/hooks/useLocale'

type Product = {
  id: number
  name: string
  description?: string
  price: number
  image_url?: string | null
}

function SharePageHeader() {
  const locale = useLocale()
  const changeLocale = useChangeLocale()
  return (
    <Box
      component="header"
      sx={{
        position: 'sticky',
        top: 0,
        zIndex: 10,
        bgcolor: 'rgba(255,255,255,0.9)',
        backdropFilter: 'saturate(180%) blur(10px)',
        borderBottom: '1px solid',
        borderColor: 'divider',
      }}
    >
      <Box
        sx={{
          mx: 'auto',
          px: { xs: 3, sm: 4 },
          py: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Link href="/" style={{ textDecoration: 'none', display: 'flex', alignItems: 'center' }}>
          <RecapoLogoText />
        </Link>
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <Box
            component="button"
            onClick={() => changeLocale('en')}
            sx={{
              bgcolor: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: locale === 'en' ? 600 : 400,
              color: locale === 'en' ? 'primary.main' : 'text.secondary',
              '&:hover': { color: 'primary.main' },
            }}
          >
            EN
          </Box>
          <Typography sx={{ color: 'divider' }}>|</Typography>
          <Box
            component="button"
            onClick={() => changeLocale('id')}
            sx={{
              bgcolor: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: locale === 'id' ? 600 : 400,
              color: locale === 'id' ? 'primary.main' : 'text.secondary',
              '&:hover': { color: 'primary.main' },
            }}
          >
            ID
          </Box>
        </Box>
      </Box>
    </Box>
  )
}

export default function SharePage() {
  const t = useTranslations('common')
  const tShare = useTranslations('share')
  const tErrors = useTranslations('errors')
  const params = useParams()
  const shareToken = params?.shareToken as string

  const { data: productsRes, isLoading, isError, error } = useQuery(
    ['publicProducts', shareToken],
    async () => {
      if (!shareToken) return null
      const res = await api.getPublicProducts(shareToken)
      if (!res.success) throw new Error(res.message || tShare('fetchFailed'))
      return res.data as Product[]
    },
    { enabled: !!shareToken }
  )

  const products = productsRes || []

  const [zoomedImage, setZoomedImage] = useState<{ src: string; alt: string } | null>(null)

  useEffect(() => {
    if (!zoomedImage) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setZoomedImage(null) }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [zoomedImage])

  const [quantities, setQuantities] = useState<Record<number, number>>({})
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [orderSubmitting, setOrderSubmitting] = useState(false)
  const [orderError, setOrderError] = useState<string | null>(null)
  const [orderSuccess, setOrderSuccess] = useState(false)
  const [mobileOrderOpen, setMobileOrderOpen] = useState(false)

  const setQty = useCallback((productId: number, qty: number) => {
    setQuantities((prev) => ({ ...prev, [productId]: Math.max(0, qty) }))
  }, [])

  const summaryItems = products.filter((p) => (quantities[p.id] ?? 0) > 0)
  const totalPrice = summaryItems.reduce((sum, p) => sum + p.price * (quantities[p.id] ?? 0), 0)

  const handlePlaceOrder = async () => {
    if (summaryItems.length === 0 || !customerName.trim() || !customerPhone.trim()) return
    setOrderError(null)
    setOrderSubmitting(true)
    try {
      const res = await api.createPublicOrderTemp(shareToken, {
        customer_name: customerName.trim(),
        customer_phone: customerPhone.trim(),
        order_items: summaryItems.map((p) => ({ product_id: p.id, qty: quantities[p.id] ?? 0 })),
      })
      if (res.success) {
        setOrderSuccess(true)
        setQuantities({})
        setCustomerName('')
        setCustomerPhone('')
        setMobileOrderOpen(false)
      } else {
        setOrderError(res.message || tShare('orderFailed'))
      }
    } catch (err: unknown) {
      setOrderError(err instanceof Error ? err.message : tShare('orderFailed'))
    } finally {
      setOrderSubmitting(false)
    }
  }

  const orderSummaryContent = (
    <>
      {summaryItems.length === 0 ? (
        <Typography sx={{ color: 'text.secondary', fontSize: '0.875rem', mb: 3, display: 'block' }}>{tShare('noItemsInSummary')}</Typography>
      ) : (
        <>
          <Box component="ul" sx={{ listStyle: 'none', p: 0, m: 0, mb: 3 }}>
            {summaryItems.map((p) => (
              <Box component="li" key={p.id} sx={{ py: 2, borderBottom: '1px solid', borderColor: 'divider' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 2 }}>
                  <Typography sx={{ fontSize: '0.875rem', fontWeight: 500 }}>{p.name}</Typography>
                  <Typography sx={{ fontSize: '0.875rem', color: 'text.secondary', whiteSpace: 'nowrap' }}>
                    {(quantities[p.id] ?? 0)} × Rp. {p.price.toLocaleString()}
                  </Typography>
                </Box>
                <Typography sx={{ fontSize: '0.875rem', color: 'primary.main', fontWeight: 600, mt: 1 }}>
                  Rp. {(p.price * (quantities[p.id] ?? 0)).toLocaleString()}
                </Typography>
              </Box>
            ))}
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
            <Typography sx={{ fontWeight: 700 }}>{tShare('total')}</Typography>
            <Typography sx={{ fontWeight: 700, color: 'primary.main' }}>Rp. {totalPrice.toLocaleString()}</Typography>
          </Box>
        </>
      )}
      <Box sx={{ mb: 2 }}>
        <Typography component="label" sx={{ display: 'block', fontSize: '0.875rem', mb: 1, color: 'text.secondary' }}>
          {tShare('customerName')}
        </Typography>
        <OutlinedInput
          value={customerName}
          onChange={(e) => setCustomerName(e.target.value)}
          placeholder={tShare('customerName')}
          fullWidth
          size="small"
        />
      </Box>
      <Box sx={{ mb: 3 }}>
        <Typography component="label" sx={{ display: 'block', fontSize: '0.875rem', mb: 1, color: 'text.secondary' }}>
          {tShare('customerPhone')}
        </Typography>
        <OutlinedInput
          value={customerPhone}
          onChange={(e) => setCustomerPhone(e.target.value)}
          placeholder={tShare('customerPhone')}
          type="tel"
          fullWidth
          size="small"
        />
      </Box>
      {orderSuccess && (
        <Typography sx={{ color: 'success.main', fontSize: '0.875rem', mb: 2 }}>{tShare('orderSuccess')}</Typography>
      )}
      {orderError && (
        <Typography sx={{ color: 'error.main', fontSize: '0.875rem', mb: 2 }}>{orderError}</Typography>
      )}
      <Button
        variant="contained"
        onClick={handlePlaceOrder}
        disabled={summaryItems.length === 0 || !customerName.trim() || !customerPhone.trim() || orderSubmitting}
        fullWidth
        size="large"
      >
        {orderSubmitting ? t('loading') : tShare('placeOrder')}
      </Button>
    </>
  )

  if (!shareToken) {
    return (
      <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
        <SharePageHeader />
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: { xs: 3, sm: 4 }, py: 6, textAlign: 'center' }}>
          <Typography sx={{ color: 'error.main' }}>{tShare('invalidLink')}</Typography>
        </Box>
      </Box>
    )
  }

  if (isLoading) {
    return (
      <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
        <SharePageHeader />
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: { xs: 3, sm: 4 }, py: 6, textAlign: 'center' }}>
          <Typography sx={{ color: 'text.secondary' }}>{t('loading')}</Typography>
        </Box>
      </Box>
    )
  }

  if (isError) {
    const err = error as Error & { status?: number }
    const is404 = err?.message?.toLowerCase().includes('not found') || err?.status === 404
    return (
      <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
        <SharePageHeader />
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: { xs: 3, sm: 4 }, py: 6, textAlign: 'center' }}>
          <Typography sx={{ fontWeight: 600, mb: 1, display: 'block' }}>
            {is404 ? tShare('shopNotFound') : tShare('errorTitle')}
          </Typography>
          <Typography sx={{ color: 'text.secondary' }}>
            {is404 ? tShare('shopNotFoundDesc') : (error as Error)?.message || tErrors('loadingError', { resource: tShare('products') })}
          </Typography>
        </Box>
      </Box>
    )
  }

  return (
    <ThemeProvider theme={createAppTheme('light')}>
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
      <SharePageHeader />

      <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
        {/* Scrollable product area */}
        <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', mx: 'auto', px: { xs: 2, sm: 4 }, width: '100%' }}>
          <Box sx={{ mt: { xs: 3, sm: 4 }, display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, alignItems: 'flex-start', gap: 4 }}>
            <Box sx={{ flex: 1, minWidth: 0 }}>
              <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.5, fontSize: { xs: '1.25rem', sm: '1.5rem' }, color: 'text.primary' }}>
                {tShare('title')}
              </Typography>
              <Typography sx={{ color: 'text.secondary', mb: 3, display: 'block', fontSize: { xs: '0.875rem', sm: '1rem' } }}>{tShare('description')}</Typography>
              {products.length === 0 ? (
                <Typography sx={{ color: 'text.secondary' }}>{tShare('noProducts')}</Typography>
              ) : (
                <Box
                  sx={{
                    display: 'grid',
                    gridTemplateColumns: { xs: 'repeat(2, 1fr)', sm: 'repeat(auto-fill, minmax(200px, 1fr))' },
                    gap: { xs: 1.5, sm: 3 },
                  }}
                >
                  {products.map((product) => (
                    <Card
                      key={product.id}
                      variant="outlined"
                      sx={{
                        p: 0,
                        borderRadius: 3,
                        bgcolor: 'white',
                        overflow: 'hidden',
                        border: '1px solid',
                        borderColor: 'grey.200',
                        transition: 'box-shadow 0.2s ease, transform 0.2s ease',
                        '&:hover': {
                          boxShadow: '0 6px 24px rgba(0,0,0,0.09)',
                          transform: 'translateY(-2px)',
                        },
                      }}
                    >
                      <Box
                        sx={{
                          width: '100%',
                          aspectRatio: '1/1',
                          background: 'linear-gradient(135deg, rgb(92, 151, 245) 0%, rgb(26, 94, 239) 100%)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          overflow: 'hidden',
                        }}
                      >
                        {product.image_url ? (
                          <img
                            src={resolveImageURL(product.image_url)}
                            alt={product.name}
                            onClick={() => setZoomedImage({ src: resolveImageURL(product.image_url)!, alt: product.name })}
                            style={{ width: '100%', height: '100%', objectFit: 'cover', cursor: 'zoom-in' }}
                          />
                        ) : (
                          <ImageIcon size={40} color="rgba(255,255,255,0.6)" strokeWidth={1.5} />
                        )}
                      </Box>
                      <Box sx={{ p: { xs: 1.5, sm: 2.5 } }}>
                        <Typography sx={{ fontWeight: 600, fontSize: { xs: '0.8125rem', sm: '0.9375rem' }, mb: 0.5, lineHeight: 1.35 }}>
                          {product.name}
                        </Typography>
                        {product.description ? (
                          <Typography sx={{ color: 'text.secondary', fontSize: '0.75rem', mb: 1, lineHeight: 1.4, display: { xs: 'none', sm: 'block' } }}>
                            {product.description}
                          </Typography>
                        ) : null}
                        <Typography sx={{ color: 'primary.main', fontWeight: 700, fontSize: { xs: '0.875rem', sm: '1rem' }, mb: { xs: 1.5, sm: 2 } }}>
                          Rp. {product.price.toLocaleString()}
                        </Typography>
                        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                          <Typography sx={{ fontSize: '0.8125rem', color: 'text.secondary', fontWeight: 500, display: { xs: 'none', sm: 'block' } }}>Qty</Typography>
                          <Box
                            sx={{
                              display: 'flex',
                              alignItems: 'center',
                              border: '1px solid',
                              borderColor: 'grey.300',
                              borderRadius: 2,
                              overflow: 'hidden',
                              width: { xs: '100%', sm: 'auto' },
                            }}
                          >
                            <IconButton
                              size="small"
                              onClick={() => setQty(product.id, (quantities[product.id] ?? 0) - 1)}
                              sx={{ borderRadius: 0, px: { xs: 1.5, sm: 0.75 }, py: { xs: 1, sm: 0.5 }, '&:hover': { bgcolor: 'grey.100' } }}
                            >
                              <Minus size={16} />
                            </IconButton>
                            <OutlinedInput
                              type="number"
                              inputProps={{ min: 0 }}
                              value={quantities[product.id] ?? 0}
                              onChange={(e) => setQty(product.id, parseInt(e.target.value, 10) || 0)}
                              size="small"
                              sx={{
                                flex: { xs: 1, sm: 'none' },
                                width: { xs: 'auto', sm: 44 },
                                fontSize: '0.875rem',
                                fontWeight: 600,
                                '& fieldset': { border: 'none' },
                                '& input': { textAlign: 'center', py: { xs: 0.75, sm: 0.5 }, px: 0.5 },
                                '& input::-webkit-outer-spin-button, & input::-webkit-inner-spin-button': { display: 'none' },
                              }}
                            />
                            <IconButton
                              size="small"
                              onClick={() => setQty(product.id, (quantities[product.id] ?? 0) + 1)}
                              sx={{ borderRadius: 0, px: { xs: 1.5, sm: 0.75 }, py: { xs: 1, sm: 0.5 }, '&:hover': { bgcolor: 'grey.100' } }}
                            >
                              <Plus size={16} />
                            </IconButton>
                          </Box>
                        </Box>
                      </Box>
                    </Card>
                  ))}
                </Box>
              )}
            </Box>

            {/* Desktop order summary */}
            <Card
              variant="outlined"
              sx={{
                display: { xs: 'none', sm: 'block' },
                width: 440,
                flexShrink: 0,
                alignSelf: 'flex-start',
                p: 4,
                borderRadius: 2,
                bgcolor: 'white',
                position: 'sticky',
                top: '24px',
              }}
            >
              <Typography sx={{ fontWeight: 700, fontSize: '1.25rem', mb: 3, display: 'block' }}>{tShare('orderSummary')}</Typography>
              {orderSummaryContent}
            </Card>
          </Box>

          {/* Bottom spacer so last products aren't hidden behind mobile panel */}
          <Box sx={{ height: { xs: 16, sm: 24 } }} />
        </Box>

        {/* Mobile sticky order summary panel */}
        <Box
          sx={{
            display: { xs: 'block', sm: 'none' },
            borderTop: '1px solid',
            borderColor: 'divider',
            bgcolor: 'white',
            flexShrink: 0,
            boxShadow: '0 -4px 16px rgba(0,0,0,0.08)',
          }}
        >
          {/* Collapsed bar — always visible */}
          <Box
            onClick={() => setMobileOrderOpen((v) => !v)}
            sx={{
              px: 2,
              py: 1.5,
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
              cursor: 'pointer',
              userSelect: 'none',
            }}
          >
            <Box sx={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
              <ShoppingCart size={22} />
              {summaryItems.length > 0 && (
                <Box
                  sx={{
                    position: 'absolute',
                    top: -6,
                    right: -8,
                    bgcolor: 'primary.main',
                    color: 'white',
                    borderRadius: '50%',
                    width: 16,
                    height: 16,
                    fontSize: '0.65rem',
                    fontWeight: 700,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {summaryItems.length}
                </Box>
              )}
            </Box>
            <Typography sx={{ fontWeight: 600, flex: 1, fontSize: '0.9375rem' }}>
              {tShare('orderSummary')}
            </Typography>
            {summaryItems.length > 0 && (
              <Typography sx={{ fontWeight: 700, color: 'primary.main', fontSize: '0.9375rem' }}>
                Rp. {totalPrice.toLocaleString()}
              </Typography>
            )}
            {mobileOrderOpen ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
          </Box>

          {/* Expanded form */}
          {mobileOrderOpen && (
            <Box
              sx={{
                px: 2,
                pb: 2,
                maxHeight: '55vh',
                overflowY: 'auto',
                borderTop: '1px solid',
                borderColor: 'divider',
              }}
            >
              <Box sx={{ pt: 2 }}>
                {orderSummaryContent}
              </Box>
            </Box>
          )}
        </Box>
      </Box>

      {zoomedImage && (
        <Box
          onClick={() => setZoomedImage(null)}
          sx={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            bgcolor: 'rgba(0,0,0,0.85)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'zoom-out',
          }}
        >
          <img
            src={zoomedImage.src}
            alt={zoomedImage.alt}
            onClick={(e) => e.stopPropagation()}
            style={{ maxWidth: '90vw', maxHeight: '90vh', objectFit: 'contain', borderRadius: 8 }}
          />
          <Box
            component="button"
            onClick={() => setZoomedImage(null)}
            sx={{
              position: 'absolute',
              top: 16,
              right: 16,
              bgcolor: 'rgba(255,255,255,0.15)',
              border: 'none',
              borderRadius: '50%',
              width: 36,
              height: 36,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              color: 'white',
              fontSize: '1.25rem',
              lineHeight: 1,
              '&:hover': { bgcolor: 'rgba(255,255,255,0.3)' },
            }}
          >
            ×
          </Box>
        </Box>
      )}
    </Box>
    </ThemeProvider>
  )
}
