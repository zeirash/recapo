"use client"

import { useQuery } from 'react-query'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { Box, Card, Flex, Text } from 'theme-ui'
import { useTranslations } from 'next-intl'
import { api } from '@/utils/api'

type Product = {
  id: number
  name: string
  price: number
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

  if (!shareToken) {
    return (
      <Box sx={{ minHeight: '100vh', p: 4 }}>
        <Link href="/">
          <Text sx={{ fontSize: 3, fontWeight: 700, color: 'primary', textDecoration: 'none' }}>
            Recapo
          </Text>
        </Link>
        <Box sx={{ py: 6, textAlign: 'center' }}>
          <Text sx={{ color: 'error' }}>{tShare('invalidLink')}</Text>
        </Box>
      </Box>
    )
  }

  if (isLoading) {
    return (
      <Box sx={{ minHeight: '100vh', p: 4 }}>
        <Link href="/">
          <Text sx={{ fontSize: 3, fontWeight: 700, color: 'primary', textDecoration: 'none' }}>
            Recapo
          </Text>
        </Link>
        <Box sx={{ py: 6, textAlign: 'center' }}>
          <Text sx={{ color: 'text.secondary' }}>{t('loading')}</Text>
        </Box>
      </Box>
    )
  }

  if (isError) {
    const err = error as Error & { status?: number }
    const is404 = err?.message?.toLowerCase().includes('not found') || err?.status === 404
    return (
      <Box sx={{ minHeight: '100vh', p: 4 }}>
        <Link href="/">
          <Text sx={{ fontSize: 3, fontWeight: 700, color: 'primary', textDecoration: 'none' }}>
            Recapo
          </Text>
        </Link>
        <Box sx={{ py: 6, textAlign: 'center' }}>
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
    <Box sx={{ minHeight: '100vh', p: 4 }}>
      <Link href="/">
        <Text sx={{ fontSize: 3, fontWeight: 700, color: 'primary', textDecoration: 'none' }}>
          Recapo
        </Text>
      </Link>

      <Box sx={{ mt: 4 }}>
        {products.length === 0 ? (
          <Text sx={{ color: 'text.secondary' }}>{tShare('noProducts')}</Text>
        ) : (
          <Flex
            sx={{
              display: 'grid',
              gridTemplateColumns: ['1fr', 'repeat(2, 1fr)', 'repeat(3, 1fr)'],
              gap: 3,
            }}
          >
            {products.map((product) => (
              <Card
                key={product.id}
                sx={{
                  p: 4,
                  borderRadius: 'medium',
                  border: '1px solid',
                  borderColor: 'border',
                  bg: 'white',
                  boxShadow: 'small',
                }}
              >
                <Flex sx={{ flexDirection: 'column', justifyContent: 'space-between', minHeight: 60 }}>
                  <Text sx={{ fontWeight: 600, fontSize: 2, mb: 2 }}>{product.name}</Text>
                  <Text sx={{ color: 'primary', fontWeight: 600 }}>
                    Rp. {product.price.toLocaleString()}
                  </Text>
                </Flex>
              </Card>
            ))}
          </Flex>
        )}
      </Box>
    </Box>
  )
}
