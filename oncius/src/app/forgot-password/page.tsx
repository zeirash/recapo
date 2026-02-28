"use client"

import { useState } from 'react'
import Link from 'next/link'
import { Box, Container, Heading, Text, Input, Button, Flex } from 'theme-ui'
import { useTranslations } from 'next-intl'
import Layout from '@/components/Layout'

const ForgotPasswordPage = () => {
  const t = useTranslations()
  const [email, setEmail] = useState('')
  const [errors, setErrors] = useState<{ [key: string]: string }>({})
  const [submitted, setSubmitted] = useState(false)

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!email) {
      newErrors.email = t('validation.emailRequired')
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = t('validation.emailInvalid')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    setSubmitted(true)
  }

  return (
    <Layout>
      <Container sx={{ maxWidth: '500px' }}>
        <Box sx={{ bg: 'background', p: 4, borderRadius: 'large', boxShadow: 'medium' }}>
          <Heading as="h1" sx={{ textAlign: 'center', mb: 2 }}>
            {t('auth.forgotPasswordTitle')}
          </Heading>

          <Text sx={{ textAlign: 'center', color: 'text.secondary', mb: 4, display: 'block' }}>
            {t('auth.forgotPasswordDescription')}
          </Text>

          {submitted ? (
            <Box sx={{ textAlign: 'center' }}>
              <Text sx={{ color: 'success', mb: 4, display: 'block' }}>
                If an account exists for {email}, a reset link will be sent shortly.
              </Text>
              <Link href="/login">
                <Text sx={{ color: 'primary', cursor: 'pointer' }}>
                  {t('auth.backToLogin')}
                </Text>
              </Link>
            </Box>
          ) : (
            <Box as="form" onSubmit={handleSubmit}>
              <Box sx={{ mb: 4 }}>
                <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                  {t('common.email')}
                </Text>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder={t('auth.enterEmail')}
                  sx={{ width: '100%' }}
                  className={errors.email ? 'error' : ''}
                />
                {errors.email && (
                  <Text sx={{ color: 'error', fontSize: 0, mt: 1 }}>
                    {errors.email}
                  </Text>
                )}
              </Box>

              <Button
                type="submit"
                variant="primary"
                sx={{ width: '100%', mb: 3 }}
              >
                {t('auth.forgotPasswordSubmit')}
              </Button>

              <Flex sx={{ justifyContent: 'center' }}>
                <Link href="/login">
                  <Text sx={{ color: 'primary', cursor: 'pointer' }}>
                    {t('auth.backToLogin')}
                  </Text>
                </Link>
              </Flex>
            </Box>
          )}
        </Box>
      </Container>
    </Layout>
  )
}

export default ForgotPasswordPage
