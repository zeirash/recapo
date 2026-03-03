"use client"

import { useState, useCallback, useEffect } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { Box, Button, Card, IconButton, Typography, OutlinedInput } from '@mui/material'
import { Plus, Minus, ImageIcon } from 'lucide-react'
import { useTranslations, useLocale } from 'next-intl'
import { api, resolveImageURL } from '@/utils/api'
import RecapoLogoText from '@/components/RecapoLogoText'
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
      } else {
        setOrderError(res.message || tShare('orderFailed'))
      }
    } catch (err: unknown) {
      setOrderError(err instanceof Error ? err.message : tShare('orderFailed'))
    } finally {
      setOrderSubmitting(false)
    }
  }

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
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
      <SharePageHeader />
      <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', mx: 'auto', px: { xs: 3, sm: 4 }, width: '100%' }}>
        <Box sx={{ mt: 4, display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, alignItems: 'flex-start', gap: 4 }}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
              {tShare('title')}
            </Typography>
            <Typography sx={{ color: 'text.secondary', mb: 4, display: 'block' }}>{tShare('description')}</Typography>
            {products.length === 0 ? (
              <Typography sx={{ color: 'text.secondary' }}>{tShare('noProducts')}</Typography>
            ) : (
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                  gap: 3,
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
                        <ImageIcon size={48} color="rgba(255,255,255,0.6)" strokeWidth={1.5} />
                      )}
                    </Box>
                    <Box sx={{ p: 2.5 }}>
                      <Typography sx={{ fontWeight: 600, fontSize: '0.9375rem', mb: 0.5, lineHeight: 1.35 }}>
                        {product.name}
                      </Typography>
                      {product.description ? (
                        <Typography sx={{ color: 'text.secondary', fontSize: '0.8125rem', mb: 1.5, lineHeight: 1.4 }}>
                          {product.description}
                        </Typography>
                      ) : null}
                      <Typography sx={{ color: 'primary.main', fontWeight: 700, fontSize: '1rem', mb: 2 }}>
                        Rp. {product.price.toLocaleString()}
                      </Typography>
                      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <Typography sx={{ fontSize: '0.8125rem', color: 'text.secondary', fontWeight: 500 }}>Qty</Typography>
                        <Box
                          sx={{
                            display: 'flex',
                            alignItems: 'center',
                            border: '1px solid',
                            borderColor: 'grey.300',
                            borderRadius: 2,
                            overflow: 'hidden',
                          }}
                        >
                          <IconButton
                            size="small"
                            onClick={() => setQty(product.id, (quantities[product.id] ?? 0) - 1)}
                            sx={{ borderRadius: 0, px: 0.75, py: 0.5, '&:hover': { bgcolor: 'grey.100' } }}
                          >
                            <Minus size={14} />
                          </IconButton>
                          <OutlinedInput
                            type="number"
                            inputProps={{ min: 0 }}
                            value={quantities[product.id] ?? 0}
                            onChange={(e) => setQty(product.id, parseInt(e.target.value, 10) || 0)}
                            size="small"
                            sx={{
                              width: 44,
                              fontSize: '0.875rem',
                              fontWeight: 600,
                              '& fieldset': { border: 'none' },
                              '& input': { textAlign: 'center', py: 0.5, px: 0.5 },
                              '& input::-webkit-outer-spin-button, & input::-webkit-inner-spin-button': { display: 'none' },
                            }}
                          />
                          <IconButton
                            size="small"
                            onClick={() => setQty(product.id, (quantities[product.id] ?? 0) + 1)}
                            sx={{ borderRadius: 0, px: 0.75, py: 0.5, '&:hover': { bgcolor: 'grey.100' } }}
                          >
                            <Plus size={14} />
                          </IconButton>
                        </Box>
                      </Box>
                    </Box>
                  </Card>
                ))}
              </Box>
            )}
          </Box>

          <Card
            variant="outlined"
            sx={{
              width: { xs: '100%', sm: 440 },
              flexShrink: 0,
              alignSelf: 'flex-start',
              p: 4,
              borderRadius: 2,
              bgcolor: 'white',
              position: { xs: 'relative', sm: 'sticky' },
              top: { xs: 0, sm: '24px' },
            }}
          >
            <Typography sx={{ fontWeight: 700, fontSize: '1.25rem', mb: 3, display: 'block' }}>{tShare('orderSummary')}</Typography>
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
            >
              {orderSubmitting ? t('loading') : tShare('placeOrder')}
            </Button>
          </Card>
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
  )
}
