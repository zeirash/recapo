"use client"

import { useQuery } from 'react-query'
import { useTranslations } from 'next-intl'
import { Box, Paper, Typography } from '@mui/material'
import Layout from '@/components/layout'
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
      <Box sx={{ maxWidth: 1200, mx: 'auto', p: { xs: '24px', sm: '32px' } }}>
        <Typography component="h1" sx={{ mb: '8px', fontSize: '18px' }}>
          {tPurchase('title')}
        </Typography>
        <Box sx={{ color: 'grey.500', fontSize: '14px', mb: '24px', display: 'block' }}>
          {tPurchase('note')}
        </Box>

        {isLoading ? (
          <Box sx={{ color: 'grey.500' }}>{t('loading')}</Box>
        ) : isError ? (
          <Box sx={{ color: 'error.main' }}>
            {(error as Error)?.message || tErrors('loadingError', { resource: tPurchase('title') })}
          </Box>
        ) : (
          <Paper
            sx={{
              borderRadius: '12px',
              boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
              border: '1px solid',
              borderColor: 'grey.200',
              bgcolor: 'white',
              overflow: 'hidden',
            }}
          >
            {products && products.length > 0 ? (
              <Box component="table" sx={{ width: '100%', borderCollapse: 'collapse' }}>
                <Box component="thead">
                  <Box component="tr" sx={{ bgcolor: 'grey.50' }}>
                    <Box
                      component="th"
                      sx={{
                        p: '16px',
                        textAlign: 'left',
                        fontSize: '12px',
                        fontWeight: 600,
                        color: 'grey.500',
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      {t('product')}
                    </Box>
                    <Box
                      component="th"
                      sx={{
                        p: '16px',
                        textAlign: 'right',
                        fontSize: '12px',
                        fontWeight: 600,
                        color: 'grey.500',
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      {t('quantity')}
                    </Box>
                  </Box>
                </Box>
                <Box component="tbody">
                  {products.map((product, index) => (
                    <Box
                      component="tr"
                      key={index}
                      sx={{
                        borderTop: '1px solid',
                        borderColor: 'grey.200',
                        '&:hover': { bgcolor: 'grey.50' },
                      }}
                    >
                      <Box component="td" sx={{ py: '8px', px: '16px', fontSize: '14px' }}>
                        <Box sx={{ fontWeight: 600 }}>{product.product_name}</Box>
                      </Box>
                      <Box component="td" sx={{ py: '8px', px: '16px', textAlign: 'right', fontSize: '14px' }}>
                        <Box sx={{ fontWeight: 600, color: 'primary.main' }}>{product.qty}</Box>
                      </Box>
                    </Box>
                  ))}
                </Box>
              </Box>
            ) : (
              <Box sx={{ p: '32px', textAlign: 'center', color: 'grey.500' }}>
                <Box sx={{ fontSize: '16px' }}>{tPurchase('empty')}</Box>
              </Box>
            )}
          </Paper>
        )}
      </Box>
    </Layout>
  )
}
