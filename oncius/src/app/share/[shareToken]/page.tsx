"use client"

import { useState, useCallback, useEffect } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { Box, Button, Card, Flex, Text, Input } from 'theme-ui'
import { useTranslations, useLocale } from 'next-intl'
import { api, resolveImageURL } from '@/utils/api'
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
      as="header"
      sx={{
        position: 'sticky',
        top: 0,
        zIndex: 10,
        bg: 'rgba(255,255,255,0.9)',
        backdropFilter: 'saturate(180%) blur(10px)',
        borderBottom: '1px solid',
        borderColor: 'border',
      }}
    >
      <Flex
        sx={{
          mx: 'auto',
          px: [3, 4],
          py: 3,
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Link href="/" style={{ textDecoration: 'none' }}>
          <Text as="span" sx={{ fontSize: [3, 4], fontWeight: 700, color: 'primary' }}>
            Recapo
          </Text>
        </Link>
        <Flex sx={{ gap: 2, alignItems: 'center' }}>
          <Box
            as="button"
            onClick={() => changeLocale('en')}
            sx={{
              bg: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: 1,
              fontWeight: locale === 'en' ? 600 : 400,
              color: locale === 'en' ? 'primary' : 'text.secondary',
              '&:hover': { color: 'primary' },
            }}
          >
            EN
          </Box>
          <Text sx={{ color: 'border' }}>|</Text>
          <Box
            as="button"
            onClick={() => changeLocale('id')}
            sx={{
              bg: 'transparent',
              border: 'none',
              cursor: 'pointer',
              fontSize: 1,
              fontWeight: locale === 'id' ? 600 : 400,
              color: locale === 'id' ? 'primary' : 'text.secondary',
              '&:hover': { color: 'primary' },
            }}
          >
            ID
          </Box>
        </Flex>
      </Flex>
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
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: [3, 4], py: 6, textAlign: 'center' }}>
          <Text sx={{ color: 'error' }}>{tShare('invalidLink')}</Text>
        </Box>
      </Box>
    )
  }

  if (isLoading) {
    return (
      <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
        <SharePageHeader />
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: [3, 4], py: 6, textAlign: 'center' }}>
          <Text sx={{ color: 'text.secondary' }}>{t('loading')}</Text>
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
        <Box sx={{ maxWidth: 1200, mx: 'auto', px: [3, 4], py: 6, textAlign: 'center' }}>
          <Text sx={{ fontWeight: 600, mb: 1, display: 'block' }}>
            {is404 ? tShare('shopNotFound') : tShare('errorTitle')}
          </Text>
          <Text sx={{ color: 'text.secondary' }}>
            {is404 ? tShare('shopNotFoundDesc') : (error as Error)?.message || tErrors('loadingError', { resource: tShare('products') })}
          </Text>
        </Box>
      </Box>
    )
  }

  return (
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden', background: 'linear-gradient(180deg, #f8fafc 0%, #f1f5f9 50%, #ffffff 100%)' }}>
      <SharePageHeader />
      <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', mx: 'auto', px: [3, 4], width: '100%' }}>
        <Flex sx={{ mt: 4, flexDirection: ['column', 'row'], alignItems: 'flex-start', gap: 4 }}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
          <Text as="h1" sx={{ fontSize: 4, fontWeight: 700, mb: 1 }}>
            {tShare('title')}
          </Text>
          <Text as="p" sx={{ color: 'text.secondary', mb: 4, display: 'block' }}>{tShare('description')}</Text>
          {products.length === 0 ? (
            <Text sx={{ color: 'text.secondary' }}>{tShare('noProducts')}</Text>
          ) : (
            <Flex
              sx={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                gap: 3,
              }}
            >
              {products.map((product) => (
                <Card
                  key={product.id}
                  sx={{
                    p: 0,
                    borderRadius: 'medium',
                    border: '1px solid',
                    borderColor: 'border',
                    bg: 'white',
                    boxShadow: 'small',
                    overflow: 'hidden',
                  }}
                >
                  <Box
                    sx={{
                      width: '100%',
                      aspectRatio: '1/1',
                      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
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
                      <Text sx={{ fontSize: 5, fontWeight: 700, color: 'white' }}>
                        {product.name.charAt(0).toUpperCase()}
                      </Text>
                    )}
                  </Box>
                  <Box sx={{ p: 3 }}>
                    <Flex sx={{ flexDirection: 'column', justifyContent: 'space-between', minHeight: 60 }}>
                      <Text sx={{ fontWeight: 600, fontSize: 2, mb: 1 }}>{product.name}</Text>
                      {product.description ? (
                        <Text sx={{ color: 'text.secondary', fontSize: 1, mb: 2 }}>{product.description}</Text>
                      ) : null}
                      <Flex sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2, mt: 1 }}>
                        <Text sx={{ color: 'primary', fontWeight: 600 }}>
                          Rp. {product.price.toLocaleString()}
                        </Text>
                        <Box as="label" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Text as="span" sx={{ fontSize: 1, color: 'text.secondary' }}>Qty</Text>
                          <Input
                            type="number"
                            min={0}
                            value={quantities[product.id] ?? 0}
                            onChange={(e) => setQty(product.id, parseInt(e.target.value, 10) || 0)}
                            sx={{
                              width: 56,
                              py: 1,
                              px: 2,
                              fontSize: 1,
                              textAlign: 'center',
                            }}
                          />
                        </Box>
                      </Flex>
                    </Flex>
                  </Box>
                </Card>
              ))}
            </Flex>
          )}
        </Box>

        <Card
          sx={{
            width: ['100%', 440],
            flexShrink: 0,
            alignSelf: 'flex-start',
            p: 4,
            borderRadius: 'medium',
            border: '1px solid',
            borderColor: 'border',
            bg: 'white',
            boxShadow: 'small',
            position: ['relative', 'sticky'],
            top: [0, '24px'],
          }}
        >
          <Text sx={{ fontWeight: 700, fontSize: 3, mb: 3, display: 'block' }}>{tShare('orderSummary')}</Text>
          {summaryItems.length === 0 ? (
            <Text sx={{ color: 'text.secondary', fontSize: 1, mb: 3, display: 'block' }}>{tShare('noItemsInSummary')}</Text>
          ) : (
            <>
              <Box as="ul" sx={{ listStyle: 'none', p: 0, m: 0, mb: 3 }}>
                {summaryItems.map((p) => (
                  <Box as="li" key={p.id} sx={{ py: 2, borderBottom: '1px solid', borderColor: 'border' }}>
                    <Flex sx={{ justifyContent: 'space-between', alignItems: 'flex-start', gap: 2 }}>
                      <Text sx={{ fontSize: 1, fontWeight: 500 }}>{p.name}</Text>
                      <Text sx={{ fontSize: 1, color: 'text.secondary', whiteSpace: 'nowrap' }}>
                        {(quantities[p.id] ?? 0)} × Rp. {p.price.toLocaleString()}
                      </Text>
                    </Flex>
                    <Text sx={{ fontSize: 1, color: 'primary', fontWeight: 600, mt: 1 }}>
                      Rp. {(p.price * (quantities[p.id] ?? 0)).toLocaleString()}
                    </Text>
                  </Box>
                ))}
              </Box>
              <Flex sx={{ justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                <Text sx={{ fontWeight: 700 }}>{tShare('total')}</Text>
                <Text sx={{ fontWeight: 700, color: 'primary' }}>Rp. {totalPrice.toLocaleString()}</Text>
              </Flex>
            </>
          )}
          <Box sx={{ mb: 2 }}>
            <Text as="label" sx={{ display: 'block', fontSize: 1, mb: 1, color: 'text.secondary' }}>
              {tShare('customerName')}
            </Text>
            <Input
              value={customerName}
              onChange={(e) => setCustomerName(e.target.value)}
              placeholder={tShare('customerName')}
              sx={{ width: '100%' }}
            />
          </Box>
          <Box sx={{ mb: 3 }}>
            <Text as="label" sx={{ display: 'block', fontSize: 1, mb: 1, color: 'text.secondary' }}>
              {tShare('customerPhone')}
            </Text>
            <Input
              value={customerPhone}
              onChange={(e) => setCustomerPhone(e.target.value)}
              placeholder={tShare('customerPhone')}
              type="tel"
              sx={{ width: '100%' }}
            />
          </Box>
          {orderSuccess && (
            <Text sx={{ color: 'success', fontSize: 1, mb: 2 }}>{tShare('orderSuccess')}</Text>
          )}
          {orderError && (
            <Text sx={{ color: 'error', fontSize: 1, mb: 2 }}>{orderError}</Text>
          )}
          <Button
            onClick={handlePlaceOrder}
            disabled={summaryItems.length === 0 || !customerName.trim() || !customerPhone.trim() || orderSubmitting}
            sx={{ width: '100%' }}
          >
            {orderSubmitting ? t('loading') : tShare('placeOrder')}
          </Button>
        </Card>
        </Flex>
      </Box>

      {zoomedImage && (
        <Box
          onClick={() => setZoomedImage(null)}
          sx={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            bg: 'rgba(0,0,0,0.85)',
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
            as="button"
            onClick={() => setZoomedImage(null)}
            sx={{
              position: 'absolute',
              top: 16,
              right: 16,
              bg: 'rgba(255,255,255,0.15)',
              border: 'none',
              borderRadius: '50%',
              width: 36,
              height: 36,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              color: 'white',
              fontSize: 3,
              lineHeight: 1,
              '&:hover': { bg: 'rgba(255,255,255,0.3)' },
            }}
          >
            ×
          </Box>
        </Box>
      )}
    </Box>
  )
}
