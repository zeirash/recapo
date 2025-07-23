"use client"

import { Box, Flex, Text, Button } from 'theme-ui'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'

const DashboardPage = () => {
  const { user, isAuthenticated } = useAuth()

  if (!isAuthenticated) {
    return (
      <Layout>
        <Box sx={{ p: 4, textAlign: 'center' }}>
          <Text>Please log in to access the dashboard.</Text>
        </Box>
      </Layout>
    )
  }

  return (
    <Layout>
      this will be dashboard page (statistic?)
    </Layout>
  )
}

export default DashboardPage
