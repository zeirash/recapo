"use client"

import { useState } from 'react'
import { useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { Box, Container, Heading, Text, Input, Button, Alert, Flex } from 'theme-ui'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'

const LoginPage = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<{ [key: string]: string }>({})

  const { login, loginLoading, loginError } = useAuth()
  const searchParams = useSearchParams()
  const message = searchParams.get('message')

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (!email) {
      newErrors.email = 'Email is required'
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = 'Email is invalid'
    }

    if (!password) {
      newErrors.password = 'Password is required'
    } else if (password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    login({ email, password })
  }

  return (
    <Layout>
      <Container sx={{ maxWidth: '500px' }}>
        <Box sx={{ bg: 'background', p: 4, borderRadius: 'large', boxShadow: 'medium' }}>
          <Heading as="h1" sx={{ textAlign: 'center', mb: 4 }}>
            Login
          </Heading>

          {message && (
            <Alert sx={{ mb: 3, bg: 'success', color: 'white' }}>
              {message}
            </Alert>
          )}

          {loginError && (
            <Alert sx={{ mb: 3, bg: 'error', color: 'white' }}>
              {loginError instanceof Error ? loginError.message : 'Login failed'}
            </Alert>
          )}

          <Box as="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: 3 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                Email
              </Text>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="Enter your email"
                sx={{ width: '100%' }}
                className={errors.email ? 'error' : ''}
              />
              {errors.email && (
                <Text sx={{ color: 'error', fontSize: 0, mt: 1 }}>
                  {errors.email}
                </Text>
              )}
            </Box>

            <Box sx={{ mb: 4 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                Password
              </Text>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter your password"
                sx={{ width: '100%' }}
                className={errors.password ? 'error' : ''}
              />
              {errors.password && (
                <Text sx={{ color: 'error', fontSize: 0, mt: 1 }}>
                  {errors.password}
                </Text>
              )}
            </Box>

            <Button
              type="submit"
              variant="primary"
              sx={{ width: '100%', mb: 3 }}
              disabled={loginLoading}
            >
              {loginLoading ? 'Logging in...' : 'Login'}
            </Button>

            <Flex sx={{ justifyContent: 'center', gap: 1 }}>
              <Text sx={{ color: 'text.secondary' }}>Don't have an account?</Text>
              <Link href="/register">
                <Text
                  sx={{ color: 'primary', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } }}
                >
                  Register here
                </Text>
              </Link>
            </Flex>
          </Box>
        </Box>
      </Container>
    </Layout>
  )
}

export default LoginPage
