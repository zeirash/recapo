"use client"

import { useState } from 'react'
import Link from 'next/link'
import { Box, Container, Heading, Text, Input, Button, Alert, Flex } from 'theme-ui'
import { useTranslations } from 'next-intl'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'

const RegisterPage = () => {
  const t = useTranslations()
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [errors, setErrors] = useState<{ [key: string]: string }>({})

  const { register, registerLoading, registerError } = useAuth()

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!formData.name.trim()) {
      newErrors.name = t('validation.nameRequired')
    }

    if (!formData.email) {
      newErrors.email = t('validation.emailRequired')
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = t('validation.emailInvalid')
    }

    if (!formData.password) {
      newErrors.password = t('validation.passwordRequired')
    } else if (formData.password.length < 6) {
      newErrors.password = t('validation.passwordMinLength')
    }

    if (!formData.confirmPassword) {
      newErrors.confirmPassword = t('validation.confirmPasswordRequired')
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = t('validation.passwordsMismatch')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))

    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    register({
      name: formData.name,
      email: formData.email,
      password: formData.password,
    })
  }

  return (
    <Layout>
      <Container sx={{ maxWidth: '500px' }}>
        <Box sx={{ bg: 'background', p: 4, borderRadius: 'large', boxShadow: 'medium' }}>
          <Heading as="h1" sx={{ textAlign: 'center', mb: 4 }}>
            {t('auth.register')}
          </Heading>

          {registerError && (
            <Alert sx={{ mb: 3, bg: 'error', color: 'white' }}>
              {registerError instanceof Error ? registerError.message : t('auth.registrationFailed')}
            </Alert>
          )}

          <Box as="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: 3 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                {t('auth.fullName')}
              </Text>
              <Input
                name="name"
                type="text"
                value={formData.name}
                onChange={handleChange}
                placeholder={t('auth.enterFullName')}
                sx={{ width: '100%' }}
                className={errors.name ? 'error' : ''}
              />
              {errors.name && (
                <Text sx={{ color: 'error', fontSize: 0, mt: 1 }}>
                  {errors.name}
                </Text>
              )}
            </Box>

            <Box sx={{ mb: 3 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                {t('common.email')}
              </Text>
              <Input
                name="email"
                type="email"
                value={formData.email}
                onChange={handleChange}
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

            <Box sx={{ mb: 3 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                {t('common.password')}
              </Text>
              <Input
                name="password"
                type="password"
                value={formData.password}
                onChange={handleChange}
                placeholder={t('auth.enterPassword')}
                sx={{ width: '100%' }}
                className={errors.password ? 'error' : ''}
              />
              {errors.password && (
                <Text sx={{ color: 'error', fontSize: 0, mt: 1 }}>
                  {errors.password}
                </Text>
              )}
            </Box>

            <Box sx={{ mb: 4 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                {t('auth.confirmPassword')}
              </Text>
              <Input
                name="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={handleChange}
                placeholder={t('auth.confirmYourPassword')}
                sx={{ width: '100%' }}
                className={errors.confirmPassword ? 'error' : ''}
              />
              {errors.confirmPassword && (
                <Text sx={{ color: 'error', fontSize: 0, mt: 1 }}>
                  {errors.confirmPassword}
                </Text>
              )}
            </Box>

            <Button
              type="submit"
              variant="primary"
              sx={{ width: '100%', mb: 3 }}
              disabled={registerLoading}
            >
              {registerLoading ? t('auth.creatingAccount') : t('auth.register')}
            </Button>

            <Flex sx={{ justifyContent: 'center', gap: 1 }}>
              <Text sx={{ color: 'text.secondary' }}>{t('auth.hasAccount')}</Text>
              <Link href="/login">
                <Text
                  sx={{ color: 'primary', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } }}
                >
                  {t('auth.loginHere')}
                </Text>
              </Link>
            </Flex>
          </Box>
        </Box>
      </Container>
    </Layout>
  )
}

export default RegisterPage
