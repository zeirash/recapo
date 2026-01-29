"use client"

import { useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Box, Heading, Text, Button, Flex, Container } from 'theme-ui'
import { useAuth } from '@/hooks/useAuth'

export default function HomePage() {
  const router = useRouter()
  const { isAuthenticated, isLoadingUser } = useAuth()

  useEffect(() => {
    if (!isLoadingUser && isAuthenticated) {
      router.replace('/dashboard')
    }
  }, [isAuthenticated, isLoadingUser, router])

  if (isLoadingUser) {
    return (
      <Container>
        <Box sx={{ py: 6, textAlign: 'center' }}>
          <Text>Loading...</Text>
        </Box>
      </Container>
    )
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <Container>
      <Box sx={{ py: 6, textAlign: 'center' }}>
        <Heading as="h1" sx={{ fontSize: [4, 5, 6], mb: 3 }}>
          Recapo
        </Heading>
        <Text sx={{ fontSize: [2, 3], mb: 4, color: 'text.secondary' }}>
          Order Management System for Jastipers
        </Text>
        <Text sx={{ fontSize: [1, 2], mb: 6, maxWidth: '600px', mx: 'auto' }}>
          Streamline your cross-border social media selling business with our comprehensive
          order management platform. Track products, manage orders, and serve customers efficiently.
        </Text>

        <Flex sx={{ gap: 3, justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link href="/login">
            <Button variant="primary" sx={{ px: 6, py: 3, fontSize: 2 }}>
              Login
            </Button>
          </Link>
          <Link href="/register">
            <Button variant="secondary" sx={{ px: 6, py: 3, fontSize: 2 }}>
              Register
            </Button>
          </Link>
        </Flex>
      </Box>
    </Container>
  )
}
