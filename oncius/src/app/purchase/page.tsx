"use client"

import { useQuery } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Card, Flex, Heading, Text } from 'theme-ui'
import Layout from '@/components/Layout'
import { api } from '@/utils/api'

type PurchaseProduct = {
  product_name: string
  price: number
  qty: number
}

export default function PurchasePage() {
  const t = useTranslations('common')
  const tPurchase = useTranslations('purchase')
  const tErrors = useTranslations('errors')

  const { data: products, isLoading, isError, error } = useQuery(
    ['purchase_list'],
    async () => {
      const res = await api.getPurchaseListProducts()
      if (!res.success) throw new Error(res.message || tPurchase('fetchFailed'))
      return res.data as PurchaseProduct[]
    }
  )

  return (
    <Layout>
      <Box sx={{ maxWidth: 1200, mx: 'auto', p: [4, 5] }}>
        <Heading as="h1" sx={{ mb: 2, fontSize: 3 }}>
          {tPurchase('title')}
        </Heading>
        <Text sx={{ color: 'text.secondary', fontSize: 1, mb: 4, display: 'block' }}>
          {tPurchase('note')}
        </Text>

        {isLoading ? (
          <Text sx={{ color: 'text.secondary' }}>{t('loading')}</Text>
        ) : isError ? (
          <Text sx={{ color: 'error' }}>
            {(error as Error)?.message || tErrors('loadingError', { resource: tPurchase('title') })}
          </Text>
        ) : (
          <Card
            sx={{
              borderRadius: 'large',
              boxShadow: 'small',
              border: '1px solid',
              borderColor: 'border',
              bg: 'white',
              overflow: 'hidden',
            }}
          >
            {products && products.length > 0 ? (
              <Box as="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                <Box as="thead">
                  <Box as="tr" sx={{ bg: 'background.secondary' }}>
                    <Box
                      as="th"
                      sx={{
                        p: 3,
                        textAlign: 'left',
                        fontSize: 0,
                        fontWeight: 600,
                        color: 'text.secondary',
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      {t('product')}
                    </Box>
                    <Box
                      as="th"
                      sx={{
                        p: 3,
                        textAlign: 'right',
                        fontSize: 0,
                        fontWeight: 600,
                        color: 'text.secondary',
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      {t('quantity')}
                    </Box>
                  </Box>
                </Box>
                <Box as="tbody">
                  {products.map((product, index) => (
                    <Box
                      as="tr"
                      key={index}
                      sx={{
                        borderTop: '1px solid',
                        borderColor: 'border',
                        '&:hover': { bg: 'background.secondary' },
                      }}
                    >
                      <Box as="td" sx={{ py: 2, px: 3, fontSize: 1 }}>
                        <Text sx={{ fontWeight: 600 }}>{product.product_name}</Text>
                      </Box>
                      <Box as="td" sx={{ py: 2, px: 3, textAlign: 'right', fontSize: 1 }}>
                        <Text sx={{ fontWeight: 600, color: 'primary' }}>{product.qty}</Text>
                      </Box>
                    </Box>
                  ))}
                </Box>
              </Box>
            ) : (
              <Box sx={{ p: 5, textAlign: 'center', color: 'text.secondary' }}>
                <Text sx={{ fontSize: 2 }}>{tPurchase('empty')}</Text>
              </Box>
            )}
          </Card>
        )}
      </Box>
    </Layout>
  )
}
