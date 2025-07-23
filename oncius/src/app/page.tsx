"use client"

import Link from 'next/link'
import { Box, Heading, Text, Button, Flex, Container } from 'theme-ui'

export default function HomePage() {
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
