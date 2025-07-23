"use client"

import { useState } from 'react'
import Link from 'next/link'
import { Box, Container, Heading, Text, Input, Button, Alert, Flex } from 'theme-ui'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'

const RegisterPage = () => {
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
      newErrors.name = 'Name is required'
    }

    if (!formData.email) {
      newErrors.email = 'Email is required'
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid'
    }

    if (!formData.password) {
      newErrors.password = 'Password is required'
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters'
    }

    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password'
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match'
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
            Create Account
          </Heading>

          {registerError && (
            <Alert sx={{ mb: 3, bg: 'error', color: 'white' }}>
              {registerError instanceof Error ? registerError.message : 'Registration failed'}
            </Alert>
          )}

          <Box as="form" onSubmit={handleSubmit}>
            <Box sx={{ mb: 3 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                Full Name
              </Text>
              <Input
                name="name"
                type="text"
                value={formData.name}
                onChange={handleChange}
                placeholder="Enter your full name"
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
                Email
              </Text>
              <Input
                name="email"
                type="email"
                value={formData.email}
                onChange={handleChange}
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

            <Box sx={{ mb: 3 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                Password
              </Text>
              <Input
                name="password"
                type="password"
                value={formData.password}
                onChange={handleChange}
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

            <Box sx={{ mb: 4 }}>
              <Text as="label" sx={{ display: 'block', mb: 1, fontWeight: 'heading' }}>
                Confirm Password
              </Text>
              <Input
                name="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={handleChange}
                placeholder="Confirm your password"
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
              {registerLoading ? 'Creating account...' : 'Create Account'}
            </Button>

            <Flex sx={{ justifyContent: 'center', gap: 1 }}>
              <Text sx={{ color: 'text.secondary' }}>Already have an account?</Text>
              <Link href="/login">
                <Text
                  sx={{ color: 'primary', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } }}
                >
                  Login here
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
